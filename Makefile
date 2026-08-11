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

# Targets that do not represent actual files
.PHONY: fmt vet vuln test run
