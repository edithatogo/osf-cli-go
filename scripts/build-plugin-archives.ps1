param(
    [string]$Version = "0.3.2",
    [string]$OutDir = "dist/plugins"
)

$ErrorActionPreference = "Stop"

$repo = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$outRoot = [System.IO.Path]::GetFullPath((Join-Path $repo $OutDir))
if (-not $outRoot.StartsWith($repo, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to write outside repository: $outRoot"
}

$runtime = [System.Runtime.InteropServices.RuntimeInformation]::RuntimeIdentifier
$binaryName = if ($runtime -like "win*") { "osf-mcp.exe" } else { "osf-mcp" }
$pluginNames = @("github-copilot-osf", "claude-osf", "codex-osf", "gemini-osf", "qwen-osf")

New-Item -ItemType Directory -Force -Path $outRoot | Out-Null

foreach ($pluginName in $pluginNames) {
    $src = Join-Path $repo "plugins/$pluginName"
    $stage = Join-Path $outRoot "$pluginName-$Version-$runtime"
    $bin = Join-Path $stage "bin"
    if (Test-Path -LiteralPath $stage) {
        Remove-Item -LiteralPath $stage -Recurse -Force
    }
    New-Item -ItemType Directory -Force -Path $bin | Out-Null
    Copy-Item -Path (Join-Path $src "*") -Destination $stage -Recurse -Force
    & go build -trimpath -ldflags "-s -w -X main.version=$Version" -o (Join-Path $bin $binaryName) ./cmd/osf-mcp
    if ($LASTEXITCODE -ne 0) {
        throw "go build failed for $pluginName"
    }

    $archive = Join-Path $outRoot "$pluginName-$Version-$runtime.zip"
    if (Test-Path -LiteralPath $archive) {
        Remove-Item -LiteralPath $archive -Force
    }
    # Compress-Archive skips dotfiles, which would silently omit plugin manifests
    # and MCP configuration. Use the .NET implementation so hidden files remain
    # part of the distributable archive.
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    [System.IO.Compression.ZipFile]::CreateFromDirectory($stage, $archive)
    $manifestPath = switch ($pluginName) {
        "claude-osf" { ".claude-plugin/plugin.json" }
        "codex-osf" { ".codex-plugin/plugin.json" }
        "github-copilot-osf" { "plugin.json" }
        "gemini-osf" { "gemini-extension.json" }
        "qwen-osf" { "qwen-extension.json" }
    }
    $documentationPath = switch ($pluginName) {
        "gemini-osf" { "GEMINI.md" }
        "qwen-osf" { "QWEN.md" }
        default { "README.md" }
    }
    $mcpPath = switch ($pluginName) {
        "gemini-osf" { "gemini-extension.json" }
        "qwen-osf" { "qwen-extension.json" }
        default { ".mcp.json" }
    }
    $requiredEntries = @($manifestPath, $mcpPath, $documentationPath, "bin/$binaryName") | Select-Object -Unique
    $zip = [System.IO.Compression.ZipFile]::OpenRead($archive)
    try {
        foreach ($requiredEntry in $requiredEntries) {
            if ($null -eq $zip.GetEntry($requiredEntry)) {
                throw "Archive $archive is missing required entry $requiredEntry"
            }
        }
    }
    finally {
        $zip.Dispose()
    }
    Write-Output $archive
}
