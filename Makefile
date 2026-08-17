# pofo Makefile: `make help` for the list of targets.

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

.PHONY: docker-image
docker-image: ## Build the Docker image (BuildKit; tag pofo:dev)
	DOCKER_BUILDKIT=1 docker build -f deploy/docker/Dockerfile -t pofo:dev .

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

# Every generator that reads a live source, in dependency order: the panels and
# the reference series first, then the simdata reconstructions anchored on them,
# then the offline snapshots (independent). This is the one command to run when
# the bundled data should catch up with the world; expect a few minutes of
# network. A generator that fails stops the chain, so nothing downstream is
# rebuilt on half-refreshed inputs.
.PHONY: refresh
refresh: cape broadsample macropanel euro-refdata gbond-refdata sp500-refdata trend-refdata trendnet-refdata sgtrend-refdata catbond-refdata simdata snapshots ## Refresh EVERY bundled series from its live source (network, several minutes)
	@echo "refreshed; now run 'make check', 'make golden' and 'make verify-catalog'."
	@echo "'make figure-drift' says which FIRE book plates the new data left behind; that is optional, and the book may lag."

.PHONY: simdata
simdata: build ## (Re)generate pkg/datasets/simdata/ then re-embed it into the binary
	./pofo -gen-simdata
	$(GO) build -o pofo ./cmd/pofo

.PHONY: simdata-qa
simdata-qa: build ## Reconstruction quality report: every engine vs the real quotes (network), opened in the browser
	./pofo -verify-simdata

# The data doctor over the whole bundled catalog, in each asset's native
# currency: series hygiene, plausibility against the asset class's band, and
# identity against the record (currency, share class, inception). Run it after
# any catalog edit and after `make refresh`. About a minute on a warm quote
# cache; a cold one takes as long as a `-warmup`. It exits non-zero only on
# error-grade data, so the warnings are yours to read, not to chase to zero.
.PHONY: verify-catalog
verify-catalog: build ## Data doctor over the whole bundled catalog (network), with a summary
	./pofo -verify-data

.PHONY: broadsample
broadsample: ## (Re)generate the bundled JST broad-sample panel (network) then rebuild
	$(GO) run ./cmd/gen-broadsample
	$(GO) build -o pofo ./cmd/pofo

.PHONY: cape
cape: ## (Re)generate the bundled Shiller CAPE series (network) then rebuild
	$(GO) run ./cmd/gen-cape
	$(GO) build -o pofo ./cmd/pofo

.PHONY: macropanel
macropanel: ## (Re)generate the bundled OECD monthly macro panel (network) then rebuild
	$(GO) run ./cmd/gen-macropanel
	$(GO) build -o pofo ./cmd/pofo

.PHONY: euro-refdata
euro-refdata: ## (Re)generate the bundled euro-area reference series (network) then rebuild
	$(GO) run ./cmd/gen-euro-refdata
	$(GO) build -o pofo ./cmd/pofo

.PHONY: gbond-refdata
gbond-refdata: ## (Re)generate the bundled German/Japanese/British government bond reference series (network); run `make simdata` after
	$(GO) run ./cmd/gen-gbond-refdata
	$(GO) build -o pofo ./cmd/pofo

.PHONY: sp500-refdata
sp500-refdata: ## (Re)generate the month-end SP500-USD reference series (network); run `make simdata` after
	$(GO) run ./cmd/gen-sp500-refdata
	$(GO) build -o pofo ./cmd/pofo

.PHONY: trend-refdata
trend-refdata: ## (Re)generate the monthly trend reference the managed-futures reconstructions are anchored on (network); run `make simdata` after
	$(GO) run ./cmd/gen-trend-refdata
	$(GO) build -o pofo ./cmd/pofo

.PHONY: trendnet-refdata
trendnet-refdata: ## (Re)generate the monthly NET managed-futures reference the deep trend tails are anchored on (network); run `make simdata` after
	$(GO) run ./cmd/gen-trendnet-refdata
	$(GO) build -o pofo ./cmd/pofo

.PHONY: catbond-refdata
catbond-refdata: ## (Re)generate the monthly NET insurance-linked-securities reference the cat bond reconstructions are anchored on (network); run `make simdata` after
	$(GO) run ./cmd/gen-catbond-refdata
	$(GO) build -o pofo ./cmd/pofo

.PHONY: sgtrend-refdata
sgtrend-refdata: ## (Re)generate the two daily NET managed-futures references (pure trend, all styles) the overlays and the fund donor chains use (network); run `make simdata` after
	$(GO) run ./cmd/gen-sgtrend-refdata
	$(GO) build -o pofo ./cmd/pofo

.PHONY: snapshots
snapshots: ## (Re)generate the offline fallback snapshots in pkg/marketdata/data/ (network) then rebuild
	$(GO) run ./cmd/gen-snapshots
	$(GO) build -o pofo ./cmd/pofo

.PHONY: book-drift
book-drift: build ## What the FIRE book's translations owe their French source
	./pofo -book-drift

# The book's plates freeze numbers read off the bundled datasets. Refreshing the
# data is allowed to move them, so those checks are kept out of `make check`
# (see frozenAgainstData in pkg/firebook): a data refresh never has to become a
# code change. This target is where the debt shows up, on demand.
.PHONY: figure-drift
figure-drift: ## What the FIRE book's frozen figures owe the bundled data (optional after make refresh)
	POFO_FIGURE_DRIFT=1 $(GO) test ./pkg/firebook/ -count=1

.PHONY: demo
demo: build ## Demo report on the example portfolios
	./pofo examples/*.txt

.PHONY: suggest
suggest: build ## Demo the -suggest analysis on a catalog-based example
	./pofo -suggest examples/msci-world.txt

.PHONY: clean
clean: ## Remove the binaries (not data/ nor pkg/datasets/)
	rm -f $(BINARIES)

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'
