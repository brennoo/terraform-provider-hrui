# Build the provider
build: fmt
	go build -o terraform-provider-hrui

.PHONY: docs
docs:
	tfplugindocs generate

# Run tests
test:
	go test ./...

# Format the code
fmt:
	@if ! command -v gofumpt &> /dev/null; then \
		echo "gofumpt not found. Please install it with 'go install mvdan.cc/gofumpt@latest'"; \
		exit 1; \
	fi
	gofumpt -l -w .
	terraform fmt -recursive examples

lint:
	golangci-lint  run ./...


# Run all checks
all: fmt lint test
