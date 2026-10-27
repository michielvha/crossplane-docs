.PHONY: build test clean install lint help check examples run-example

# Build variables
BINARY_NAME=crossplane-docs
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/michielvha/crossplane-docs/cmd.version=$(VERSION) -X github.com/michielvha/crossplane-docs/cmd.commit=$(COMMIT) -X github.com/michielvha/crossplane-docs/cmd.date=$(DATE)"

help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	go build $(LDFLAGS) -o $(BINARY_NAME) .

install: ## Install the binary
	go install $(LDFLAGS) .

test: ## Run tests
	go test -race -cover ./...

lint: ## Run linters
	golangci-lint config verify
	golangci-lint run

check: test lint ## Run tests, lint, vet, and module checks
	go vet ./...
	go mod tidy -diff
	@test -z "$$(gofmt -l .)" || { gofmt -l .; exit 1; }

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	go clean

run-example: build ## Generate docs from the bundled XRD
	./$(BINARY_NAME) xrd examples/xrd.yaml

examples: build ## Refresh the expected example output after reviewing behavior changes
	./$(BINARY_NAME) xrd examples/xrd.yaml -o examples/xrd.md
	./$(BINARY_NAME) xrd examples/xrd.yaml --show-nested=false -o examples/xrd-shallow.md
	./$(BINARY_NAME) composition examples/composition.yaml -o examples/composition.md
	./$(BINARY_NAME) composition examples/pipeline.yaml -o examples/pipeline.md

.DEFAULT_GOAL := help
