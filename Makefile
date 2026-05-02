.PHONY: fmt test race vet stubs reviews coverage build check lint vuln

fmt:
	go fmt ./...

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

stubs:
	go run ./tools/checkstubs

reviews:
	go run ./tools/checkreviews

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

build:
	pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/build.ps1

lint:
	golangci-lint run

vuln:
	govulncheck ./...

check: fmt test race vet stubs reviews coverage
