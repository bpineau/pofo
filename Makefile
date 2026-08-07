# pofo Makefile — `make help` for the list of targets.

GO        ?= go
BINARIES  := pofo
PKGS      := ./...
# Local staticcheck if available, otherwise a pinned version via `go run`.
STATICCHECK ?= $(shell command -v staticcheck 2>/dev/null || echo "$(GO) run honnef.co/go/tools/cmd/staticcheck@2025.1")

.DEFAULT_GOAL := build

.PHONY: build
build: ## Build the ./pofo binary (pkg/datasets/ embedded)
	$(GO) build -o pofo ./cmd/pofo

.PHONY: install
install: ## Install the pofo binary (go install → GOBIN or GOPATH/bin)
	$(GO) install ./cmd/pofo

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
	$(GO) test -v ./pkg/datasets/golden/

.PHONY: cover
cover: ## Tests with coverage
	$(GO) test -cover $(PKGS)

.PHONY: check
check: fmt-check lint test ## Everything: format, lint, tests (CI target)

.PHONY: warmup
warmup: build ## Pre-fetch the cache (quotes + fees) for the catalog
	./pofo -warmup

.PHONY: simdata
simdata: build ## (Re)generate pkg/datasets/simdata/ then re-embed it into the binary
	./pofo -gen-simdata
	$(GO) build -o pofo ./cmd/pofo

.PHONY: broadsample
broadsample: ## (Re)generate the bundled JST broad-sample panel (network) then rebuild
	$(GO) run ./cmd/gen-broadsample
	$(GO) build -o pofo ./cmd/pofo

.PHONY: demo
demo: build ## Demo report on the example portfolios
	./pofo examples/*.txt

.PHONY: suggest
suggest: build ## Demo the -suggest analysis on a catalog-based example
	./pofo -suggest examples/world-equity.txt

.PHONY: verify
verify: build ## Run the -verify-data doctor over the bundled catalog
	./pofo -verify-data

.PHONY: clean
clean: ## Remove the binaries (not data/ nor pkg/datasets/)
	rm -f $(BINARIES)

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'
