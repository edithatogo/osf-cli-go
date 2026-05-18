param(
    [string]$Version = "0.2.0",
    [string]$OutDir = "dist/mcpb",
    [switch]$UseMcpbCli
)

$ErrorActionPreference = "Stop"

$repo = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$outRoot = [System.IO.Path]::GetFullPath((Join-Path $repo $OutDir))
if (-not $outRoot.StartsWith($repo, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to write outside repository: $outRoot"
}

$runtime = [System.Runtime.InteropServices.RuntimeInformation]::RuntimeIdentifier
$runningOnWindows = $runtime -like "win*"
$binaryName = if ($runningOnWindows) { "osf-mcp.exe" } else { "osf-mcp" }
$bundleName = "osf-cli-go-$Version-$runtime"
$bundleDir = Join-Path $outRoot $bundleName
$serverDir = Join-Path $bundleDir "server"

New-Item -ItemType Directory -Force -Path $serverDir | Out-Null
Copy-Item -LiteralPath (Join-Path $repo "packaging/mcpb/manifest.json") -Destination (Join-Path $bundleDir "manifest.json") -Force
Copy-Item -LiteralPath (Join-Path $repo "packaging/mcpb/.mcpbignore") -Destination (Join-Path $bundleDir ".mcpbignore") -Force

$binaryPath = Join-Path $serverDir $binaryName
& go build -trimpath -ldflags "-s -w -X main.version=$Version" -o $binaryPath ./cmd/osf-mcp

$mcpbPath = Join-Path $outRoot "$bundleName.mcpb"
if (Test-Path -LiteralPath $mcpbPath) {
    Remove-Item -LiteralPath $mcpbPath -Force
}

if ($UseMcpbCli) {
    & mcpb validate $bundleDir
    & mcpb pack $bundleDir $mcpbPath
} else {
    Compress-Archive -Path (Join-Path $bundleDir "*") -DestinationPath $mcpbPath -Force
}

Write-Output $mcpbPath
