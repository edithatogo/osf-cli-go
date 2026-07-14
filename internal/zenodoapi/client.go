package zenodoapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/auth"
	"github.com/edithatogo/osf-cli-go/internal/observability"
)

const (
	defaultBaseURL          = "https://zenodo.org/api/"
	defaultHTTPTimeout      = 30 * time.Second
	defaultMaxResponseBytes = int64(8 << 20)
	defaultMaxPages         = 100
	defaultMaxConcurrency   = 4
	defaultMaxRetries       = 2
	defaultRetryDelay       = 200 * time.Millisecond
)

var (
	// ErrResponseTooLarge indicates a response exceeded the configured memory budget.
	ErrResponseTooLarge = errors.New("zenodo response exceeds configured size limit")
	// ErrPaginationLimit indicates a search exceeded its page budget.
	ErrPaginationLimit = errors.New("zenodo pagination exceeds configured page limit")
	// ErrCrossOriginPagination indicates an upstream link escaped the API origin or path.
	ErrCrossOriginPagination = errors.New("zenodo pagination link leaves configured API origin")
)

// Client is a bounded, read-only client for published Zenodo REST records.
type Client struct {
	baseURL          *url.URL
	httpClient       *http.Client
	token            string
	emitter          observability.Emitter
	maxResponseBytes int64
	maxPages         int
	maxConcurrency   int
	maxRetries       int
	retryDelay       time.Duration
	rateMu           sync.RWMutex
	lastRateLimit    RateLimit
	semaphore        chan struct{}
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient sets the HTTP transport. Nil restores the bounded default.
func WithHTTPClient(client *http.Client) Option { return func(c *Client) { c.httpClient = client } }

// WithToken configures an optional bearer token. Public reads do not require it.
func WithToken(token string) Option { return func(c *Client) { c.token = strings.TrimSpace(token) } }

// WithObserver enables redacted provider-tagged request events.
func WithObserver(emitter observability.Emitter) Option {
	return func(c *Client) { c.emitter = emitter }
}

// WithMaxResponseBytes sets the per-response in-memory budget.
func WithMaxResponseBytes(limit int64) Option { return func(c *Client) { c.maxResponseBytes = limit } }

// WithMaxPages sets the automatic search pagination budget.
func WithMaxPages(limit int) Option { return func(c *Client) { c.maxPages = limit } }

// WithMaxConcurrency bounds simultaneous requests made by this client.
func WithMaxConcurrency(limit int) Option { return func(c *Client) { c.maxConcurrency = limit } }

// WithRetryPolicy sets retry count and base delay for idempotent GET requests.
func WithRetryPolicy(maxRetries int, baseDelay time.Duration) Option {
	return func(c *Client) { c.maxRetries, c.retryDelay = maxRetries, baseDelay }
}

// New constructs a read-only Zenodo client.
func New(baseURL string, options ...Option) (*Client, error) {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse Zenodo base URL: %w", err)
	}
	if (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, errors.New("zenodo base URL must be a plain HTTP(S) URL without credentials, query, or fragment")
	}
	host := strings.ToLower(parsed.Hostname())
	isLocal := host == "127.0.0.1" || host == "localhost" || host == "::1"
	if parsed.Scheme == "http" && !isLocal {
		return nil, errors.New("zenodo base URL requires HTTPS except for local tests")
	}
	if !isLocal && host != "zenodo.org" && host != "sandbox.zenodo.org" {
		return nil, fmt.Errorf("unapproved Zenodo API host %q", parsed.Host)
	}
	if !strings.HasSuffix(parsed.Path, "/") {
		parsed.Path += "/"
	}
	client := &Client{
		baseURL: parsed, httpClient: &http.Client{Timeout: defaultHTTPTimeout},
		maxResponseBytes: defaultMaxResponseBytes, maxPages: defaultMaxPages, maxConcurrency: defaultMaxConcurrency,
		maxRetries: defaultMaxRetries, retryDelay: defaultRetryDelay,
	}
	for _, option := range options {
		option(client)
	}
	if client.httpClient == nil {
		client.httpClient = &http.Client{Timeout: defaultHTTPTimeout}
	}
	if client.maxResponseBytes <= 0 || client.maxPages <= 0 || client.maxConcurrency <= 0 || client.maxRetries < 0 || client.retryDelay < 0 {
		return nil, errors.New("zenodo response, page, and retry budgets must be positive and retries non-negative")
	}
	configuredHTTP := *client.httpClient
	previousRedirect := configuredHTTP.CheckRedirect
	configuredHTTP.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		if !client.sameAPI(request.URL) {
			return ErrCrossOriginPagination
		}
		if previousRedirect != nil {
			return previousRedirect(request, via)
		}
		if len(via) >= 10 {
			return errors.New("stopped after 10 Zenodo redirects")
		}
		return nil
	}
	client.httpClient = &configuredHTTP
	client.semaphore = make(chan struct{}, client.maxConcurrency)
	return client, nil
}

// LastRateLimit returns the most recently observed response limit headers.
func (client *Client) LastRateLimit() RateLimit {
	client.rateMu.RLock()
	defer client.rateMu.RUnlock()
	return client.lastRateLimit
}

// SearchRecords searches published records and follows bounded same-origin pagination.
func (client *Client) SearchRecords(ctx context.Context, query string, limit int) ([]Record, error) {
	if limit < 0 {
		return nil, errors.New("zenodo search limit must not be negative")
	}
	endpoint, err := client.resolve("records/")
	if err != nil {
		return nil, err
	}
	params := endpoint.Query()
	if query = strings.TrimSpace(query); query != "" {
		params.Set("q", query)
	}
	pageSize := 100
	if limit > 0 && limit < pageSize {
		pageSize = limit
	}
	params.Set("size", strconv.Itoa(pageSize))
	endpoint.RawQuery = params.Encode()

	var records []Record
	seen := make(map[string]bool)
	for pageNumber := 0; endpoint != nil; pageNumber++ {
		if pageNumber >= client.maxPages {
			return nil, ErrPaginationLimit
		}
		if seen[endpoint.String()] {
			return nil, errors.New("zenodo pagination cycle detected")
		}
		seen[endpoint.String()] = true
		body, _, err := client.get(ctx, endpoint)
		if err != nil {
			return nil, err
		}
		page, next, err := decodeSearchPage(body)
		if err != nil {
			return nil, fmt.Errorf("decode zenodo search response: %w", err)
		}
		for _, record := range page {
			if limit > 0 && len(records) >= limit {
				return records, nil
			}
			records = append(records, record)
		}
		if limit > 0 && len(records) >= limit {
			return records, nil
		}
		if next == "" {
			break
		}
		endpoint, err = client.resolvePaginationFrom(endpoint, next)
		if err != nil {
			return nil, err
		}
	}
	return records, nil
}

// GetRecord returns one published Zenodo record.
func (client *Client) GetRecord(ctx context.Context, id string) (Record, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return Record{}, errors.New("zenodo record id is required")
	}
	endpoint, err := client.resolve("records/" + url.PathEscape(id))
	if err != nil {
		return Record{}, err
	}
	body, _, err := client.get(ctx, endpoint)
	if err != nil {
		return Record{}, err
	}
	var record Record
	if err := json.Unmarshal(body, &record); err != nil {
		return Record{}, fmt.Errorf("decode zenodo record response: %w", err)
	}
	if record.ID == "" {
		return Record{}, errors.New("decode zenodo record response: id is missing")
	}
	return record, nil
}

// ListRecordFiles returns the files embedded in a published record response.
func (client *Client) ListRecordFiles(ctx context.Context, id string) ([]File, error) {
	record, err := client.GetRecord(ctx, id)
	if err != nil {
		return nil, err
	}
	return append([]File(nil), record.Files...), nil
}

func (client *Client) resolve(reference string) (*url.URL, error) {
	resolved, err := client.baseURL.Parse(reference)
	if err != nil {
		return nil, err
	}
	if !client.sameAPI(resolved) {
		return nil, ErrCrossOriginPagination
	}
	return resolved, nil
}

func (client *Client) resolvePagination(reference string) (*url.URL, error) {
	return client.resolvePaginationFrom(client.baseURL, reference)
}

func (client *Client) resolvePaginationFrom(current *url.URL, reference string) (*url.URL, error) {
	next, err := url.Parse(reference)
	if err != nil {
		return nil, fmt.Errorf("parse Zenodo pagination link: %w", err)
	}
	resolved := current.ResolveReference(next)
	if !client.sameAPI(resolved) {
		return nil, ErrCrossOriginPagination
	}
	return resolved, nil
}

func (client *Client) sameAPI(candidate *url.URL) bool {
	return candidate != nil && candidate.Scheme == client.baseURL.Scheme && strings.EqualFold(candidate.Host, client.baseURL.Host) && strings.HasPrefix(candidate.Path, client.baseURL.Path)
}

func (client *Client) get(ctx context.Context, endpoint *url.URL) ([]byte, *http.Response, error) {
	select {
	case client.semaphore <- struct{}{}:
		defer func() { <-client.semaphore }()
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	}
	started := time.Now()
	retries := 0
	var finalStatus int
	var finalErr error
	defer func() {
		outcome := observability.OutcomeOK
		level := observability.LevelInfo
		if finalErr != nil {
			outcome = observability.OutcomeError
			level = observability.LevelError
			if errors.Is(finalErr, context.Canceled) {
				outcome = observability.OutcomeCancel
			}
		}
		observability.Emit(ctx, client.emitter, observability.Event{
			Name: "api.request", Level: level, Outcome: outcome,
			DurationMS: time.Since(started).Milliseconds(), RetryCount: retries,
			EndpointClass: "zenodo_api", Error: observability.RedactedError(finalErr, client.token),
			Fields: map[string]any{"provider": "zenodo", "method": http.MethodGet, "status": finalStatus},
		})
	}()

	for attempt := 0; ; attempt++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
		if err != nil {
			finalErr = err
			return nil, nil, err
		}
		request.Header.Set("Accept", "application/json")
		if client.token != "" {
			request.Header.Set("Authorization", "Bearer "+client.token)
		}
		response, err := client.httpClient.Do(request)
		if err != nil {
			if attempt < client.maxRetries && ctx.Err() == nil {
				retries++
				if err := sleepContext(ctx, client.retryDelay); err != nil {
					finalErr = err
					return nil, nil, err
				}
				continue
			}
			finalErr = err
			return nil, nil, err
		}
		finalStatus = response.StatusCode
		rateLimit := parseRateLimit(response.Header)
		client.rateMu.Lock()
		client.lastRateLimit = rateLimit
		client.rateMu.Unlock()
		if retryableStatus(response.StatusCode) && attempt < client.maxRetries {
			_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4<<10))
			_ = response.Body.Close()
			retries++
			delay := retryDelay(response.Header.Get("Retry-After"), client.retryDelay)
			if err := sleepContext(ctx, delay); err != nil {
				finalErr = err
				return nil, nil, err
			}
			continue
		}
		body, readErr := readBounded(response.Body, client.maxResponseBytes)
		closeErr := response.Body.Close()
		if readErr != nil {
			finalErr = readErr
			return nil, response, readErr
		}
		if closeErr != nil {
			finalErr = closeErr
			return nil, response, closeErr
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			apiErr := parseAPIError(response, body, client.token)
			finalErr = apiErr
			return nil, response, apiErr
		}
		return body, response, nil
	}
}

func decodeSearchPage(data []byte) ([]Record, string, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, "", errors.New("empty response")
	}
	if data[0] == '[' {
		var records []Record
		return records, "", json.Unmarshal(data, &records)
	}
	var outer struct {
		Hits  json.RawMessage `json:"hits"`
		Links struct {
			Next string `json:"next"`
		} `json:"links"`
	}
	if err := json.Unmarshal(data, &outer); err != nil {
		return nil, "", err
	}
	if len(outer.Hits) == 0 || bytes.Equal(bytes.TrimSpace(outer.Hits), []byte("null")) {
		return nil, "", errors.New("search response is missing hits")
	}
	var hits struct {
		Hits []Record `json:"hits"`
	}
	if err := json.Unmarshal(outer.Hits, &hits); err != nil {
		return nil, "", err
	}
	return hits.Hits, strings.TrimSpace(outer.Links.Next), nil
}

func parseAPIError(response *http.Response, body []byte, token string) *APIError {
	var payload struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}
	_ = json.Unmarshal(body, &payload)
	message := strings.TrimSpace(payload.Message)
	if message == "" && len(body) > 0 {
		message = strings.TrimSpace(string(body))
	}
	message = auth.Redact(message, token)
	if len(message) > 1024 {
		message = message[:1024] + "..."
	}
	return &APIError{
		StatusCode: response.StatusCode, Method: response.Request.Method,
		Path: response.Request.URL.Path, Message: message,
		RetryAfter: retryDelay(response.Header.Get("Retry-After"), 0),
		RateLimit:  parseRateLimit(response.Header),
	}
}

func parseRateLimit(header http.Header) RateLimit {
	return RateLimit{
		Limit:     parseInt64(header.Get("X-RateLimit-Limit")),
		Remaining: parseInt64(header.Get("X-RateLimit-Remaining")),
		ResetUnix: parseInt64(header.Get("X-RateLimit-Reset")),
	}
}

func parseInt64(value string) int64 {
	parsed, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	return parsed
}

func retryableStatus(status int) bool {
	return status == http.StatusTooManyRequests || status == http.StatusBadGateway || status == http.StatusServiceUnavailable || status == http.StatusGatewayTimeout
}

func retryDelay(value string, fallback time.Duration) time.Duration {
	value = strings.TrimSpace(value)
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		delay := time.Duration(seconds) * time.Second
		if delay > 5*time.Second {
			return 5 * time.Second
		}
		return delay
	}
	if when, err := http.ParseTime(value); err == nil {
		delay := time.Until(when)
		if delay < 0 {
			return 0
		}
		if delay > 5*time.Second {
			return 5 * time.Second
		}
		return delay
	}
	return fallback
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return ctx.Err()
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
