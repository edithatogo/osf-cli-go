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
    Write-Warning "Skipping local race tests because gcc is not available. GitHub Actions still enforces race tests."
}
go vet ./...
go run ./tools/checkstubs
go test ./... "-coverprofile=coverage.out"
go tool cover "-func=coverage.out"
