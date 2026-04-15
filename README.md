# service_hub_notifier

`service_hub_notifier` is the notification transport service for Service Hub.

It receives normalized delivery payloads from Phoenix and sends them to provider APIs (Telegram, Slack).

## Responsibilities

- Expose `POST /api/v1/deliveries` for normalized delivery requests.
- Validate and authenticate internal service-to-service traffic.
- Deliver Telegram and Slack messages.
- Return normalized success/failure responses with retry hints.
- Emit structured JSON logs and lightweight metrics.

## Non-responsibilities

- No business-rule orchestration (`who`, `when`, `whether`).
- No queueing/retry scheduling policy (owned by Phoenix + Oban).
- No escalation, cooldown, dedupe, or routing decisions.

## Runtime endpoints

- `GET /health`:
  - process liveness check.
  - returns `200` with `{"status":"ok"}`.
- `GET /ready`:
  - readiness check for required runtime config.
  - returns `200` with `{"status":"ready", ...}` when ready.
  - returns `503` with `{"status":"not_ready", ...}` when required config is missing.
- `GET /metrics`:
  - Prometheus text exposition.
  - includes `delivery_total`, `delivery_failed_total`, and `provider_latency_ms` summary data.

## Environment variables

- `PORT` (default: `8081`)
- `DELIVERY_TIMEOUT_MS` (default: `5000`)
- `SHUTDOWN_TIMEOUT_MS` (default: `10000`)
- `TELEGRAM_API_BASE_URL` (default: `https://api.telegram.org`)
- `INTERNAL_SERVICE_TOKEN` (required for production readiness and auth)

## Internal service authentication

`POST /api/v1/deliveries` requires:

`Authorization: Bearer <INTERNAL_SERVICE_TOKEN>`

Requests without a valid token return `401` with `error_code: "unauthorized"`.

## Run locally

```bash
go run ./cmd/server
```

Graceful shutdown is enabled for `SIGTERM`/`SIGINT` via `http.Server.Shutdown`.

## Testing workflow

Layered test strategy:

1. Unit tests (`go test ./...`)
2. HTTP integration tests (real server with mock providers)
3. Container/network integration tests (Docker Compose)

### Local test commands

```bash
go test ./... -race -cover
docker compose -f docker-compose.test.yml up --build --abort-on-container-exit --exit-code-from phoenix_mock
docker compose -f docker-compose.test.yml down -v --remove-orphans
```

`docker-compose.test.yml` validates:

- notifier liveness/readiness
- internal auth enforcement
- service DNS connectivity on private network
- notifier-to-provider mock connectivity

## CI pipeline

Workflow: `.github/workflows/ci.yml`

- `unit`: `gofmt`, `go vet`, `go test ./... -race -cover`
- `build-container`: build image and verify startup + `/health`
- `integration`: compose-based container/network tests
- `release-image` (push events): build/push versioned images to GHCR

## Production deployment model

Single-VM deployment uses a private Docker bridge network:

- `nginx` (edge)
- `phoenix` (`service_hub`)
- `notifier` (`service_hub_notifier`)

Notifier must stay private (no published host port) and is reachable by internal DNS:

`http://notifier:8081`

Reference compose stack: `../service_hub/docker-compose.prod.yml`.

## Image release strategy

Images are pushed with immutable tags (no `latest` dependency):

- semantic version tags (for release pushes like `v1.2.3`)
- git SHA tags (`sha-<short-commit>`)

Examples:

- `ghcr.io/<owner>/<repo>:v1.2.3`
- `ghcr.io/<owner>/<repo>:sha-abc1234`

## API contract

### `POST /api/v1/deliveries`

Request shape:

```json
{
  "delivery_attempt_key": "event-id:channel-id",
  "provider": "telegram",
  "destination": {
    "token": "<bot-token>",
    "chat_ref": "@alerts",
    "parse_mode": "HTML"
  },
  "notification": {
    "event_name": "health.alert",
    "check_type": "health",
    "severity": "alert",
    "message": "Health check failed",
    "service_id": 1,
    "deployment_id": 2,
    "metadata": {
      "host": "example.com",
      "env": "production"
    }
  },
  "event": {
    "id": "event-id",
    "name": "health.alert",
    "tags": {
      "source": "automatic"
    }
  }
}
```

Success response:

```json
{
  "status": "delivered",
  "provider_message_id": "123",
  "provider_response_code": "200",
  "provider_response": {
    "ok": true
  }
}
```

Failure response:

```json
{
  "status": "failed",
  "retryable": false,
  "error_code": "invalid_destination",
  "error_message": "missing chat_ref",
  "provider_response_code": "400",
  "provider_response": {
    "ok": false
  }
}
```
