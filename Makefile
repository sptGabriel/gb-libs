.DEFAULT_GOAL := help

.PHONY: help
help: ## Print this help.
	@echo "Usage: make [target]"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {gsub(/Makefile:/, "", $$1); printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: dev-install
dev-install: ## Installs configurations (pre-commit hooks, etc) intended for when cloning the repository.
	@echo "======> Installing pre-commit-config hooks"
	@command -v python3 >/dev/null 2>&1 || (echo "python3 not installed. Aborting..." >&2; exit 1)
	@command -v pre-commit >/dev/null 2>&1 || (echo "pre-commit not installed. Aborting..." >&2; exit 1)
	@pre-commit install

.PHONY: lint
lint: ## Runs linters on the codebase.
	@echo "======> Running linters."
	@go tool golangci-lint run ./...
	@echo "======> Done linting."

.PHONY: lint-fix
lint-fix: ## Runs linters and formatters on the codebase.
	@$(MAKE) format
	@echo "======> Running linters."
	@go tool golangci-lint run ./... --fix
	@echo "======> Done linting."

.PHONY: format
format: ## Runs formatters in the codebase.
	@echo "======> Running formatters."
	@go tool golangci-lint fmt ./...
	@echo "======> Done formatting."

.PHONY: clean
clean: ## Cleans project.
	@echo "======> Running cleaning process."
	@rm -rf dist
	@rm -rf .coverage
	@go clean -testcache
	@echo "======> Done cleaning solution."

.PHONY: tidy
tidy: ## Tides the solution.
	@echo "======> Tidying solution."
	@go mod tidy
	@echo "======> Done restoring solution."

.PHONY: test-local
test-local: ## Runs the test suite with ryuk disabled. Set verbose=true to enable verbose mode.
	@echo "======> Testcontainers ryuk is disabled";
	@if [ "$(verbose)" = "true" ]; then \
		TESTCONTAINERS_RYUK_DISABLED=true $(MAKE) test verbose=true; \
	else \
		TESTCONTAINERS_RYUK_DISABLED=true $(MAKE) test; \
	fi;

.PHONY: test
test: ## Runs the test.
	@echo "======> Running Test";
	@${MAKE} clean
	@if [ "$(verbose)" = "true" ]; then \
		go test ./... -timeout=60s -v; \
	else \
		go test ./... -timeout=60s; \
	fi;