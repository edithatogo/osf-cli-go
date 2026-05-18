# Architecture

## Package Layout

```
cmd/osf/            — main package, CLI entry point
internal/
  auth/             — credential loading, auth-mode selection, redaction of secrets in output
  cli/              — Cobra command tree, command contracts, terminal behavior
  download/         — safe file writes, conflict policies, folder-tree plans
  osfapi/           — OSF API v2 HTTP client (JSON:API)
  output/           — table and JSON output helpers
docs/               — user-facing documentation
```

## Command Routing

Commands are defined with [Cobra](https://github.com/spf13/cobra) in
`internal/cli/`. The entrypoint in `cmd/osf/main.go` calls `cli.Run()`,
which creates the root command, attaches subcommands, and executes it.

The root command (`osf`) exposes these command groups:

- `auth` → `whoami`, `login`
- `projects` → `list`, `get`
- `components` → `list`
- `files` → `list`, `download`, `upload`, `mkdir`, `rm`
- `export`
- `search`
- `preprints` → `list`
- `registrations` → `create`
- `open`
- `whoami`
- `completion` → `bash`, `zsh`, `fish`, `powershell`

The root command supports `--output json` to emit a machine-readable
contract describing the available commands and exit codes.

## Readonly Client Interface

All command handlers depend on the `readonlyClient` interface defined in
`internal/cli/readonly.go`:

```go
type readonlyClient interface {
    CurrentUser(context.Context) (osfapi.User, error)
    ListProjects(context.Context) ([]osfapi.Node, error)
    GetNode(context.Context, string) (osfapi.Node, error)
    ListNodeContributors(context.Context, string) ([]osfapi.Contributor, error)
    ListNodeChildren(context.Context, string) ([]osfapi.Node, error)
    ListStorageFiles(context.Context, string, ...string) ([]osfapi.StorageFile, error)
    GetStorageFile(context.Context, string) (osfapi.StorageFile, error)
    OpenDownload(context.Context, string) (io.ReadCloser, error)
    GetNodeFilesProvider(context.Context, string) (string, error)
    UploadFile(context.Context, string, string, io.Reader, string) error
    CreateFolder(context.Context, string, string) error
    DeleteFile(context.Context, string, string) error
    ListPreprints(context.Context) ([]osfapi.Node, error)
    SearchOSF(context.Context, string) ([]osfapi.Node, error)
    ListNodeAddons(context.Context, string) ([]osfapi.Node, error)
}
```

This interface serves two purposes:

1. **Testability** — Tests provide a fake implementation that returns
   fixture data, avoiding real HTTP calls.
2. **Decoupling** — Commands don't depend on `osfapi.Client` directly;
   auth checks (e.g. requiring a token for `ListProjects`) happen in
   the `defaultReadonlyClient` wrapper, not in the command code.

### `defaultReadonlyClient`

`newDefaultReadonlyClient()` constructs a production client that:

1. Loads credentials from `os.LookupEnv` (via `auth.LoadCredentials`).
2. Selects `OSF_TOKEN` first, then `OSF_USERNAME` plus `OSF_PASSWORD`, otherwise anonymous public access.
3. Creates an `osfapi.Client` with the selected credentials.
4. Wraps it in a `defaultReadonlyClient` that guards authenticated
   methods (`CurrentUser`, `ListProjects`) by returning
   `MissingTokenError` when no credentials were loaded.

## OSF API Client

`internal/osfapi/client.go` implements the HTTP layer over the
[OSF API v2](https://developer.osf.io/) JSON:API specification.

Key design points:

- **Functional options** (`WithHTTPClient`, `WithBearerToken`,
  `WithCredentials`, `WithUsernamePassword`) for
  configuration without breaking the constructor signature.
- **Default request timeout** of 30 seconds for production HTTP clients, so
  slow OSF/public endpoints fail instead of hanging indefinitely.
- **Automatic pagination** via `collectPages`, which follows the
  `links.next` cursor in each JSON:API response until exhaustion.
- **Error parsing** via `parseAPIError`, which handles both the
  JSON:API `errors[]` array and top-level `title`/`detail` fields.
- **Relative URL resolution** via `resolveReference`, so pagination
  links (often relative) are resolved against the base URL.
- WaterButler write methods (`UploadFile`, `CreateFolder`, `DeleteFile`)
  encode remote path segments, preserve context cancellation, and return
  actionable response-body errors for non-2xx responses.
- The low-level HTTP helpers are unexported; exported methods such as
  `GetNode`, `ListStorageFiles`, and `UploadFile` are typed wrappers around them.

### Testing

The CI test suite uses `httptest.NewServer` with fixture JSON files
under `internal/osfapi/testdata`. CLI tests use a fake `readonlyClient`
so commands can be exercised without network calls.

### Types

`internal/osfapi/types.go` defines the JSON:API resource objects and
their embedded attributes structs:

- `User` / `UserAttributes` — the `/v2/users/me/` response
- `Node` / `NodeAttributes` — projects and components
- `Contributor` / `ContributorAttributes` — contributor entries
- `StorageFile` / `StorageFileAttributes` — OSF Storage files and folders
- `FileVersion` / `FileVersionAttributes` — file-version metadata
- `Links` — the standard JSON:API links object (self, next, prev, download)
- `APIError` — structured error with status code, method, path, title, detail

## Download Subsystem

`internal/download/` provides safe file download with three concerns:

### Path Protection

`path.go` prevents directory traversal:

1. `NormalizeDestination` — resolves the root directory to an absolute path.
2. `NormalizeRemotePath` — cleans remote paths, rejecting `..`.
3. `ResolveDestination` — joins root + remote and validates with two checks:
   - **Lexical check** (`withinBaseLexical`): the resolved path must be
     under the root via `filepath.Rel`.
   - **Symlink check** (`withinBase`): walks up from the resolved path,
     evaluating symlinks on existing components, then rechecks lexical
     containment. This prevents symlink-based escapes.

### Atomic Writes

`WriteStreamAtomically` in `write.go`:

1. Creates the parent directory (`os.MkdirAll` with 0755).
2. Checks if the destination exists and applies the conflict policy.
3. Writes to a temporary file (`.dst.XXXXXX.tmp`) in the same directory.
4. Sets file permissions.
5. On `ConflictOverwrite`, removes the existing destination first.
6. Renames (`os.Rename`) the temp file to the final path — this is an
   atomic operation on most platforms, ensuring no partial files.

### Conflict Policy

`conflict.go` defines three policies:

| Policy      | Behavior                              |
|-------------|---------------------------------------|
| `fail`      | Return an error if destination exists |
| `skip`      | Return written=false, continue        |
| `overwrite` | Remove existing, then rename          |

### Folder-Tree Downloads

`folder.go` builds on the single-file writer:

1. `NewFolderDownloadPlan` validates all entries (path safety, reader
   availability) before any I/O begins.
2. `Execute` iterates planned files, opening readers and calling
   `WriteStreamAtomically` for each.
3. On the first write error, execution halts and a partial manifest
   is returned alongside the error.
4. `FolderDownloadManifest` records every entry's remote path, local
   path, byte count, status, and any error message.

## Authentication

`internal/auth/auth.go` handles credential loading and redaction:

- **Token loading** (`LoadToken`): reads `OSF_TOKEN` from any `Source`
  (default: `EnvSource`, backed by `os.LookupEnv`). Returns
  `MissingTokenError` when unset or empty.
- **Credential loading** (`LoadCredentials`): prefers `OSF_TOKEN`, then falls
  back to `OSF_USERNAME` plus `OSF_PASSWORD`; missing or partial fallback
  credentials are represented by `MissingCredentialsError`.
- **Source abstraction**: `Source` is an interface with a single
  `Lookup` method. `FuncSource` adapts a function; `EnvSource` reads
  from the process environment. This makes token loading testable.
- **Redaction** (`Redact`): scrubs three patterns from loggable text:
  - Explicitly provided secret strings (tokens, passwords).
  - `Bearer <token>` patterns in HTTP headers.
  - `Basic <base64>` patterns in HTTP headers.
  - `OSF_TOKEN=<value>`, `OSF_USERNAME=<value>`, and
    `OSF_PASSWORD=<value>` patterns in shell commands.
  - Any 24+ character base64-like string (OSF token heuristic).
- **RedactError** (`RedactError`): wraps `Redact` for error values,
  returning nil for nil errors.

## Output

`internal/output/output.go` provides two formatters:

- `WriteJSON` — encodes any value as compact JSON (no HTML escaping)
  with a trailing newline.
- `WriteTable` — uses `text/tabwriter` to produce aligned,
  tab-separated columns with optional header row.
