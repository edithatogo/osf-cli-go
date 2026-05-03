# Usage

## Authentication

Set the `OSF_TOKEN` environment variable to a valid OSF personal access token:

```bash
export OSF_TOKEN="your_token_here"
```

On Windows:

```powershell
$env:OSF_TOKEN = "your_token_here"
```

Without a token, only **public** node metadata and storage files are accessible.
Commands that require authentication (`auth whoami`, `projects list`) will report a missing token error.

## Output Modes

Every command supports two output modes controlled by the `--output` flag:

- `table` (default) — human-readable aligned columns
- `json` — machine-readable JSON

The shorthand `--json` is equivalent to `--output json`.
Passing both `--json` and `--output table` is an error.

## Command Reference

### `osf auth whoami`

Show the authenticated OSF account details.

```bash
osf auth whoami
osf auth whoami --json
```

### `osf projects list`

List all project-category nodes accessible to the current user.

```bash
osf projects list
osf projects list --json
```

### `osf projects get <guid-or-url>`

Show a single project or component by its OSF GUID or URL.

```bash
osf projects get abc123
osf projects get https://osf.io/abc123
```

### `osf components list <project-guid-or-url>`

List child components of a project or component.

```bash
osf components list abc123
osf components list https://osf.io/abc123
```

### `osf files list <project-or-component-guid>`

List OSF Storage files and folders for a node.

```bash
osf files list abc123
osf files list abc123 --json
```

### `osf files download`

Download files or folder trees.

**Single file** by ID, API URL, or direct download URL:

```bash
osf files download --file abc123 ./output/
osf files download --file https://api.osf.io/v2/files/abc123/ ./output/
osf files download --file https://files.osf.io/v1/resources/abc123/providers/osfstorage/xyz ./output/
```

**Folder tree** by node GUID or URL:

```bash
osf files download --tree abc123 ./output/
osf files download --tree https://osf.io/abc123 ./output/
```

### `osf completion <bash|zsh|fish|powershell>`

Generate shell completion scripts.

```bash
osf completion bash > /etc/bash_completion.d/osf
osf completion zsh > /usr/local/share/zsh/site-functions/_osf
```

## Conflict Policy

The `--conflict` flag controls what happens when a local file already exists,
applying to both `--file` and `--tree` downloads:

| Policy      | Behavior                                                     |
|-------------|--------------------------------------------------------------|
| `fail`      | Abort the download with an error (default).                  |
| `skip`      | Leave the existing file untouched and continue.              |
| `overwrite` | Replace the existing file after the download completes.      |

```bash
osf files download --file abc123 ./out/ --conflict skip
osf files download --tree abc123 ./out/ --conflict overwrite
```

## Environment Variables

| Variable    | Required | Description                            |
|-------------|----------|----------------------------------------|
| `OSF_TOKEN` | No*      | OSF personal access token              |

*Required for authenticated commands (`auth whoami`, `projects list`).
Public node and storage file access works without it.

## Exit Codes

| Code | Meaning                         |
|------|---------------------------------|
| 0    | Success                         |
| 1    | General error or planned stub   |
| 2    | Usage or argument error         |
