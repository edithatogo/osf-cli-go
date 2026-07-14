// Package observability defines the versioned, redacted operational event
// contract shared by the CLI, API client, transfer code, and MCP server.
package observability

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/edithatogo/osf-cli-go/internal/auth"
)

const (
	SchemaVersion = "osf.event.v1"
	LevelDebug    = "debug"
	LevelInfo     = "info"
	LevelWarn     = "warn"
	LevelError    = "error"
	OutcomeOK     = "ok"
	OutcomeError  = "error"
	OutcomeCancel = "canceled"
)

// Event is the stable JSON envelope emitted for an operational action.
type Event struct {
	SchemaVersion string         `json:"schemaVersion"`
	Timestamp     string         `json:"timestamp"`
	Level         string         `json:"level"`
	Name          string         `json:"name"`
	Provider      string         `json:"provider,omitempty"`
	OperationID   string         `json:"operationId"`
	RequestID     string         `json:"requestId"`
	DurationMS    int64          `json:"durationMs"`
	RetryCount    int            `json:"retryCount"`
	Outcome       string         `json:"outcome"`
	EndpointClass string         `json:"endpointClass,omitempty"`
	Error         *Error         `json:"error,omitempty"`
	Fields        map[string]any `json:"fields,omitempty"`
}

// Error is a redacted, stable error classification.
type Error struct {
	Class   string `json:"class"`
	Message string `json:"message"`
}

// Emitter receives operational events. Implementations must be safe for
// concurrent use because API requests and MCP calls may overlap.
type Emitter interface {
	Emit(Event)
}

type contextKey string

const (
	operationKey contextKey = "osf-operation-id"
	requestKey   contextKey = "osf-request-id"
	emitterKey   contextKey = "osf-event-emitter"
)

// NopEmitter discards events and is the default when observability is disabled.
type NopEmitter struct{}

// Emit implements Emitter.
func (NopEmitter) Emit(Event) {}

// JSONEmitter writes one redacted JSON event per line.
type JSONEmitter struct {
	mu       sync.Mutex
	w        io.Writer
	minLevel string
}

// NewJSONEmitter creates a concurrent JSON event sink with the given minimum level.
func NewJSONEmitter(w io.Writer, minLevel string) *JSONEmitter {
	if w == nil {
		w = io.Discard
	}
	return &JSONEmitter{w: w, minLevel: normalizeLevel(minLevel)}
}

// Emit implements Emitter.
func (e *JSONEmitter) Emit(event Event) {
	if e == nil || !levelAllowed(event.Level, e.minLevel) {
		return
	}
	event = sanitizeEvent(event)
	e.mu.Lock()
	defer e.mu.Unlock()
	_ = json.NewEncoder(e.w).Encode(event)
}

// Close closes the sink when it owns a file-backed writer.
func (e *JSONEmitter) Close() error { return nil }

// FileEmitter is a JSONEmitter with an owned file destination.
type FileEmitter struct {
	*JSONEmitter
	file *os.File
}

// Close closes the owned event log file.
func (e *FileEmitter) Close() error {
	if e == nil || e.file == nil {
		return nil
	}
	return e.file.Close()
}

// OpenFromEnv enables events only when OSF_EVENT_LOG is set. The value may be
// a file path or "stderr". OSF_EVENT_LEVEL defaults to info. No stdout route
// is supported because structured events must never pollute command output.
func OpenFromEnv(stderr io.Writer) (Emitter, io.Closer, error) {
	destination := strings.TrimSpace(os.Getenv("OSF_EVENT_LOG"))
	if destination == "" {
		return NopEmitter{}, io.NopCloser(strings.NewReader("")), nil
	}
	if destination == "stderr" {
		return NewJSONEmitter(stderr, os.Getenv("OSF_EVENT_LEVEL")), io.NopCloser(strings.NewReader("")), nil
	}
	if destination == "stdout" {
		return nil, nil, fmt.Errorf("OSF_EVENT_LOG=stdout is not supported; use stderr or a file")
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return nil, nil, fmt.Errorf("create observability log directory: %w", err)
	}
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, nil, fmt.Errorf("open observability log: %w", err)
	}
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return nil, nil, fmt.Errorf("set observability log permissions: %w", err)
	}
	return &FileEmitter{JSONEmitter: NewJSONEmitter(file, os.Getenv("OSF_EVENT_LEVEL")), file: file}, file, nil
}

// Emit records an event with IDs and redaction applied.
func Emit(ctx context.Context, emitter Emitter, event Event) {
	if emitter == nil {
		return
	}
	if event.SchemaVersion == "" {
		event.SchemaVersion = SchemaVersion
	}
	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if event.Level == "" {
		event.Level = LevelInfo
	}
	if event.OperationID == "" {
		event.OperationID = OperationID(ctx)
	}
	if event.OperationID == "" {
		event.OperationID = NewID("op")
	}
	if event.RequestID == "" {
		event.RequestID = NewID("req")
	}
	if event.Outcome == "" {
		event.Outcome = OutcomeOK
	}
	event.Provider = normalizeProvider(event.Provider)
	emitter.Emit(event)
}

// WithOperationID associates an operation ID with a context.
func WithOperationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, operationKey, strings.TrimSpace(id))
}

// OperationID returns the operation ID associated with a context.
func OperationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(operationKey).(string)
	return id
}

// WithRequestID associates a request ID with a context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestKey, strings.TrimSpace(id))
}

// RequestID returns the request ID associated with a context.
func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(requestKey).(string)
	return id
}

// WithEmitter associates an event sink with a context.
func WithEmitter(ctx context.Context, emitter Emitter) context.Context {
	return context.WithValue(ctx, emitterKey, emitter)
}

// EmitterFromContext returns the event sink associated with a context.
func EmitterFromContext(ctx context.Context) Emitter {
	if ctx == nil {
		return nil
	}
	emitter, _ := ctx.Value(emitterKey).(Emitter)
	return emitter
}

var idCounter uint64

// NewID returns a non-secret operation or request identifier.
func NewID(prefix string) string {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err == nil {
		return strings.TrimSpace(prefix) + "-" + hex.EncodeToString(raw[:])
	}
	return fmt.Sprintf("%s-%d-%d", strings.TrimSpace(prefix), time.Now().UnixNano(), atomic.AddUint64(&idCounter, 1))
}

// ClassifyError maps errors to stable operational classes.
func ClassifyError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return "network"
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "missing osf") || strings.Contains(message, "unauthorized"):
		return "authentication"
	case strings.Contains(message, "forbidden"):
		return "authorization"
	case strings.Contains(message, "not found"):
		return "not_found"
	case strings.Contains(message, "decode") || strings.Contains(message, "json"):
		return "decode"
	case strings.Contains(message, "invalid") || strings.Contains(message, "required"):
		return "validation"
	default:
		return "internal"
	}
}

// EndpointClass returns a low-cardinality class without retaining a URL.
func EndpointClass(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "unknown"
	}
	host := strings.ToLower(parsed.Hostname())
	switch {
	case strings.Contains(host, "files.osf.io"):
		return "storage"
	case strings.Contains(host, "api.osf.io"):
		return "api"
	case host != "":
		return "external"
	default:
		return "unknown"
	}
}

// RedactedError creates a stable error object without exposing credentials.
func RedactedError(err error, secrets ...string) *Error {
	if err == nil {
		return nil
	}
	return &Error{Class: ClassifyError(err), Message: redactMessage(auth.Redact(err.Error(), secrets...))}
}

func sanitizeEvent(event Event) Event {
	event.Provider = normalizeProvider(event.Provider)
	event.Fields = redactMap(event.Fields)
	if event.Error != nil {
		event.Error = &Error{Class: event.Error.Class, Message: redactMessage(auth.Redact(event.Error.Message))}
	}
	return event
}

func normalizeProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "osf", "zenodo", "cross-provider":
		return strings.ToLower(strings.TrimSpace(provider))
	default:
		return "unknown"
	}
}

func redactMap(fields map[string]any) map[string]any {
	if fields == nil {
		return nil
	}
	result := make(map[string]any, len(fields))
	for key, value := range fields {
		lower := strings.ToLower(strings.ReplaceAll(key, "_", ""))
		switch {
		case strings.Contains(lower, "token"), strings.Contains(lower, "password"), strings.Contains(lower, "authorization"), strings.Contains(lower, "secret"), strings.Contains(lower, "credential"), strings.Contains(lower, "apikey"):
			result[key] = "[REDACTED]"
		case strings.Contains(lower, "path"), lower == "source", lower == "destination", lower == "filepath":
			result[key] = "[REDACTED_PATH]"
		case lower == "error":
			result[key] = redactMessage(auth.Redact(fmt.Sprint(value)))
		case lower == "fields":
			if nested, ok := value.(map[string]any); ok {
				result[key] = redactMap(nested)
			} else {
				result[key] = auth.Redact(fmt.Sprint(value))
			}
		default:
			result[key] = redactValue(value)
		}
	}
	return result
}

func redactValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return redactMap(typed)
	case []any:
		items := make([]any, len(typed))
		for i, item := range typed {
			items[i] = redactValue(item)
		}
		return items
	case string:
		return redactMessage(auth.Redact(typed))
	default:
		return value
	}
}

func redactMessage(message string) string {
	if message == "" {
		return message
	}
	home, _ := os.UserHomeDir()
	if home != "" {
		message = strings.ReplaceAll(message, home, "[REDACTED_PATH]")
	}
	working, _ := os.Getwd()
	if working != "" {
		message = strings.ReplaceAll(message, working, "[REDACTED_PATH]")
	}
	parts := strings.Fields(message)
	for i, part := range parts {
		trimmed := strings.Trim(part, "\"'(),;:")
		if strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, "~/") || (len(trimmed) > 2 && trimmed[1] == ':' && (trimmed[2] == '\\' || trimmed[2] == '/')) {
			parts[i] = "[REDACTED_PATH]"
		}
	}
	return strings.Join(parts, " ")
}

func normalizeLevel(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case LevelDebug, LevelWarn, LevelError:
		return strings.ToLower(strings.TrimSpace(level))
	default:
		return LevelInfo
	}
}

func levelAllowed(level, minimum string) bool {
	rank := map[string]int{LevelDebug: 0, LevelInfo: 1, LevelWarn: 2, LevelError: 3}
	return rank[normalizeLevel(level)] >= rank[normalizeLevel(minimum)]
}
