param(
    [string]$Version = "0.2.0",
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
$pluginNames = @("claude-osf", "codex-osf", "gemini-osf", "qwen-osf")

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

    $archive = Join-Path $outRoot "$pluginName-$Version-$runtime.zip"
    if (Test-Path -LiteralPath $archive) {
        Remove-Item -LiteralPath $archive -Force
    }
    Compress-Archive -Path (Join-Path $stage "*") -DestinationPath $archive -Force
    Write-Output $archive
}
