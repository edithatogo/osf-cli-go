# Release Checklist Evidence

Date: 2026-07-06

Scope: MCP registry/plugin distribution track release-readiness check. This was
a local validation pass only. No git tag was created, no GitHub Release was
published, and no registry package was pushed by this pass.

## Versioning

- `go run ./cmd/osf --version`: pass
- Reported version: `0.0.0-dev`
- Boundary: this is an untagged local checkout. A real release must inject the
  semver tag through the existing release build metadata before publication.

## CLI And Completion Checks

- `go run ./cmd/osf --help`: pass, output captured locally
- `go run ./cmd/osf completion bash`: pass, output captured locally
- `go run ./cmd/osf completion powershell`: pass, output captured locally

Captured output sizes:

| Output | Bytes |
| --- | ---: |
| `osf --help` | 1169 |
| `osf --version` | 10 |
| `osf completion bash` | 15973 |
| `osf completion powershell` | 10736 |

## Binary And Checksum Smoke

Temporary output directory: `/tmp/osf-release-check.DreaJF`

Commands:

```sh
go build -trimpath -o /tmp/osf-release-check.DreaJF/osf ./cmd/osf
go build -trimpath -o /tmp/osf-release-check.DreaJF/osf-mcp ./cmd/osf-mcp
shasum -a 256 /tmp/osf-release-check.DreaJF/osf /tmp/osf-release-check.DreaJF/osf-mcp
```

Results:

| Artifact | Bytes | SHA-256 |
| --- | ---: | --- |
| `osf` | 12277010 | `0f3830dcc56672a777ed60fbe3f303b0d14ab5c410c11db646f4461a49722f48` |
| `osf-mcp` | 11382258 | `220441e6c52a7ca7df8d65ecd869f8cada877572ba9e814f0f15fb5b5c478916` |

Boundary: this local pass built macOS binaries only. Cross-platform release
artifacts remain covered by the GitHub Actions release/MCPB/plugin workflows.

## Quality Gates

- `go test ./...`: pass
- `go vet ./...`: pass
- `go run ./tools/checkstubs`: pass
- `go run ./tools/checkregistries`: pass
- `go run ./tools/checkreviews`: pass

## Safety Review

- Live OSF validation remains opt-in and was not run in this release checklist
  pass.
- No OSF credentials were required or read.
- Smithery and MCP.Directory statuses are recorded in
  `registry/directory-submissions.json`; publication receipts remain evidence,
  not a new publication action from this pass.
