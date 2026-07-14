package zenodooai

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/observability"
	"github.com/edithatogo/osf-cli-go/internal/repository"
)

const (
	defaultBaseURL          = "https://zenodo.org/oai2d"
	defaultTimeout          = 30 * time.Second
	defaultMaxResponseBytes = int64(8 << 20)
	defaultMaxPages         = 100
	defaultMaxRecords       = 5000
	defaultMaxConcurrency   = 2
	defaultMaxRetries       = 2
	defaultRetryDelay       = 200 * time.Millisecond
)

// Client is a bounded public OAI-PMH client. OAI-PMH does not use Zenodo API tokens.
type Client struct {
	baseURL          *url.URL
	httpClient       *http.Client
	emitter          observability.Emitter
	maxResponseBytes int64
	maxPages         int
	maxRecords       int
	maxConcurrency   int
	maxRetries       int
	retryDelay       time.Duration
	now              func() time.Time
	semaphore        chan struct{}
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient replaces the bounded default HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) { c.httpClient = httpClient }
}

// WithObserver enables provider-tagged structured request events.
func WithObserver(emitter observability.Emitter) Option {
	return func(c *Client) { c.emitter = emitter }
}

// WithLimits configures response bytes, page count, and record count budgets.
func WithLimits(responseBytes int64, pages, records int) Option {
	return func(c *Client) { c.maxResponseBytes, c.maxPages, c.maxRecords = responseBytes, pages, records }
}

// WithMaxConcurrency bounds simultaneous OAI requests made by this client.
func WithMaxConcurrency(limit int) Option { return func(c *Client) { c.maxConcurrency = limit } }

// WithRetryPolicy configures retries for transient transport failures.
func WithRetryPolicy(maxRetries int, delay time.Duration) Option {
	return func(c *Client) { c.maxRetries, c.retryDelay = maxRetries, delay }
}

// WithClock supplies a clock for deterministic resumption-expiry checks.
func WithClock(now func() time.Time) Option { return func(c *Client) { c.now = now } }

// New constructs a public Zenodo OAI-PMH client.
func New(baseURL string, options ...Option) (*Client, error) {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse Zenodo OAI-PMH URL: %w", err)
	}
	host := strings.ToLower(parsed.Hostname())
	local := host == "localhost" || host == "127.0.0.1" || host == "::1"
	if parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Scheme != "https" && (!local || parsed.Scheme != "http")) {
		return nil, errors.New("zenodo OAI-PMH URL must be plain HTTPS without credentials, query, or fragment")
	}
	if !local && host != "zenodo.org" && host != "sandbox.zenodo.org" {
		return nil, fmt.Errorf("unapproved Zenodo OAI-PMH host %q", parsed.Host)
	}
	client := &Client{
		baseURL: parsed, httpClient: &http.Client{Timeout: defaultTimeout}, maxResponseBytes: defaultMaxResponseBytes,
		maxPages: defaultMaxPages, maxRecords: defaultMaxRecords, maxConcurrency: defaultMaxConcurrency, maxRetries: defaultMaxRetries,
		retryDelay: defaultRetryDelay, now: time.Now,
	}
	for _, option := range options {
		option(client)
	}
	if client.httpClient == nil {
		client.httpClient = &http.Client{Timeout: defaultTimeout}
	}
	if client.now == nil || client.maxResponseBytes <= 0 || client.maxPages <= 0 || client.maxRecords <= 0 || client.maxConcurrency <= 0 || client.maxRetries < 0 || client.retryDelay < 0 {
		return nil, errors.New("zenodo OAI-PMH response, page, and record limits must be positive and retries non-negative")
	}
	configured := *client.httpClient
	previousRedirect := configured.CheckRedirect
	configured.CheckRedirect = func(request *http.Request, via []*http.Request) error {
		if !client.sameOrigin(request.URL) {
			return ErrCrossOrigin
		}
		if previousRedirect != nil {
			return previousRedirect(request, via)
		}
		if len(via) >= 10 {
			return errors.New("stopped after 10 Zenodo OAI-PMH redirects")
		}
		return nil
	}
	client.httpClient = &configured
	client.semaphore = make(chan struct{}, client.maxConcurrency)
	return client, nil
}

// ListRecords retrieves one deterministic page. Callers can persist Page.Next.
func (client *Client) ListRecords(ctx context.Context, request Request) (Page, error) {
	query, prefix, setName, err := client.recordQuery(request)
	if err != nil {
		return Page{}, err
	}
	envelope, err := client.fetch(ctx, "ListRecords", query)
	if err != nil {
		return Page{}, err
	}
	responseDate, err := envelope.responseTime()
	if err != nil {
		return Page{}, fmt.Errorf("decode Zenodo OAI-PMH response date: %w", err)
	}
	next, err := envelope.ListRecords.Token.value()
	if err != nil {
		return Page{}, fmt.Errorf("decode Zenodo OAI-PMH resumption expiry: %w", err)
	}
	page := Page{Next: next}
	page.Next.MetadataPrefix, page.Next.Set = prefix, setName
	for _, native := range envelope.ListRecords.Records {
		if strings.TrimSpace(native.Header.Identifier) == "" || strings.TrimSpace(native.Header.Datestamp) == "" {
			return Page{}, errors.New("decode Zenodo OAI-PMH record: identifier and datestamp are required")
		}
		var metadata *repository.NativeMetadata
		if raw := []byte(native.Metadata.Value); len(bytes.TrimSpace(raw)) > 0 {
			value, err := repository.NewNativeMetadata("application/xml", raw)
			if err != nil {
				return Page{}, fmt.Errorf("preserve OAI-PMH metadata for %q: %w", native.Header.Identifier, err)
			}
			metadata = &value
		}
		if native.Header.Status != "deleted" && metadata == nil {
			return Page{}, fmt.Errorf("decode Zenodo OAI-PMH record %q: metadata is required", native.Header.Identifier)
		}
		page.Records = append(page.Records, Record{
			Header:         Header{Identifier: native.Header.Identifier, Datestamp: native.Header.Datestamp, SetSpecs: append([]string(nil), native.Header.SetSpecs...), Deleted: native.Header.Status == "deleted"},
			NativeMetadata: metadata, AboutXML: []byte(strings.TrimSpace(native.About.Value)),
			Provenance: Provenance{BaseURL: client.baseURL.String(), ResponseDate: responseDate, MetadataPrefix: prefix, Set: setName, Datestamp: native.Header.Datestamp},
		})
	}
	return page, nil
}

// Harvest follows resumption tokens within configured page and record budgets.
func (client *Client) Harvest(ctx context.Context, request Request) ([]Record, error) {
	var records []Record
	for pageNumber := 0; ; pageNumber++ {
		if pageNumber >= client.maxPages {
			return nil, ErrPageLimit
		}
		page, err := client.ListRecords(ctx, request)
		if err != nil {
			return nil, err
		}
		if len(records)+len(page.Records) > client.maxRecords {
			return nil, ErrRecordLimit
		}
		records = append(records, page.Records...)
		if page.Next.Empty() {
			return records, nil
		}
		request = Request{Token: page.Next}
	}
}

// ListSets returns all advertised sets using bounded resumption.
func (client *Client) ListSets(ctx context.Context) ([]Set, error) {
	var sets []Set
	var token ResumptionToken
	for page := 0; ; page++ {
		if page >= client.maxPages {
			return nil, ErrPageLimit
		}
		query := url.Values{"verb": {"ListSets"}}
		if !token.Empty() {
			if err := client.validateToken(token); err != nil {
				return nil, err
			}
			query = url.Values{"verb": {"ListSets"}, "resumptionToken": {token.Value}}
		}
		envelope, err := client.fetch(ctx, "ListSets", query)
		if err != nil {
			return nil, err
		}
		for _, item := range envelope.ListSets.Sets {
			sets = append(sets, Set{Spec: item.Spec, Name: item.Name, Description: []byte(strings.TrimSpace(item.Description.Value))})
		}
		token, err = envelope.ListSets.Token.value()
		if err != nil {
			return nil, fmt.Errorf("decode Zenodo OAI-PMH resumption expiry: %w", err)
		}
		if token.Empty() {
			return sets, nil
		}
	}
}

// ListMetadataFormats returns schemas available for the repository or identifier.
func (client *Client) ListMetadataFormats(ctx context.Context, identifier string) ([]MetadataFormat, error) {
	query := url.Values{"verb": {"ListMetadataFormats"}}
	if identifier = strings.TrimSpace(identifier); identifier != "" {
		query.Set("identifier", identifier)
	}
	envelope, err := client.fetch(ctx, "ListMetadataFormats", query)
	if err != nil {
		return nil, err
	}
	formats := make([]MetadataFormat, 0, len(envelope.Formats.Formats))
	for _, item := range envelope.Formats.Formats {
		formats = append(formats, MetadataFormat{Prefix: item.Prefix, Schema: item.Schema, NamespaceURL: item.Namespace})
	}
	return formats, nil
}

func (client *Client) recordQuery(request Request) (url.Values, string, string, error) {
	if !request.Token.Empty() {
		if strings.TrimSpace(request.MetadataPrefix) != "" || strings.TrimSpace(request.Set) != "" || !request.From.IsZero() || !request.Until.IsZero() {
			return nil, "", "", errors.New("zenodo OAI-PMH resumption token cannot be combined with other selectors")
		}
		if err := client.validateToken(request.Token); err != nil {
			return nil, "", "", err
		}
		return url.Values{"verb": {"ListRecords"}, "resumptionToken": {request.Token.Value}}, request.Token.MetadataPrefix, request.Token.Set, nil
	}
	prefix := strings.TrimSpace(request.MetadataPrefix)
	if prefix == "" {
		return nil, "", "", errors.New("zenodo OAI-PMH metadata prefix is required")
	}
	if !request.From.IsZero() && !request.Until.IsZero() && request.From.After(request.Until) {
		return nil, "", "", errors.New("zenodo OAI-PMH from date must not be after until date")
	}
	query := url.Values{"verb": {"ListRecords"}, "metadataPrefix": {prefix}}
	setName := strings.TrimSpace(request.Set)
	if setName != "" {
		query.Set("set", setName)
	}
	if !request.From.IsZero() {
		query.Set("from", request.From.UTC().Format(time.RFC3339))
	}
	if !request.Until.IsZero() {
		query.Set("until", request.Until.UTC().Format(time.RFC3339))
	}
	return query, prefix, setName, nil
}

func (client *Client) validateToken(token ResumptionToken) error {
	if token.Empty() {
		return errors.New("zenodo OAI-PMH resumption token is empty")
	}
	if !token.ExpiresAt.IsZero() && !client.now().Before(token.ExpiresAt) {
		return ErrTokenExpired
	}
	return nil
}

func (client *Client) fetch(ctx context.Context, verb string, query url.Values) (oaiEnvelope, error) {
	select {
	case client.semaphore <- struct{}{}:
		defer func() { <-client.semaphore }()
	case <-ctx.Done():
		return oaiEnvelope{}, ctx.Err()
	}
	endpoint := *client.baseURL
	endpoint.RawQuery = query.Encode()
	started := time.Now()
	var finalErr error
	retries := 0
	defer func() {
		outcome, level := observability.OutcomeOK, observability.LevelInfo
		if finalErr != nil {
			outcome, level = observability.OutcomeError, observability.LevelError
			if errors.Is(finalErr, context.Canceled) {
				outcome = observability.OutcomeCancel
			}
		}
		observability.Emit(ctx, client.emitter, observability.Event{Name: "api.request", Level: level, Outcome: outcome, DurationMS: time.Since(started).Milliseconds(), RetryCount: retries, EndpointClass: "zenodo_oai_pmh", Error: observability.RedactedError(finalErr), Fields: map[string]any{"provider": "zenodo", "protocol": "oai-pmh", "verb": verb}})
	}()
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
		if err != nil {
			finalErr = err
			return oaiEnvelope{}, err
		}
		req.Header.Set("Accept", "application/xml, text/xml")
		response, err := client.httpClient.Do(req)
		if err != nil {
			if attempt < client.maxRetries && ctx.Err() == nil {
				retries++
				if err := sleep(ctx, client.retryDelay); err != nil {
					finalErr = err
					return oaiEnvelope{}, err
				}
				continue
			}
			finalErr = err
			return oaiEnvelope{}, err
		}
		if retryable(response.StatusCode) && attempt < client.maxRetries {
			_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
			_ = response.Body.Close()
			retries++
			if err := sleep(ctx, parseRetryAfter(response.Header.Get("Retry-After"), client.retryDelay)); err != nil {
				finalErr = err
				return oaiEnvelope{}, err
			}
			continue
		}
		body, readErr := readBounded(response.Body, client.maxResponseBytes)
		closeErr := response.Body.Close()
		if readErr != nil {
			finalErr = readErr
			return oaiEnvelope{}, readErr
		}
		if closeErr != nil {
			finalErr = closeErr
			return oaiEnvelope{}, closeErr
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			httpErr := &HTTPError{StatusCode: response.StatusCode, RetryAfter: parseRetryAfter(response.Header.Get("Retry-After"), 0)}
			finalErr = httpErr
			return oaiEnvelope{}, httpErr
		}
		envelope, err := decodeEnvelope(body, verb)
		if err != nil {
			finalErr = err
			return oaiEnvelope{}, finalErr
		}
		if envelope.Error.Code != "" {
			protocolErr := &ProtocolError{Code: envelope.Error.Code, Message: strings.TrimSpace(envelope.Error.Message), Verb: verb}
			finalErr = protocolErr
			return oaiEnvelope{}, protocolErr
		}
		return envelope, nil
	}
}

func decodeEnvelope(body []byte, verb string) (oaiEnvelope, error) {
	var envelope oaiEnvelope
	decoder := xml.NewDecoder(bytes.NewReader(body))
	decoder.Strict = true
	if err := decoder.Decode(&envelope); err != nil {
		return oaiEnvelope{}, fmt.Errorf("decode Zenodo OAI-PMH %s response: %w", verb, err)
	}
	if envelope.XMLName.Space != "http://www.openarchives.org/OAI/2.0/" {
		return oaiEnvelope{}, errors.New("decode Zenodo OAI-PMH response: unexpected XML namespace")
	}
	return envelope, nil
}

func (client *Client) sameOrigin(candidate *url.URL) bool {
	return candidate != nil && candidate.Scheme == client.baseURL.Scheme && strings.EqualFold(candidate.Host, client.baseURL.Host)
}

func readBounded(reader io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, ErrResponseTooLarge
	}
	return data, nil
}

func retryable(status int) bool {
	return status == http.StatusTooManyRequests || status == http.StatusBadGateway || status == http.StatusServiceUnavailable || status == http.StatusGatewayTimeout
}

func parseRetryAfter(value string, fallback time.Duration) time.Duration {
	seconds, err := strconv.Atoi(strings.TrimSpace(value))
	if err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return fallback
}

func sleep(ctx context.Context, delay time.Duration) error {
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
