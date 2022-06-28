SHELL := /bin/bash

.PHONY: all check format vet build test tidy

help:
	@echo "Please use \`make <target>\` where <target> is one of"
	@echo "  check               to format, vet "
	@echo "  build               to create bin directory and build"
	@echo "  generate            to generate code"
	@echo "  unit_test           to run unit test"
	@echo "  integration_test    to run integration test"

check: format vet

format:
	@echo "go fmt"
	@go fmt ./...
	@echo "ok"

vet:
	@echo "go vet"
	@go vet ./...
	@echo "ok"

build: tidy check
	@echo "build go-locale"
	@go build ./...
	@echo "ok"

test:
	@echo "run unit test"
	@go test -race -cover -coverprofile=coverage.txt -v ./...
	@go tool cover -html="coverage.txt" -o "coverage.html"
	@echo "ok"

tidy:
	@echo "Tidy and check the go mod files"
	@go mod tidy && go mod verify
	@echo "Done"
