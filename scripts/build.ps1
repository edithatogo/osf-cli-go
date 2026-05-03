param(
    [string]$Output = "bin\osf.exe",
    [string]$Version = "",
    [string]$Commit = "",
    [string]$BuildDate = ""
)

$ErrorActionPreference = "Stop"

function Get-GitValue {
    param(
        [string[]]$Arguments,
        [string]$Fallback
    )

    try {
        $value = & git @Arguments 2>$null
        if ($LASTEXITCODE -eq 0 -and $value) {
            return $value.Trim()
        }
    } catch {
    }

    return $Fallback
}

if (-not $Version) {
    $Version = Get-GitValue -Arguments @("describe", "--tags", "--always", "--dirty") -Fallback "0.0.0-dev"
}

if (-not $Commit) {
    $Commit = Get-GitValue -Arguments @("rev-parse", "--short", "HEAD") -Fallback "unknown"
}

if (-not $BuildDate) {
    $BuildDate = (Get-Date).ToUniversalTime().ToString("o")
}

$outputDir = Split-Path -Parent $Output
if ($outputDir) {
    New-Item -ItemType Directory -Force -Path $outputDir | Out-Null
}

$ldflags = @(
    "-s -w",
    "-X github.com/edithatogo/osf-cli-go/internal/cli.version=$Version",
    "-X github.com/edithatogo/osf-cli-go/internal/cli.buildCommit=$Commit",
    "-X github.com/edithatogo/osf-cli-go/internal/cli.buildDate=$BuildDate"
) -join " "

& go build -trimpath -ldflags $ldflags -o $Output ./cmd/osf

