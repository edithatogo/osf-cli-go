# Command Reference

All commands support `--output table|json`; `--json` is shorthand for `--output json`.

## Global Commands

### `osf --help`

Print CLI help.

### `osf --version`

Print the CLI version and build metadata when available.

### `osf --output json`

Print the machine-readable command contract.

## Authentication

### `osf auth whoami`

Show the authenticated OSF account.

Requires `OSF_TOKEN` or `OSF_USERNAME` plus `OSF_PASSWORD`.

Examples:

```bash
osf auth whoami
osf auth whoami --json
```

### `osf auth login`

Print guided token-bootstrap instructions. OSF personal access tokens are still created in OSF Account Settings unless OSF later documents a supported token creation API.

Options:

- `--token <token>`, acknowledge a generated PAT as supplied
- `--print-env`, with `--token`, print a shell export command for `OSF_TOKEN`

Examples:

```bash
osf auth login
osf auth login --token "$OSF_TOKEN" --print-env
```

### `osf whoami`

Alias for `osf auth whoami`.

## Projects And Components

### `osf projects list`

List project-category nodes accessible to the authenticated user.

Requires `OSF_TOKEN` or `OSF_USERNAME` plus `OSF_PASSWORD`.

### `osf projects get <guid-or-url>`

Show one OSF project or component by GUID or OSF URL.

Public projects can be read without `OSF_TOKEN` when OSF permits public access.

### `osf projects create --title <title>`

Create an OSF project or component. Requires typed confirmation unless `--yes` is supplied.

Options:

- `--title <title>`, required node title
- `--category <category>`, default `project`
- `--description <text>`, optional node description
- `--yes`, skip the typed confirmation prompt

### `osf projects update <guid-or-url>`

Update an OSF node title or description. Omitted fields are read from the current node and preserved. Requires typed confirmation unless `--yes` is supplied.

Options:

- `--title <title>`, new node title
- `--description <text>`, new node description
- `--yes`, skip the typed confirmation prompt

### `osf projects delete <guid-or-url>`

Delete an OSF node. Requires typed confirmation unless `--yes` is supplied.

Aliases:

- `osf projects rm <guid-or-url>`

Options:

- `--yes`, skip the typed confirmation prompt

### `osf components list <project-guid-or-url>`

List child components for a project or component.

## Files

### `osf files list <project-or-component-guid> [folder-id-or-path]`

List files and folders from the node's OSF Storage provider.

An optional OSF Storage folder ID or API path segment lists the contents of a nested folder.

JSON output includes an optional `md5` checksum from OSF's
`attributes.extra.hashes.md5` metadata. No file content is downloaded to
calculate it.

### `osf files download --file <file-id-or-url> <destination>`

Download one file by OSF file ID, OSF file API URL, or direct WaterButler download URL.

Options:

- `--conflict fail|skip|overwrite`, default `fail`

### `osf files download --tree <node-guid-or-url> <destination>`

Download a node's OSF Storage folder tree and write a manifest.

Options:

- `--conflict fail|skip|overwrite`, default `fail`

### `osf files upload --node <guid> <local-path>`

Upload a local file to a node's OSF Storage provider.

Options:

- `--conflict fail|overwrite`, default `fail`

### `osf files mkdir --node <guid> <folder-name>`

Create a folder in a node's OSF Storage provider. Nested paths such as `data/raw` are supported.

### `osf files rm --node <guid> <file-name>`

Delete a file from a node's OSF Storage provider.

Options:

- `--yes`, skip the typed confirmation prompt

### `osf files addons <node-id>`

List configured storage add-ons for a node.

## Search And Preprints

### `osf search <query>`

Search OSF node content and print matching project/component-like results.

Options:

- `--limit <n>`, maximum records to return, default `20`; use `0` to follow all pages

### `osf preprints list`

List OSF preprints visible to the current request context.

Options:

- `--provider <provider>`, filter by preprint provider
- `--limit <n>`, maximum records to return, default `20`; use `0` to follow all pages

## Registrations

### `osf registrations create <node-id>`

Create a draft registration for an existing node.

Options:

- `--schema <schema-id>`, required registration schema ID
- `--title <title>`, optional draft title
- `--description <text>`, optional draft description
- `--yes`, skip the typed confirmation prompt

## Export

### `osf export <guid-or-url>`

Export a node snapshot containing metadata, contributors, child components, and OSF Storage file listings.

Examples:

```bash
osf export abc123
osf export https://osf.io/abc123 --json
```

## Browser And Shell Integration

### `osf open <guid-or-url>`

Open an OSF node in the system browser.

### `osf completion <bash|zsh|fish|powershell>`

Generate a shell completion script.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Runtime or API error |
| 2 | Usage or argument error |
