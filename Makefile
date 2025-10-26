.PHONY: build test clean install lint help

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
	go test -v ./...

lint: ## Run linters
	golangci-lint run

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	go clean

run-example: build ## Run example on RDS XRD
	./$(BINARY_NAME) generate ../argocd-k8s-resources/manifests/crossplane/overlays/prd/custom-resources/native/aws/rds/v1alpha1/rds-instance-xrd.yaml

.DEFAULT_GOAL := help
