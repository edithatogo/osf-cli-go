# OSF CLI Go MCP server

Use this repository as a read-only MCP server for Open Science Framework
projects, components, files, contributors, registrations, preprints, and
search results.

## Install

The supported local-development command is:

```text
go run ./cmd/osf-mcp
```

The repository requires Go 1.24 or newer. For a released binary, install the
versioned command instead:

```text
go install github.com/edithatogo/osf-cli-go/cmd/osf-mcp@v0.3.2
```

## Configure

Set `OSF_TOKEN` in the environment before starting the server. The server also
accepts `OSF_USERNAME` and `OSF_PASSWORD` as a compatibility fallback. Never
place credentials in an MCP configuration file or commit them to the
repository.

Example Cline configuration:

```json
{
  "mcpServers": {
    "osf": {
      "command": "go",
      "args": ["run", "./cmd/osf-mcp"],
      "env": {
        "OSF_TOKEN": "${env:OSF_TOKEN}",
        "OSF_USERNAME": "${env:OSF_USERNAME}",
        "OSF_PASSWORD": "${env:OSF_PASSWORD}"
      }
    }
  }
}
```

The server communicates over stdio. Logs are written to stderr and sensitive
values are redacted. The MCP tools are read-only and include explicit
authentication and failure responses.

## Verify

From the repository root, run:

```text
go run ./tools/checkreleasecontract
go test ./...
```

Do not claim that a Cline Marketplace listing is approved until the official
marketplace repository shows the submission as accepted.
