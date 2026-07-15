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

func TestRunServesAndStopsInMemoryMCP(t *testing.T) {
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runDone := make(chan error, 1)
	go func() {
		runDone <- run(ctx, io.Discard, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "coverage-client", Version: "1.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("connect MCP client: %v", err)
	}
	if err := session.Close(); err != nil {
		t.Fatalf("close MCP session: %v", err)
	}
	if err := <-runDone; err != nil {
		t.Fatalf("run returned error: %v", err)
	}
}
