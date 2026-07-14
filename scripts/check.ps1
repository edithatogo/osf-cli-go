param(
    [switch]$AllowRaceSkip
)

$ErrorActionPreference = "Stop"

$env:GOTELEMETRY = "off"
$env:GOCACHE = Join-Path (Get-Location) ".gocache"
$env:GOMODCACHE = Join-Path (Get-Location) ".gomodcache"

go fmt ./...
go test ./...
if (Get-Command gcc -ErrorAction SilentlyContinue) {
    $env:CGO_ENABLED = "1"
    go test -race ./...
} else {
    if (-not $AllowRaceSkip) {
        throw "gcc is required for local race tests. Install gcc or rerun with -AllowRaceSkip for this Windows host. GitHub Actions still enforces race tests."
    }
    Write-Warning "Skipping local race tests because gcc is not available and -AllowRaceSkip was supplied. GitHub Actions still enforces race tests."
}
go vet ./...
go run ./tools/checkstubs
go run ./tools/checkreviews
go run ./tools/checkregistries
go run ./tools/checkfeaturematrix
go run ./tools/checkzenodoapi
go run ./tools/checkreleasecontract
go test ./... "-coverprofile=coverage.out"
go tool cover "-func=coverage.out"
