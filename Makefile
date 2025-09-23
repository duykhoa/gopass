# Makefile for GoPass UI MVP

APP_NAME=gopass-ui
GO_FILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")
GO_VERSION ?= 1.25
GOOS ?= linux
default: run

.PHONY: run
run:
	go run cmd/main.go

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	GOOS=$(GOOS) GOARCH=amd64 go build -ldflags="-s -w" -o bin/$(APP_NAME)-$(GOOS)-amd64 cmd/main.go

.PHONY: build-all
build-all:
	@for os in linux darwin windows; do \
		GOOS=$$os GOARCH=amd64 go build -ldflags="-s -w" -o bin/$(APP_NAME)-$$os-amd64 cmd/main.go; \
	done

.PHONY: clean
clean:
	rm -rf bin/

integration-test:
	go test -tags=integration ./internal/service/...
