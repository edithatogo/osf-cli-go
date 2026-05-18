package main

import (
	"context"
	"log"

	"github.com/edithatogo/osf-cli-go/internal/auth"
	"github.com/edithatogo/osf-cli-go/internal/mcpserver"
	"github.com/edithatogo/osf-cli-go/internal/osfapi"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var version = "0.0.0-dev"

func main() {
	credentials, _ := auth.LoadCredentials(auth.EnvSource{})
	client, err := osfapi.New("", osfapi.WithCredentials(credentials))
	if err != nil {
		log.Fatal(err)
	}

	server := mcpserver.New(client, mcpserver.Options{Version: version})
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
