# syntax=docker/dockerfile:1.7
FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -extldflags=-static" \
    -trimpath \
    -o /app ./cmd/api

# The backfill runner ships in the same image so the Job that invokes it can
# reuse the backend's image tag and DB secret rather than needing its own
# build, registry entry and credential path. It is a separate entrypoint, never
# started by the API container.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -extldflags=-static" \
    -trimpath \
    -o /backfill ./cmd/backfill

# ── Development stage ──────────────────────────────────────────────────────
# Alpine-based: includes wget for docker-compose healthchecks.
# Used by docker-compose.yml via `target: dev`.
FROM alpine:3.21 AS dev

RUN addgroup -g 1000 appgroup \
 && adduser -D -u 1000 -G appgroup appuser

COPY --from=builder /app /app
COPY --from=builder /backfill /backfill
COPY --chown=appuser:appgroup migrations /migrations

USER appuser:appgroup

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=5s --start-period=30s --retries=5 \
    CMD wget -qO- http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app"]

# ── Production stage ───────────────────────────────────────────────────────
# Distroless: no shell, no package manager — minimal CVE surface.
# Used by Helm/k8s. Build with: docker build --target production .
FROM gcr.io/distroless/static-debian12:nonroot AS production

COPY --from=builder /app /app
COPY --from=builder /backfill /backfill
COPY --chown=nonroot:nonroot migrations /migrations

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/app"]
