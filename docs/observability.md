# Structured observability

The CLI, API client, and MCP server can emit versioned JSON Lines events when
`OSF_EVENT_LOG` is set. Observability is disabled by default and never writes
events to command stdout, so `--output json` remains machine-readable.

## Configuration

```sh
OSF_EVENT_LOG=/tmp/osf-events.jsonl osf projects list --output json
OSF_EVENT_LEVEL=debug OSF_EVENT_LOG=/tmp/osf-events.jsonl osf files list <node>
```

`OSF_EVENT_LOG` accepts a file path or `stderr`. File destinations are created
with owner-only permissions and appended to; `stdout` is rejected. The event
level defaults to `info` and supports `debug`, `info`, `warn`, and `error`.

Each event uses the `osf.event.v1` schema and includes timestamp, level, event
name, operation ID, request ID, duration, retry count, outcome, endpoint class,
and a classified error when applicable. API endpoint URLs, authorization
headers, tokens, passwords, credentials, and local paths are never retained in
event fields. Paths are represented as `[REDACTED_PATH]`.

## Destinations and retention

The project does not collect or transmit telemetry. Operators own the local
log destination and retention policy. Use a rotating system service or remove
the file after troubleshooting; do not commit event logs or attach them to
public issues without inspecting redaction.

## Troubleshooting and compatibility

Use operation and request IDs to correlate a CLI command with API request
events. `outcome` is `ok`, `error`, or `canceled`; error classes are stable
categories such as `authentication`, `authorization`, `network`, `timeout`,
`validation`, `decode`, and `internal`. New fields may be added, but existing
field meanings and the schema version remain stable for the 1.0 contract.
