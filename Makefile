.PHONY: build
build:
	go build -ldflags "-X main.date=$$(date -u +%Y-%m-%dT%H:%M:%SZ) -X main.commit=$$(git rev-parse HEAD) -X main.version=<unreleased>" ./cmd/co
