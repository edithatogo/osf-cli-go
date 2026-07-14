package zenodopublish

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/auth"
)

const (
	defaultSandboxBaseURL = "https://sandbox.zenodo.org/api/"
	defaultTimeout        = 30 * time.Second
	defaultResponseLimit  = int64(8 << 20)
)

var (
	// ErrProductionWrite indicates that a lifecycle action targeted production Zenodo.
	ErrProductionWrite = errors.New("zenodo publication actions are restricted to the sandbox")
	// ErrCrossOrigin indicates that a provider link escaped the configured sandbox API.
	ErrCrossOrigin = errors.New("zenodo publication link leaves configured sandbox API origin")
	// ErrResponseTooLarge indicates that a control response exceeded its memory budget.
	ErrResponseTooLarge = errors.New("zenodo publication response exceeds configured size limit")
)

// APIError preserves safe HTTP failure details without headers or credentials.
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

// Result reports a dry-run or completed lifecycle operation.
type Result struct {
	Plan       Plan   `json:"plan"`
	RecordID   string `json:"recordId"`
	DOI        string `json:"doi,omitempty"`
	ConceptDOI string `json:"conceptDoi,omitempty"`
	Executed   bool   `json:"executed"`
}

// AuditSink receives already-redacted lifecycle evidence.
type AuditSink func(AuditEvent)

// Client performs explicitly gated lifecycle writes against the Zenodo sandbox.
type Client struct {
	baseURL          *url.URL
	httpClient       *http.Client
	token            string
	scopes           []Scope
	maxResponseBytes int64
	auditSink        AuditSink
}

// Option configures a lifecycle Client.
type Option func(*Client)

// WithHTTPClient replaces the bounded default HTTP client.
func WithHTTPClient(client *http.Client) Option { return func(c *Client) { c.httpClient = client } }

// WithMaxResponseBytes sets the control-response memory budget.
func WithMaxResponseBytes(limit int64) Option { return func(c *Client) { c.maxResponseBytes = limit } }

// WithAuditSink receives redacted dry-run, success, and failure events.
func WithAuditSink(sink AuditSink) Option { return func(c *Client) { c.auditSink = sink } }

// New constructs a sandbox-only lifecycle client with explicitly declared token scopes.
func New(baseURL, token string, scopes []Scope, options ...Option) (*Client, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("ZENODO_TOKEN is required for sandbox publication actions")
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
		baseURL: parsed, token: token, scopes: append([]Scope(nil), scopes...),
		httpClient: &http.Client{Timeout: defaultTimeout}, maxResponseBytes: defaultResponseLimit,
	}
	for _, option := range options {
		option(client)
	}
	if client.httpClient == nil {
		client.httpClient = &http.Client{Timeout: defaultTimeout}
	}
	if client.maxResponseBytes <= 0 {
		return nil, errors.New("zenodo publication response budget must be positive")
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

// Execute validates the complete request before network I/O. Dry-runs return a
// plan and audit event without constructing an HTTP request.
func (client *Client) Execute(ctx context.Context, request Request, now time.Time) (Result, error) {
	request.Scopes = append([]Scope(nil), client.scopes...)
	plan, err := BuildPlan(request, now)
	if err != nil {
		return Result{}, err
	}
	result := Result{Plan: plan, RecordID: plan.RecordID}
	if plan.DryRun {
		client.emit(plan.Audit("planned", nil))
		return result, nil
	}
	result, err = client.execute(ctx, result)
	if err != nil {
		err = safeError(err, client.token)
		client.emit(plan.Audit("failed", err))
		return Result{}, err
	}
	result.Executed = true
	client.emit(plan.Audit("completed", nil))
	return result, nil
}

type redactedError struct {
	cause   error
	message string
}

func (err redactedError) Error() string { return err.message }
func (err redactedError) Unwrap() error { return err.cause }

func safeError(err error, token string) error {
	message := auth.Redact(err.Error(), token)
	for _, sentinel := range []error{ErrCrossOrigin, ErrResponseTooLarge, ErrInvalidTransition, context.Canceled, context.DeadlineExceeded} {
		if errors.Is(err, sentinel) {
			return redactedError{cause: sentinel, message: message}
		}
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		safeAPIError := &APIError{StatusCode: apiErr.StatusCode, Method: apiErr.Method, Path: apiErr.Path, Message: auth.Redact(apiErr.Message, token)}
		return redactedError{cause: safeAPIError, message: message}
	}
	return errors.New(message)
}

func (client *Client) execute(ctx context.Context, result Result) (Result, error) {
	plan := result.Plan
	switch plan.Action {
	case ActionReserveDOI:
		response, err := client.updateMetadata(ctx, plan, true)
		if err != nil {
			return Result{}, err
		}
		result.DOI = response.Metadata.PrereserveDOI.DOI
		return result, nil
	case ActionPublish:
		if _, err := client.updateMetadata(ctx, plan, false); err != nil {
			return Result{}, fmt.Errorf("apply validated Zenodo metadata before publication: %w", err)
		}
		response, err := client.action(ctx, plan.RecordID, "publish")
		if err != nil {
			return Result{}, err
		}
		result.DOI, result.ConceptDOI = response.DOI, response.ConceptDOI
		return result, nil
	case ActionNewVersion:
		response, err := client.action(ctx, plan.RecordID, "newversion")
		if err != nil {
			return Result{}, err
		}
		latestDraft, err := client.approveLink(response.Links.LatestDraft)
		if err != nil {
			return Result{}, fmt.Errorf("validate Zenodo latest-draft link: %w", err)
		}
		response, err = client.get(ctx, latestDraft)
		if err != nil {
			return Result{}, fmt.Errorf("retrieve Zenodo latest-version draft: %w", err)
		}
		if response.ID == "" {
			return Result{}, errors.New("decode Zenodo latest-version draft: id is missing")
		}
		result.RecordID = response.ID
		return result, nil
	case ActionDiscard:
		endpoint, err := client.resolve("deposit/depositions/" + url.PathEscape(plan.RecordID))
		if err != nil {
			return Result{}, err
		}
		_, err = client.do(ctx, http.MethodDelete, endpoint, nil)
		return result, err
	default:
		return Result{}, ErrInvalidTransition
	}
}

type depositionResponse struct {
	ID         string          `json:"-"`
	RawID      json.RawMessage `json:"id"`
	DOI        string          `json:"doi"`
	ConceptDOI string          `json:"conceptdoi"`
	Links      struct {
		LatestDraft string `json:"latest_draft"`
	} `json:"links"`
	Metadata struct {
		PrereserveDOI struct {
			DOI string `json:"doi"`
		} `json:"prereserve_doi"`
	} `json:"metadata"`
}

type metadataPayload struct {
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	UploadType       string    `json:"upload_type"`
	Creators         []Creator `json:"creators"`
	AccessRight      Access    `json:"access_right"`
	License          string    `json:"license,omitempty"`
	EmbargoDate      string    `json:"embargo_date,omitempty"`
	AccessConditions string    `json:"access_conditions,omitempty"`
}

func (client *Client) updateMetadata(ctx context.Context, plan Plan, reserveDOI bool) (depositionResponse, error) {
	metadata := metadataPayload{
		Title: plan.Metadata.Title, Description: plan.Metadata.Description, UploadType: plan.Metadata.UploadType,
		Creators: append([]Creator(nil), plan.Metadata.Creators...), AccessRight: plan.Metadata.Access,
		License: plan.Metadata.License, AccessConditions: plan.Metadata.AccessConditions,
	}
	if plan.Metadata.EmbargoDate != nil {
		metadata.EmbargoDate = plan.Metadata.EmbargoDate.Format("2006-01-02")
	}
	payload := struct {
		Metadata      metadataPayload `json:"metadata"`
		PrereserveDOI bool            `json:"prereserve_doi,omitempty"`
	}{Metadata: metadata, PrereserveDOI: reserveDOI}
	body, err := json.Marshal(payload)
	if err != nil {
		return depositionResponse{}, fmt.Errorf("encode Zenodo publication metadata: %w", err)
	}
	endpoint, err := client.resolve("deposit/depositions/" + url.PathEscape(plan.RecordID))
	if err != nil {
		return depositionResponse{}, err
	}
	return client.do(ctx, http.MethodPut, endpoint, body)
}

func (client *Client) action(ctx context.Context, recordID, action string) (depositionResponse, error) {
	endpoint, err := client.resolve("deposit/depositions/" + url.PathEscape(recordID) + "/actions/" + action)
	if err != nil {
		return depositionResponse{}, err
	}
	return client.do(ctx, http.MethodPost, endpoint, nil)
}

func (client *Client) get(ctx context.Context, endpoint *url.URL) (depositionResponse, error) {
	return client.do(ctx, http.MethodGet, endpoint, nil)
}

func (client *Client) do(ctx context.Context, method string, endpoint *url.URL, body []byte) (depositionResponse, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint.String(), reader)
	if err != nil {
		return depositionResponse{}, fmt.Errorf("create Zenodo publication request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+client.token)
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := client.httpClient.Do(request)
	if err != nil {
		return depositionResponse{}, fmt.Errorf("execute Zenodo publication request: %w", err)
	}
	payload, err := readBounded(response.Body, client.maxResponseBytes)
	closeErr := response.Body.Close()
	if err != nil {
		return depositionResponse{}, err
	}
	if closeErr != nil {
		return depositionResponse{}, fmt.Errorf("close Zenodo publication response: %w", closeErr)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return depositionResponse{}, &APIError{
			StatusCode: response.StatusCode, Method: method, Path: endpoint.EscapedPath(),
			Message: strings.TrimSpace(auth.Redact(string(payload), client.token)),
		}
	}
	if len(bytes.TrimSpace(payload)) == 0 {
		return depositionResponse{}, nil
	}
	var decoded depositionResponse
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return depositionResponse{}, fmt.Errorf("decode Zenodo publication response: %w", err)
	}
	decoded.ID = rawID(decoded.RawID)
	return decoded, nil
}

func readBounded(reader io.Reader, limit int64) ([]byte, error) {
	payload, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read Zenodo publication response: %w", err)
	}
	if int64(len(payload)) > limit {
		return nil, ErrResponseTooLarge
	}
	return payload, nil
}

func rawID(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(string(raw))
}

func (client *Client) resolve(path string) (*url.URL, error) {
	resolved, err := client.baseURL.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("resolve Zenodo publication endpoint: %w", err)
	}
	if !client.sameAPI(resolved) {
		return nil, ErrCrossOrigin
	}
	return resolved, nil
}

func (client *Client) approveLink(raw string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || !parsed.IsAbs() || !client.sameAPI(parsed) {
		return nil, ErrCrossOrigin
	}
	return parsed, nil
}

func (client *Client) sameAPI(target *url.URL) bool {
	if target == nil || !strings.EqualFold(target.Scheme, client.baseURL.Scheme) || !strings.EqualFold(target.Host, client.baseURL.Host) {
		return false
	}
	return strings.HasPrefix(target.EscapedPath(), client.baseURL.EscapedPath())
}

func (client *Client) emit(event AuditEvent) {
	if client.auditSink != nil {
		client.auditSink(event)
	}
}
