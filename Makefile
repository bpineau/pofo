# portfodor Makefile — `make help` for the list of targets.

GO        ?= go
BINARIES  := portfodor
PKGS      := ./...
# Local staticcheck if available, otherwise a pinned version via `go run`.
STATICCHECK ?= $(shell command -v staticcheck 2>/dev/null || echo "$(GO) run honnef.co/go/tools/cmd/staticcheck@2025.1")

.DEFAULT_GOAL := build

.PHONY: build
build: ## Build the ./portfodor binary (datasets/ embedded)
	$(GO) build -o portfodor ./cmd/portfodor

.PHONY: fmt
fmt: ## Reformat all the code (gofmt -w)
	gofmt -w .

.PHONY: fmt-check
fmt-check: ## Fail if any code is not gofmt-formatted
	@out="$$(gofmt -l .)"; \
	if [ -n "$$out" ]; then \
		echo "unformatted files:"; echo "$$out"; exit 1; \
	fi

.PHONY: vet
vet: ## go vet on all packages
	$(GO) vet $(PKGS)

.PHONY: lint
lint: vet ## vet + staticcheck
	$(STATICCHECK) $(PKGS)

.PHONY: test
test: ## Unit tests + examples (no network)
	$(GO) test $(PKGS)

.PHONY: golden
golden: ## Golden tests (computations vs external references)
	$(GO) test -v ./datasets/golden/

.PHONY: cover
cover: ## Tests with coverage
	$(GO) test -cover $(PKGS)

.PHONY: check
check: fmt-check lint test ## Everything: format, lint, tests (CI target)

.PHONY: warmup
warmup: build ## Pre-fetch the cache (quotes + fees) for the catalog
	./portfodor -warmup

.PHONY: simdata
simdata: build ## (Re)generate datasets/simdata/ then re-embed it into the binary
	./portfodor -gen-simdata
	$(GO) build -o portfodor ./cmd/portfodor

.PHONY: demo
demo: build ## Demo report on the example portfolios
	./portfodor examples/*.txt

.PHONY: clean
clean: ## Remove the binaries (not data/ nor datasets/)
	rm -f $(BINARIES)

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'
