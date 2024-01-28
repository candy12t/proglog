SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := help

.PHONY: help
help: ## Prints help for targets with comments
	@cat $(MAKEFILE_LIST) | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: run
run: ## Launch go server
	go run cmd/proglog/main.go

.PHONY: test
test: ## Run go test
	go test ./... -race -count=1 $(option)

.PHONY: protobuf-compile
protobuf-compile: ## Compile protobuf
	protoc api/v1/*.proto \
		--go_out=. \
		--go_opt=paths=source_relative \
		--proto_path=.

.PHONY: setup
setup: ## Install tools for development
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
