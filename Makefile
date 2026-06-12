# Makefile de portfodor — `make help` pour la liste des cibles.

GO        ?= go
BINARIES  := portfodor simgen-bin
PKGS      := ./...
# staticcheck local s'il existe, sinon version épinglée via `go run`.
STATICCHECK ?= $(shell command -v staticcheck 2>/dev/null || echo "$(GO) run honnef.co/go/tools/cmd/staticcheck@2025.1")

.DEFAULT_GOAL := build

.PHONY: build
build: ## Compile les binaires (./portfodor et ./simgen-bin)
	$(GO) build -o portfodor ./cmd/portfodor
	$(GO) build -o simgen-bin ./cmd/simgen

.PHONY: fmt
fmt: ## Reformate tout le code (gofmt -w)
	gofmt -w .

.PHONY: fmt-check
fmt-check: ## Échoue si du code n'est pas au format gofmt
	@out="$$(gofmt -l .)"; \
	if [ -n "$$out" ]; then \
		echo "fichiers non formatés:"; echo "$$out"; exit 1; \
	fi

.PHONY: vet
vet: ## go vet sur tous les paquets
	$(GO) vet $(PKGS)

.PHONY: lint
lint: vet ## vet + staticcheck
	$(STATICCHECK) $(PKGS)

.PHONY: test
test: ## Tests unitaires + exemples (hors réseau)
	$(GO) test $(PKGS)

.PHONY: golden
golden: ## Tests étalon (calculs vs références externes)
	$(GO) test -v ./golden/

.PHONY: cover
cover: ## Tests avec couverture
	$(GO) test -cover $(PKGS)

.PHONY: check
check: fmt-check lint test ## Tout: format, lint, tests (cible CI)

.PHONY: warmup
warmup: build ## Précharge le cache (cotations + frais) du catalogue
	./portfodor -warmup

.PHONY: simdata
simdata: build ## (Re)génère les historiques simulés validés dans simdata/
	./simgen-bin

.PHONY: demo
demo: build ## Rapport de démonstration sur les portefeuilles d'exemple
	./portfodor examples/*.txt

.PHONY: clean
clean: ## Supprime les binaires (pas les caches data/ ni simdata/)
	rm -f $(BINARIES)

.PHONY: help
help: ## Affiche cette aide
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'
