// Package zenodotransfer implements authenticated, sandbox-only Zenodo draft
// file transfers. It deliberately excludes publication and production writes.
package zenodotransfer

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/auth"
	"github.com/edithatogo/osf-cli-go/internal/download"
	"github.com/edithatogo/osf-cli-go/internal/observability"
)

const (
	defaultSandboxBaseURL = "https://sandbox.zenodo.org/api/"
	defaultTimeout        = 2 * time.Minute
	defaultResponseLimit  = int64(8 << 20)
	defaultFileLimit      = int64(50 << 30)
	defaultFileCountLimit = 100
	defaultMaxRetries     = 2
	defaultRetryDelay     = 200 * time.Millisecond
)

var (
	// ErrProductionWrite indicates that a transfer tried to target production Zenodo.
	ErrProductionWrite = errors.New("zenodo transfers are restricted to the sandbox")
	// ErrResponseTooLarge indicates that a control response exceeded its memory budget.
	ErrResponseTooLarge = errors.New("zenodo transfer response exceeds configured size limit")
	// ErrFileTooLarge indicates that a source exceeds the configured transfer budget.
	ErrFileTooLarge = errors.New("zenodo transfer file exceeds configured size limit")
	// ErrFileCountLimit indicates that a draft cannot accept another file.
	ErrFileCountLimit = errors.New("zenodo draft reached configured file count limit")
	// ErrRemoteConflict indicates that the requested draft filename already exists.
	ErrRemoteConflict = errors.New("zenodo draft file already exists")
	// ErrCrossOrigin indicates an API or file link escaped the configured sandbox origin.
	ErrCrossOrigin = errors.New("zenodo transfer link leaves configured sandbox API origin")
	// ErrInvalidContentRange indicates a resumed response started at the wrong byte.
	ErrInvalidContentRange = errors.New("zenodo download returned an invalid content range")
)

// Draft identifies a disposable Zenodo sandbox deposition and its upload bucket.
type Draft struct {
	ID        string `json:"id"`
	BucketURL string `json:"bucketUrl"`
}

// RemoteFile describes a draft file returned by the Zenodo depositions API.
type RemoteFile struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	Checksum    string `json:"checksum"`
	DownloadURL string `json:"downloadUrl"`
}

// UploadResult reports a whole-file sandbox upload. Resumed remains false
// because the documented Zenodo bucket PUT has no partial-upload protocol.
type UploadResult struct {
	Remote         RemoteFile `json:"remote"`
	Bytes          int64      `json:"bytes"`
	RetryCount     int        `json:"retryCount"`
	Resumed        bool       `json:"resumed"`
	Skipped        bool       `json:"skipped"`
	Completed      bool       `json:"completed"`
	CheckpointPath string     `json:"checkpointPath,omitempty"`
}

// APIError preserves safe sandbox HTTP failure details without response headers
// or credentials.
type APIError struct {
	StatusCode int
	Method     string
	Path       string
	Message    string
}

func (err *APIError) Error() string {
	if err == nil {
		return "<nil>"
	}
	message := fmt.Sprintf("Zenodo sandbox %s %s returned %d", err.Method, err.Path, err.StatusCode)
	if err.Message != "" {
		message += ": " + err.Message
	}
	return message
}

// Client performs authenticated writes only against a configured Zenodo sandbox.
type Client struct {
	baseURL          *url.URL
	httpClient       *http.Client
	token            string
	emitter          observability.Emitter
	maxResponseBytes int64
	maxFileBytes     int64
	maxFiles         int
	maxRetries       int
	retryDelay       time.Duration
}

// Option configures a transfer Client.
type Option func(*Client)

// WithHTTPClient replaces the bounded default HTTP client.
func WithHTTPClient(client *http.Client) Option { return func(c *Client) { c.httpClient = client } }

// WithObserver enables redacted provider-tagged transfer events.
func WithObserver(emitter observability.Emitter) Option {
	return func(c *Client) { c.emitter = emitter }
}

// WithMaxResponseBytes sets the control-response memory budget.
func WithMaxResponseBytes(limit int64) Option { return func(c *Client) { c.maxResponseBytes = limit } }

// WithMaxFileBytes sets the local upload size budget.
func WithMaxFileBytes(limit int64) Option { return func(c *Client) { c.maxFileBytes = limit } }

// WithMaxFiles sets the per-draft file-count budget.
func WithMaxFiles(limit int) Option { return func(c *Client) { c.maxFiles = limit } }

// WithRetryPolicy configures retries for safe GET, PUT, and DELETE operations.
func WithRetryPolicy(maxRetries int, delay time.Duration) Option {
	return func(c *Client) { c.maxRetries, c.retryDelay = maxRetries, delay }
}

// New constructs a sandbox-only transfer client. The token must come from the
// dedicated ZENODO_TOKEN credential boundary; this package never reads OSF_TOKEN.
func New(baseURL, token string, options ...Option) (*Client, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("ZENODO_TOKEN is required for sandbox transfers")
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultSandboxBaseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse Zenodo sandbox base URL: %w", err)
	}
	if (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, errors.New("ZENODO_BASE_URL must be a plain HTTP(S) URL without credentials, query, or fragment")
	}
	host := strings.ToLower(parsed.Hostname())
	local := host == "127.0.0.1" || host == "localhost" || host == "::1"
	if !local && (parsed.Scheme != "https" || host != "sandbox.zenodo.org") {
		return nil, ErrProductionWrite
	}
	if !strings.HasSuffix(parsed.Path, "/") {
		parsed.Path += "/"
	}
	client := &Client{
		baseURL: parsed, token: token, httpClient: &http.Client{Timeout: defaultTimeout},
		maxResponseBytes: defaultResponseLimit, maxFileBytes: defaultFileLimit,
		maxFiles:   defaultFileCountLimit,
		maxRetries: defaultMaxRetries, retryDelay: defaultRetryDelay,
	}
	for _, option := range options {
		option(client)
	}
	if client.httpClient == nil {
		client.httpClient = &http.Client{Timeout: defaultTimeout}
	}
	if client.maxResponseBytes <= 0 || client.maxFileBytes <= 0 || client.maxFiles <= 0 || client.maxRetries < 0 || client.retryDelay < 0 {
		return nil, errors.New("Zenodo transfer response, file, and retry budgets must be positive and retries non-negative")
	}
	configuredHTTP := *client.httpClient
	previousRedirect := configuredHTTP.CheckRedirect
	configuredHTTP.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		if !client.sameAPI(request.URL) {
			return ErrCrossOrigin
		}
		if previousRedirect != nil {
			return previousRedirect(request, via)
		}
		if len(via) >= 10 {
			return errors.New("stopped after 10 Zenodo sandbox redirects")
		}
		return nil
	}
	client.httpClient = &configuredHTTP
	return client, nil
}

// CreateDraft creates an empty disposable sandbox deposition. Creation is not
// retried automatically because a lost response could otherwise create orphans.
func (client *Client) CreateDraft(ctx context.Context) (Draft, error) {
	endpoint, err := client.resolve("deposit/depositions")
	if err != nil {
		return Draft{}, err
	}
	body, _, err := client.do(ctx, http.MethodPost, endpoint, func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader([]byte("{}"))), nil
	}, 2, false)
	if err != nil {
		return Draft{}, err
	}
	var payload struct {
		ID    json.RawMessage `json:"id"`
		Links struct {
			Bucket string `json:"bucket"`
		} `json:"links"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return Draft{}, fmt.Errorf("decode Zenodo draft: %w", err)
	}
	id := rawID(payload.ID)
	if id == "" {
		return Draft{}, errors.New("decode Zenodo draft: id or bucket link is missing")
	}
	cleanupFailure := func(cause error) (Draft, error) {
		if cleanupErr := client.DeleteDraft(context.WithoutCancel(ctx), id); cleanupErr != nil {
			return Draft{}, errors.Join(cause, fmt.Errorf("cleanup malformed Zenodo draft %s: %w", id, cleanupErr))
		}
		return Draft{}, cause
	}
	if strings.TrimSpace(payload.Links.Bucket) == "" {
		return cleanupFailure(errors.New("decode Zenodo draft: id or bucket link is missing"))
	}
	if _, err := client.approveLink(payload.Links.Bucket); err != nil {
		return cleanupFailure(fmt.Errorf("validate Zenodo draft bucket: %w", err))
	}
	return Draft{ID: id, BucketURL: payload.Links.Bucket}, nil
}

// DeleteDraft removes an unpublished disposable sandbox deposition.
func (client *Client) DeleteDraft(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("Zenodo draft id is required for cleanup")
	}
	endpoint, err := client.resolve("deposit/depositions/" + url.PathEscape(id))
	if err != nil {
		return err
	}
	_, _, err = client.do(ctx, http.MethodDelete, endpoint, nil, 0, true)
	if isStatus(err, http.StatusNotFound) {
		return nil
	}
	return err
}

// ListDraftFiles returns the current draft inventory used for conflict checks.
func (client *Client) ListDraftFiles(ctx context.Context, id string) ([]RemoteFile, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("Zenodo draft id is required")
	}
	endpoint, err := client.resolve("deposit/depositions/" + url.PathEscape(id) + "/files")
	if err != nil {
		return nil, err
	}
	body, _, err := client.do(ctx, http.MethodGet, endpoint, nil, 0, true)
	if err != nil {
		return nil, err
	}
	var files []filePayload
	if err := json.Unmarshal(body, &files); err != nil {
		return nil, fmt.Errorf("decode Zenodo draft files: %w", err)
	}
	result := make([]RemoteFile, 0, len(files))
	for _, file := range files {
		remote, err := client.remoteFile(file)
		if err != nil {
			return nil, err
		}
		result = append(result, remote)
	}
	return result, nil
}

// UploadFile sends one regular file with explicit remote conflict handling.
// Zenodo's documented bucket PUT is whole-file, so transient retries rewind to
// byte zero and no partial remote-resume claim is made.
func (client *Client) UploadFile(ctx context.Context, draft Draft, sourcePath, remoteName string, policy download.ConflictPolicy) (UploadResult, error) {
	var result UploadResult
	if err := policy.Validate(); err != nil {
		return result, err
	}
	if strings.TrimSpace(draft.ID) == "" || strings.TrimSpace(draft.BucketURL) == "" {
		return result, errors.New("Zenodo draft id and bucket URL are required")
	}
	remoteName = strings.TrimSpace(remoteName)
	if remoteName == "" || remoteName != filepath.Base(remoteName) || strings.ContainsAny(remoteName, "/\\") || remoteName == "." {
		return result, errors.New("Zenodo remote filename must be a single non-empty path segment")
	}
	info, err := os.Stat(sourcePath)
	if err != nil {
		return result, fmt.Errorf("stat Zenodo upload source: %w", err)
	}
	if !info.Mode().IsRegular() {
		return result, errors.New("Zenodo upload source must be a regular file")
	}
	if info.Size() > client.maxFileBytes {
		return result, fmt.Errorf("%w: %d > %d bytes", ErrFileTooLarge, info.Size(), client.maxFileBytes)
	}
	existing, err := client.ListDraftFiles(ctx, draft.ID)
	if err != nil {
		return result, fmt.Errorf("check Zenodo upload conflict: %w", err)
	}
	replacing := false
	for _, remote := range existing {
		if remote.Name != remoteName {
			continue
		}
		replacing = true
		switch policy {
		case download.ConflictFail:
			return result, fmt.Errorf("%w: %s", ErrRemoteConflict, remoteName)
		case download.ConflictSkip:
			result.Remote = remote
			result.Skipped = true
			result.Completed = true
			return result, nil
		}
	}
	if len(existing) >= client.maxFiles && !replacing {
		return result, fmt.Errorf("%w: %d >= %d", ErrFileCountLimit, len(existing), client.maxFiles)
	}

	bucket, err := client.approveLink(draft.BucketURL)
	if err != nil {
		return result, fmt.Errorf("validate Zenodo upload bucket: %w", err)
	}
	bucket.Path = strings.TrimSuffix(bucket.Path, "/") + "/" + url.PathEscape(remoteName)
	expectedChecksum, err := md5File(sourcePath)
	if err != nil {
		return result, err
	}
	checkpoint := sourcePath + ".zenodo-upload.resume.json"
	var uploaded RemoteFile
	retries := 0
	transfer, err := download.ResumeFileUpload(ctx, download.UploadOptions{
		SourcePath: sourcePath, SourceIdentity: "zenodo:draft:" + draft.ID + ":" + remoteName,
		CheckpointPath: checkpoint, Emitter: client.emitter,
	}, func(ctx context.Context, offset, total int64, reader io.Reader) (int64, bool, error) {
		if offset != 0 {
			return offset, false, errors.New("Zenodo bucket API does not support partial upload resume")
		}
		seeker, ok := reader.(io.ReadSeeker)
		if !ok {
			return 0, false, errors.New("Zenodo upload source is not seekable")
		}
		payload, retryCount, err := client.putFile(ctx, bucket, seeker, total)
		retries += retryCount
		if err != nil {
			return 0, false, err
		}
		remote, err := client.remoteFile(payload)
		if err != nil {
			return 0, false, err
		}
		if remote.Size != total {
			return 0, false, fmt.Errorf("Zenodo upload size mismatch: got %d, want %d", remote.Size, total)
		}
		if normalizeChecksum(remote.Checksum) != expectedChecksum {
			return 0, false, fmt.Errorf("Zenodo upload checksum mismatch: got %s, want %s", normalizeChecksum(remote.Checksum), expectedChecksum)
		}
		uploaded = remote
		return total, true, nil
	})
	result.Bytes = transfer.Bytes
	result.Resumed = transfer.Resumed
	result.Completed = transfer.Completed
	result.CheckpointPath = transfer.CheckpointPath
	result.RetryCount = retries
	result.Remote = uploaded
	if err != nil {
		return result, err
	}
	return result, nil
}

// DownloadFile atomically downloads and verifies a sandbox file, retaining a
// non-secret checkpoint after interruption or cancellation.
func (client *Client) DownloadFile(ctx context.Context, remote RemoteFile, destination string, policy download.ConflictPolicy) (download.ResumeResult, error) {
	if remote.Size < 0 || strings.TrimSpace(remote.DownloadURL) == "" {
		return download.ResumeResult{}, errors.New("Zenodo remote file size and download URL are required")
	}
	checksum := normalizeChecksum(remote.Checksum)
	if !validMD5(checksum) {
		return download.ResumeResult{}, errors.New("Zenodo remote file requires a valid MD5 checksum")
	}
	endpoint, err := client.approveLink(remote.DownloadURL)
	if err != nil {
		return download.ResumeResult{}, fmt.Errorf("validate Zenodo download link: %w", err)
	}
	return download.ResumeStreamAtomically(func(offset int64) (io.ReadCloser, error) {
		return client.openDownload(ctx, endpoint, offset)
	}, download.ResumeOptions{
		Destination: destination, Source: endpoint.String(), ExpectedSize: &remote.Size,
		ExpectedChecksum: checksum, Policy: policy, Context: ctx, Emitter: client.emitter,
	})
}

// ValidateResumableDownload deterministically interrupts a sandbox download at
// the requested byte count and then requires the normal checkpointed transfer
// to continue with a provider range response. It is intended for the opt-in
// disposable sandbox validation harness.
func (client *Client) ValidateResumableDownload(ctx context.Context, remote RemoteFile, destination string, interruptAfter int64) (download.ResumeResult, error) {
	if interruptAfter <= 0 || interruptAfter >= remote.Size {
		return download.ResumeResult{}, errors.New("Zenodo resume validation offset must be within the remote file")
	}
	checksum := normalizeChecksum(remote.Checksum)
	if !validMD5(checksum) {
		return download.ResumeResult{}, errors.New("Zenodo remote file requires a valid MD5 checksum")
	}
	endpoint, err := client.approveLink(remote.DownloadURL)
	if err != nil {
		return download.ResumeResult{}, fmt.Errorf("validate Zenodo download link: %w", err)
	}
	first, firstErr := download.ResumeStreamAtomically(func(offset int64) (io.ReadCloser, error) {
		body, err := client.openDownload(ctx, endpoint, offset)
		if err != nil {
			return nil, err
		}
		return &limitedReadCloser{Reader: io.LimitReader(body, interruptAfter), Closer: body}, nil
	}, download.ResumeOptions{
		Destination: destination, Source: endpoint.String(), ExpectedSize: &remote.Size,
		ExpectedChecksum: checksum, Policy: download.ConflictOverwrite, Context: ctx, Emitter: client.emitter,
	})
	partialInfo, statErr := os.Stat(destination + ".part")
	if firstErr == nil || first.Completed || statErr != nil || partialInfo.Size() != interruptAfter {
		return download.ResumeResult{}, fmt.Errorf("Zenodo resume validation did not stop at %d bytes: result=%+v error=%v", interruptAfter, first, firstErr)
	}
	result, err := client.DownloadFile(ctx, remote, destination, download.ConflictOverwrite)
	if err != nil {
		return result, fmt.Errorf("resume Zenodo validation download: %w", err)
	}
	if !result.Resumed {
		return result, errors.New("Zenodo validation download restarted instead of resuming")
	}
	return result, nil
}

type limitedReadCloser struct {
	io.Reader
	io.Closer
}

type filePayload struct {
	ID       json.RawMessage `json:"id"`
	Filename string          `json:"filename"`
	Key      string          `json:"key"`
	Filesize int64           `json:"filesize"`
	Size     int64           `json:"size"`
	Checksum string          `json:"checksum"`
	Links    struct {
		Download string `json:"download"`
		Content  string `json:"content"`
	} `json:"links"`
}

func (client *Client) remoteFile(payload filePayload) (RemoteFile, error) {
	name := strings.TrimSpace(payload.Filename)
	if name == "" {
		name = strings.TrimSpace(payload.Key)
	}
	size := payload.Filesize
	if size == 0 && payload.Size > 0 {
		size = payload.Size
	}
	downloadURL := strings.TrimSpace(payload.Links.Download)
	if downloadURL == "" {
		downloadURL = strings.TrimSpace(payload.Links.Content)
	}
	remote := RemoteFile{ID: rawID(payload.ID), Name: name, Size: size, Checksum: normalizeChecksum(payload.Checksum), DownloadURL: downloadURL}
	if remote.ID == "" || remote.Name == "" || remote.Size < 0 || remote.Checksum == "" || remote.DownloadURL == "" {
		return RemoteFile{}, errors.New("decode Zenodo file: id, name, size, checksum, or download link is missing")
	}
	if _, err := client.approveLink(remote.DownloadURL); err != nil {
		return RemoteFile{}, fmt.Errorf("validate Zenodo file download link: %w", err)
	}
	return remote, nil
}

func (client *Client) putFile(ctx context.Context, endpoint *url.URL, source io.ReadSeeker, size int64) (filePayload, int, error) {
	body, retries, err := client.do(ctx, http.MethodPut, endpoint, func() (io.ReadCloser, error) {
		if _, err := source.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		return io.NopCloser(io.LimitReader(source, size)), nil
	}, size, true)
	if err != nil {
		return filePayload{}, retries, err
	}
	var payload filePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return filePayload{}, retries, fmt.Errorf("decode Zenodo uploaded file: %w", err)
	}
	return payload, retries, nil
}

func (client *Client) openDownload(ctx context.Context, endpoint *url.URL, offset int64) (io.ReadCloser, error) {
	for attempt := 0; ; attempt++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
		if err != nil {
			return nil, err
		}
		request.Header.Set("Accept", "application/octet-stream")
		request.Header.Set("Authorization", "Bearer "+client.token)
		if offset > 0 {
			request.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
		}
		response, err := client.httpClient.Do(request)
		if err != nil {
			if attempt < client.maxRetries && ctx.Err() == nil {
				if err := sleepContext(ctx, client.retryDelay); err != nil {
					return nil, err
				}
				continue
			}
			return nil, err
		}
		if retryableStatus(response.StatusCode) && attempt < client.maxRetries {
			_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4<<10))
			_ = response.Body.Close()
			if err := sleepContext(ctx, retryDelay(response.Header.Get("Retry-After"), client.retryDelay)); err != nil {
				return nil, err
			}
			continue
		}
		if offset > 0 && response.StatusCode == http.StatusOK {
			_ = response.Body.Close()
			return nil, download.ErrRangeUnsupported
		}
		if offset > 0 && response.StatusCode == http.StatusPartialContent && !validContentRange(response.Header.Get("Content-Range"), offset) {
			_ = response.Body.Close()
			return nil, ErrInvalidContentRange
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			body, _ := readBounded(response.Body, client.maxResponseBytes)
			_ = response.Body.Close()
			return nil, client.responseError(request, response, body)
		}
		return response.Body, nil
	}
}

func (client *Client) do(ctx context.Context, method string, endpoint *url.URL, bodyFactory func() (io.ReadCloser, error), contentLength int64, retryable bool) ([]byte, int, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	retries := 0
	for attempt := 0; ; attempt++ {
		var body io.ReadCloser
		var err error
		if bodyFactory != nil {
			body, err = bodyFactory()
			if err != nil {
				return nil, retries, fmt.Errorf("prepare Zenodo transfer request: %w", err)
			}
		}
		request, err := http.NewRequestWithContext(ctx, method, endpoint.String(), body)
		if err != nil {
			if body != nil {
				_ = body.Close()
			}
			return nil, retries, err
		}
		request.Header.Set("Accept", "application/json")
		request.Header.Set("Authorization", "Bearer "+client.token)
		if body != nil {
			contentType := "application/octet-stream"
			if method == http.MethodPost {
				contentType = "application/json"
			}
			request.Header.Set("Content-Type", contentType)
			request.ContentLength = contentLength
		}
		response, err := client.httpClient.Do(request)
		if err != nil {
			if retryable && attempt < client.maxRetries && ctx.Err() == nil {
				retries++
				if err := sleepContext(ctx, client.retryDelay); err != nil {
					return nil, retries, err
				}
				continue
			}
			return nil, retries, fmt.Errorf("Zenodo sandbox request: %w", err)
		}
		if response.StatusCode >= 300 && response.StatusCode < 400 {
			location, locationErr := response.Location()
			if locationErr == nil && !client.sameAPI(location) {
				_ = response.Body.Close()
				return nil, retries, ErrCrossOrigin
			}
		}
		if retryable && retryableStatus(response.StatusCode) && attempt < client.maxRetries {
			_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4<<10))
			_ = response.Body.Close()
			retries++
			if err := sleepContext(ctx, retryDelay(response.Header.Get("Retry-After"), client.retryDelay)); err != nil {
				return nil, retries, err
			}
			continue
		}
		responseBody, readErr := readBounded(response.Body, client.maxResponseBytes)
		closeErr := response.Body.Close()
		if readErr != nil {
			return nil, retries, readErr
		}
		if closeErr != nil {
			return nil, retries, closeErr
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return nil, retries, client.responseError(request, response, responseBody)
		}
		return responseBody, retries, nil
	}
}

func (client *Client) responseError(request *http.Request, response *http.Response, body []byte) error {
	var payload struct {
		Message string `json:"message"`
	}
	_ = json.Unmarshal(body, &payload)
	message := strings.TrimSpace(payload.Message)
	if message == "" {
		message = strings.TrimSpace(string(body))
	}
	message = auth.Redact(message, client.token)
	if len(message) > 1024 {
		message = message[:1024] + "..."
	}
	return &APIError{StatusCode: response.StatusCode, Method: request.Method, Path: request.URL.Path, Message: message}
}

func (client *Client) resolve(reference string) (*url.URL, error) {
	resolved := client.baseURL.ResolveReference(&url.URL{Path: reference})
	if !client.sameAPI(resolved) {
		return nil, ErrCrossOrigin
	}
	return resolved, nil
}

func (client *Client) approveLink(raw string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, err
	}
	if !client.sameAPI(parsed) || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, ErrCrossOrigin
	}
	return parsed, nil
}

func (client *Client) sameAPI(candidate *url.URL) bool {
	return candidate != nil && candidate.Scheme == client.baseURL.Scheme && strings.EqualFold(candidate.Host, client.baseURL.Host) && strings.HasPrefix(candidate.Path, client.baseURL.Path)
}

func rawID(value json.RawMessage) string {
	trimmed := strings.TrimSpace(string(value))
	if len(trimmed) >= 2 && trimmed[0] == '"' {
		var text string
		if json.Unmarshal(value, &text) == nil {
			return strings.TrimSpace(text)
		}
	}
	if _, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return trimmed
	}
	return ""
}

func md5File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open Zenodo upload source for checksum: %w", err)
	}
	defer func() { _ = file.Close() }()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("checksum Zenodo upload source: %w", err)
	}
	return "md5:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func normalizeChecksum(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) == 32 && !strings.Contains(value, ":") {
		return "md5:" + value
	}
	return value
}

func validMD5(value string) bool {
	if !strings.HasPrefix(value, "md5:") || len(value) != len("md5:")+32 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "md5:"))
	return err == nil
}

func validContentRange(value string, offset int64) bool {
	prefix := fmt.Sprintf("bytes %d-", offset)
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(value)), prefix)
}

func readBounded(reader io.Reader, limit int64) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, ErrResponseTooLarge
	}
	return body, nil
}

func retryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || status == http.StatusBadGateway || status == http.StatusServiceUnavailable || status == http.StatusGatewayTimeout
}

func isStatus(err error, status int) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == status
}

func retryDelay(value string, fallback time.Duration) time.Duration {
	seconds, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || seconds < 0 {
		return fallback
	}
	delay := time.Duration(seconds) * time.Second
	if delay > 5*time.Second {
		return 5 * time.Second
	}
	return delay
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
