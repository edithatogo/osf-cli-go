package main

import (
	"context"
	"os"

	"github.com/edithatogo/osf-cli-go/internal/auth"
	"github.com/edithatogo/osf-cli-go/internal/buildinfo"
	"github.com/edithatogo/osf-cli-go/internal/mcpserver"
	"github.com/edithatogo/osf-cli-go/internal/observability"
	"github.com/edithatogo/osf-cli-go/internal/osfapi"
	"github.com/edithatogo/osf-cli-go/internal/zenodooai"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var version = "0.0.0-dev"

func main() {
	effectiveVersion := buildinfo.Version(version)
	emitter, closer, err := observability.OpenFromEnv(os.Stderr)
	if err != nil {
		_, _ = os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(2)
	}
	defer func() { _ = closer.Close() }()
	ctx := observability.WithEmitter(observability.WithOperationID(context.Background(), observability.NewID("op")), emitter)
	observability.Emit(ctx, emitter, observability.Event{
		Name:   "mcp.server",
		Fields: map[string]any{"state": "starting", "transport": "stdio", "version": effectiveVersion},
	})
	credentials, _ := auth.LoadCredentials(auth.EnvSource{})
	client, err := osfapi.New("", osfapi.WithCredentials(credentials), osfapi.WithObserver(emitter))
	if err != nil {
		observability.Emit(ctx, emitter, observability.Event{Level: observability.LevelError, Name: "mcp.server", Outcome: observability.OutcomeError, Error: observability.RedactedError(err)})
		os.Exit(1)
	}

	oai, err := zenodooai.New("", zenodooai.WithObserver(emitter))
	if err != nil {
		observability.Emit(ctx, emitter, observability.Event{Level: observability.LevelError, Name: "mcp.server", Outcome: observability.OutcomeError, Error: observability.RedactedError(err)})
		os.Exit(1)
	}
	server := mcpserver.New(client, mcpserver.Options{Version: effectiveVersion, Events: emitter, ZenodoOAI: oai})
	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		observability.Emit(ctx, emitter, observability.Event{Level: observability.LevelError, Name: "mcp.server", Outcome: observability.OutcomeError, Error: observability.RedactedError(err)})
		os.Exit(1)
	}
	observability.Emit(ctx, emitter, observability.Event{Name: "mcp.server", Fields: map[string]any{"state": "stopped"}})
}
