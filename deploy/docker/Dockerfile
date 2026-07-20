# syntax=docker/dockerfile:1.7
# pofo: a minimal static runtime image (multi-stage).
#
# Build:  DOCKER_BUILDKIT=1 docker build -f deploy/docker/Dockerfile -t pofo:latest .
#         (or: make docker-image)
# Run:    docker run --rm -p 127.0.0.1:8787:8787 -v pofo-cache:/var/lib/pofo pofo:latest
#
# The whole stack is stdlib-only Go, so CGO_ENABLED=0 yields a fully static
# binary and the runtime stage needs only CA certificates (for the HTTPS quote
# fetches) and tzdata (date handling). The datasets, examples and book are
# embedded via go:embed, so no data files are copied into the image; only the
# on-disk quote cache is a mounted volume.
#
# Requires BuildKit (DOCKER_BUILDKIT=1) for the --mount=type=cache lines, which
# persist /root/.cache/go-build and /go/pkg/mod across rebuilds.

ARG GO_VERSION=1.26.3
ARG ALPINE_VERSION=3.20

# ---------- 1. build stage ----------
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /src

# go.mod first so the (dependency-free) module layer caches independently of
# the source; keeps the boundary clean if a dependency is ever added.
COPY go.mod ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Application source, including the go:embed datasets, examples and book.
COPY . .

ENV CGO_ENABLED=0 GOOS=linux
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -trimpath -ldflags="-s -w" -o /out/pofo ./cmd/pofo

# ---------- 2. runtime stage ----------
FROM alpine:${ALPINE_VERSION} AS runtime

# ca-certificates for the HTTPS quote fetches, tzdata for date handling.
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 1000 pofo \
    && adduser -S -u 1000 -G pofo -h /var/lib/pofo -s /sbin/nologin pofo \
    && install -d -o pofo -g pofo -m 0750 /var/lib/pofo

COPY --from=builder /out/pofo /usr/local/bin/pofo

USER pofo
WORKDIR /var/lib/pofo
EXPOSE 8787
VOLUME ["/var/lib/pofo"]

# Serve the web app on all interfaces inside the container; publish it to the
# host's loopback (-p 127.0.0.1:8787:8787) behind whatever private edge you
# run. The quote cache lives on the mounted volume so restarts do not re-fetch.
ENTRYPOINT ["/usr/local/bin/pofo"]
CMD ["-serve", "-listen", "0.0.0.0:8787", "-data", "/var/lib/pofo"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s \
  CMD wget -qO- http://127.0.0.1:8787/ >/dev/null 2>&1 || exit 1
