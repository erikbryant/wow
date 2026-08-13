fmt:
	go fmt ./...

vet: fmt
	go vet ./...

vuln: vet
	govulncheck ./...

test: vuln
	go test ./...

run: test
	go run ./cmd/wow

build:
	go build -o ./bin/ ./cmd/items/
	go build -o ./bin/ ./cmd/secret/

# Targets that do not represent actual files
.PHONY: fmt vet vuln test run
