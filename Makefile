.PHONY: fmt test race vet stubs reviews coverage check

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

check: fmt test race vet stubs reviews coverage
