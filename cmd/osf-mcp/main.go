package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/edithatogo/osf-cli-go/internal/auth"
	"github.com/edithatogo/osf-cli-go/internal/mcpserver"
	"github.com/edithatogo/osf-cli-go/internal/osfapi"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var version = "0.0.0-dev"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil)).With(
		"service", "osf-mcp",
		"version", version,
	)
	credentials, _ := auth.LoadCredentials(auth.EnvSource{})
	client, err := osfapi.New("", osfapi.WithCredentials(credentials))
	if err != nil {
		logger.Error("initialize OSF client", "error", auth.RedactError(err))
		os.Exit(1)
	}

	server := mcpserver.New(client, mcpserver.Options{Version: version})
	logger.Info("starting MCP server", "transport", "stdio")
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		logger.Error("MCP server stopped", "error", auth.RedactError(err))
		os.Exit(1)
	}
	logger.Info("MCP server stopped")
}
