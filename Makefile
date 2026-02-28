.PHONY: build run test docs docs-check

DOCGEN=go run github.com/princjef/gomarkdoc/cmd/gomarkdoc@v1.1.0
VERSION?=$(shell cat VERSION)
LDFLAGS=-X main.version=v$(VERSION)


build:
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/yanzi ./cmd/yanzi

run:
	go run -ldflags "$(LDFLAGS)" ./cmd/yanzi $(ARGS)

test:
	go test ./...

docs:
	$(DOCGEN) -o docs/API.md ./cmd/yanzi ./internal/...

docs-check:
	$(DOCGEN) --check -o docs/API.md ./cmd/yanzi ./internal/...
