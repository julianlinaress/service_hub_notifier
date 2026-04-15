# service_hub_notifier

`service_hub_notifier` is the delivery transport service for Service Hub notifications.

## Responsibilities

- Expose `POST /api/v1/deliveries` for normalized delivery requests
- Deliver Telegram messages
- Deliver Slack messages
- Return normalized success/failure responses
- Emit structured JSON logs

## Non-responsibilities

- No business rule decisions (who/when/whether to notify)
- No queueing or retry orchestration
- No escalation, cooldown, dedupe policy, or routing decisions

All orchestration remains in Phoenix + Oban.

## Run locally

```bash
go run ./cmd/server
```

Environment variables:

- `PORT` (default: `8081`)
- `DELIVERY_TIMEOUT_MS` (default: `5000`)

## API

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
