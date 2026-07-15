package main

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type failingTransport struct{ err error }

func (t failingTransport) Connect(context.Context) (mcp.Connection, error) { return nil, t.err }

func TestRunReturnsTransportError(t *testing.T) {
	want := errors.New("transport unavailable")
	if err := run(context.Background(), io.Discard, failingTransport{err: want}); !errors.Is(err, want) {
		t.Fatalf("run error = %v, want %v", err, want)
	}
}

func TestRunRejectsInvalidEventLog(t *testing.T) {
	t.Setenv("OSF_EVENT_LOG", "stdout")
	if err := run(context.Background(), io.Discard, failingTransport{err: errors.New("unused")}); err == nil {
		t.Fatal("run succeeded with invalid event log destination")
	}
}
