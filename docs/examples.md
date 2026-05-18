# Examples

These examples use placeholder GUIDs. Replace `abc123` with the OSF project or component GUID you want to inspect.

## Inspect A Public Project

```bash
osf projects get abc123
osf components list abc123
osf files list abc123
```

## Use Authenticated Project Listing

PowerShell:

```powershell
$env:OSF_TOKEN = "your-token"
osf auth whoami
osf projects list --json
```

Bash:

```bash
export OSF_TOKEN="your-token"
osf auth whoami
osf projects list --json
```

## Download A Single File

```bash
osf files download --file file123 ./downloads/
```

Use an explicit conflict policy when rerunning a download:

```bash
osf files download --file file123 ./downloads/ --conflict skip
osf files download --file file123 ./downloads/ --conflict overwrite
```

## Download A Folder Tree

```bash
osf files download --tree abc123 ./project-files/
```

The folder-tree command validates remote paths before writing and writes a manifest for the attempted files.

## Upload And Manage Files

```bash
osf files upload --node abc123 ./analysis.csv
osf files mkdir --node abc123 data/raw
osf files addons abc123
osf files rm --node abc123 old-analysis.csv
```

Use `--yes` only when the file deletion is intentional and already reviewed:

```bash
osf files rm --node abc123 --yes old-analysis.csv
```

## Search And Export

```bash
osf search "open science"
osf preprints list --provider osf --json
osf export abc123 --json
```

## Create A Draft Registration

```bash
osf registrations create abc123 --schema schema-1 --title "Analysis plan"
```

Use `--yes` only when the draft registration inputs are already reviewed:

```bash
osf registrations create abc123 --schema schema-1 --title "Analysis plan" --yes
```

## Generate Completion

PowerShell:

```powershell
osf completion powershell
```

Bash:

```bash
osf completion bash > osf-completion.bash
```
