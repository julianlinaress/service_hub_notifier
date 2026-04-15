ARG GO_VERSION=1.22.7

FROM golang:${GO_VERSION}-alpine3.20 AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/service_hub_notifier ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot AS runtime

WORKDIR /app

COPY --from=builder /out/service_hub_notifier /app/service_hub_notifier

EXPOSE 8081
STOPSIGNAL SIGTERM
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD ["/app/service_hub_notifier", "healthcheck"]
USER nonroot:nonroot

ENTRYPOINT ["/app/service_hub_notifier"]
