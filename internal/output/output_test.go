package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestWriteJSONKeepsHTMLCharacters(t *testing.T) {
	var buf bytes.Buffer

	if err := WriteJSON(&buf, map[string]string{"value": "<tag>&"}); err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}

	got := buf.String()
	if !strings.HasSuffix(got, "\n") {
		t.Fatalf("WriteJSON output = %q, want trailing newline", got)
	}
	if strings.Contains(got, "\\u003c") || strings.Contains(got, "\\u003e") || strings.Contains(got, "\\u0026") {
		t.Fatalf("WriteJSON escaped HTML characters unexpectedly: %q", got)
	}

	var decoded map[string]string
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("WriteJSON output is not valid JSON: %v", err)
	}
	if decoded["value"] != "<tag>&" {
		t.Fatalf("decoded value = %q, want %q", decoded["value"], "<tag>&")
	}
}

func TestWriteJSONWithNilValue(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteJSON(&buf, nil); err != nil {
		t.Fatalf("WriteJSON returned error: %v", err)
	}
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Fatalf("WriteJSON output = %q, want trailing newline", buf.String())
	}
}

func TestWriteJSONPropagatesWriterError(t *testing.T) {
	err := WriteJSON(failingWriter{}, map[string]string{"value": "x"})
	if err == nil {
		t.Fatal("WriteJSON returned nil error, want write failure")
	}
}

func TestWriteTableWritesHeadersAndRows(t *testing.T) {
	var buf bytes.Buffer

	if err := WriteTable(&buf, []string{"ID", "NAME"}, [][]string{{"1", "Alpha"}, {"2", "Beta"}}); err != nil {
		t.Fatalf("WriteTable returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "ID") || !strings.Contains(got, "Beta") {
		t.Fatalf("WriteTable output missing expected rows: %q", got)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Fatalf("WriteTable output = %q, want trailing newline", got)
	}
}

func TestWriteTableWithOnlyHeaders(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTable(&buf, []string{"ID", "NAME"}, nil); err != nil {
		t.Fatalf("WriteTable returned error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "ID") || !strings.Contains(got, "NAME") {
		t.Fatalf("WriteTable output missing headers: %q", got)
	}
	if !strings.HasSuffix(got, "\n") {
		t.Fatalf("WriteTable output = %q, want trailing newline", got)
	}
}

func TestWriteTableWithEmptyHeadersAndNilRows(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteTable(&buf, nil, nil); err != nil {
		t.Fatalf("WriteTable returned error: %v", err)
	}
}

func TestWriteTablePropagatesWriterError(t *testing.T) {
	err := WriteTable(failingWriter{}, []string{"ID"}, [][]string{{"1"}})
	if err == nil {
		t.Fatal("WriteTable returned nil error, want write failure")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}
