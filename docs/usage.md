# Usage

## Authentication

Set the `OSF_TOKEN` environment variable to a valid OSF personal access token.
This is the preferred mode for automation:

```bash
export OSF_TOKEN="your_token_here"
```

On Windows:

```powershell
$env:OSF_TOKEN = "your_token_here"
```

Without a token, only **public** node metadata and storage files are accessible.
Commands that require authentication (`auth whoami`, `projects list`) will report a missing token error unless username/password fallback credentials are present.

Username/password fallback is available through environment variables:

```bash
export OSF_USERNAME="you@example.org"
export OSF_PASSWORD="your_password"
```

On Windows:

```powershell
$env:OSF_USERNAME = "you@example.org"
$env:OSF_PASSWORD = "your_password"
```

`OSF_TOKEN` always takes precedence when both credential sets are present.
Username/password support may not work for accounts that require SSO or two-factor authentication. Use `osf auth login` for guided token setup; OSF tokens are created in Account Settings.

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

### `osf auth login`

Print guided token-bootstrap instructions. If OSF later exposes a supported token creation API, this command is the place where that flow should be added.

```bash
osf auth login
osf auth login --token "$OSF_TOKEN"
osf auth login --token "$OSF_TOKEN" --print-env
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

### `osf projects create --title <title>`

Create an OSF project or component. The command prompts for typed confirmation unless `--yes` is supplied.

```bash
osf projects create --title "Replication package" --description "Analysis files"
osf projects create --title "Replication package" --yes
```

### `osf projects update <guid-or-url>`

Update an OSF node title or description. The command preserves omitted fields and prompts for typed confirmation unless `--yes` is supplied.

```bash
osf projects update abc123 --title "Updated title"
osf projects update https://osf.io/abc123 --description "Updated description" --yes
```

### `osf projects delete <guid-or-url>`

Delete an OSF node. The command prompts for typed confirmation unless `--yes` is supplied.

```bash
osf projects delete abc123
osf projects rm https://osf.io/abc123 --yes
```

### `osf components list <project-guid-or-url>`

List child components of a project or component.

```bash
osf components list abc123
osf components list https://osf.io/abc123
```

### `osf files list <project-or-component-guid> [folder-id-or-path]`

List OSF Storage files and folders for a node.

```bash
osf files list abc123
osf files list abc123 <folder-id>
osf files list abc123 --json
```

JSON listings include an optional `md5` value when OSF provides
`attributes.extra.hashes.md5`; the CLI does not download files to calculate a
checksum.

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

### `osf files upload --node <guid> <local-path>`

Upload a file to a node's OSF Storage.

```bash
osf files upload --node abc123 ./report.pdf
osf files upload --node abc123 ./data.csv --conflict overwrite
```

The `--conflict` flag accepts `fail` (default) or `overwrite`.

### `osf files mkdir --node <guid> <folder-name>`

Create a folder in a node's OSF Storage.

```bash
osf files mkdir --node abc123 "My Folder"
osf files mkdir --node abc123 data/raw
```

### `osf files rm --node <guid> <file-name>`

Delete a file from a node's OSF Storage.

```bash
osf files rm --node abc123 old-data.csv
osf files rm --node abc123 --yes old-data.csv
```

### `osf files addons <node-id>`

List configured storage add-ons for a node.

```bash
osf files addons abc123
osf files addons abc123 --json
```

### `osf search <query>`

Search OSF projects and components by text query.

```bash
osf search "open science"
osf search "reproducibility" --json
osf search "medical industrial action" --limit 10
```

### `osf preprints list`

List all preprints available on OSF.

```bash
osf preprints list
osf preprints list --json
osf preprints list --provider osf
osf preprints list --limit 10
```

### `osf registrations create <node-id>`

Create a draft registration for an existing node.

```bash
osf registrations create abc123 --schema schema-1 --title "Analysis plan"
osf registrations create abc123 --schema schema-1 --title "Analysis plan" --yes
```

### `osf open <guid-or-url>`

Open an OSF node in the default web browser.

```bash
osf open abc123
osf open https://osf.io/abc123
```

### `osf whoami`

Show the active authenticated OSF account (alias for `auth whoami`).

```bash
osf whoami
osf whoami --json
```

### `osf export <guid-or-url>`

Export a full node snapshot including metadata, contributors, files, and components.

```bash
osf export abc123
osf export abc123 --json
osf export https://osf.io/abc123 --json
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

| Variable       | Required | Description                            |
|----------------|----------|----------------------------------------|
| `OSF_TOKEN`    | No*      | OSF personal access token              |
| `OSF_USERNAME` | No       | OSF username/email fallback            |
| `OSF_PASSWORD` | No       | OSF password fallback                  |

*Required for authenticated commands (`auth whoami`, `projects list`) unless both `OSF_USERNAME` and `OSF_PASSWORD` are set and OSF accepts password-based auth for the account.
Public node and storage file access works without it.

## Exit Codes

| Code | Meaning                         |
|------|---------------------------------|
| 0    | Success                         |
| 1    | General error or planned stub   |
| 2    | Usage or argument error         |
