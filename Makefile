.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: vuln
vuln:
	govulncheck ./...

.PHONY: test
test:
	go test -race ./...

.PHONY: verify
verify:
	go mod verify

.PHONY: check
check: fmt vet vuln test verify

.PHONY: build
build: check
	mkdir -p ./bin
	go build -o ./bin ./cmd/items
	go build -o ./bin ./cmd/secret
	go build -o ./bin ./cmd/wow
