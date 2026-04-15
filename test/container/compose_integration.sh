#!/bin/sh
set -eu

HEALTH_URL="http://notifier:8081/health"
DELIVERIES_URL="http://notifier:8081/api/v1/deliveries"

for _ in $(seq 1 30); do
  if curl -fsS "$HEALTH_URL" >/dev/null; then
    break
  fi
  sleep 1
done

curl -fsS "$HEALTH_URL" >/dev/null

ready_code=$(curl -sS -o /tmp/ready_resp.json -w "%{http_code}" "http://notifier:8081/ready")
if [ "$ready_code" != "200" ]; then
  echo "ready endpoint expected status 200"
  cat /tmp/ready_resp.json
  exit 1
fi

case "$(cat /tmp/ready_resp.json)" in
  *'"status":"ready"'*) ;;
  *)
    echo "ready endpoint response missing ready status"
    cat /tmp/ready_resp.json
    exit 1
    ;;
esac

unauthorized_payload='{
  "delivery_attempt_key": "compose-event:unauthorized",
  "provider": "slack",
  "destination": {
    "webhook_url": "http://slack_mock:80/webhook"
  },
  "notification": {
    "event_name": "health.alert",
    "check_type": "health",
    "severity": "alert",
    "message": "unauthorized check",
    "service_id": 1,
    "deployment_id": 2,
    "metadata": {
      "host": "compose.local",
      "env": "test"
    }
  },
  "event": {
    "id": "compose-unauthorized",
    "name": "health.alert",
    "tags": {
      "source": "phoenix_mock"
    }
  }
}'

status_code=$(curl -sS -o /tmp/unauthorized_resp.json -w "%{http_code}" \
  -X POST "$DELIVERIES_URL" \
  -H "Content-Type: application/json" \
  -d "$unauthorized_payload")

if [ "$status_code" != "401" ]; then
  echo "unauthorized request expected status 401"
  cat /tmp/unauthorized_resp.json
  exit 1
fi

AUTH_HEADER="Authorization: Bearer ${INTERNAL_SERVICE_TOKEN}"

telegram_payload='{
  "delivery_attempt_key": "compose-event:telegram",
  "provider": "telegram",
  "destination": {
    "token": "compose-token",
    "chat_ref": "@alerts"
  },
  "notification": {
    "event_name": "health.alert",
    "check_type": "health",
    "severity": "alert",
    "message": "compose telegram delivery",
    "service_id": 1,
    "deployment_id": 2,
    "metadata": {
      "host": "compose.local",
      "env": "test"
    }
  },
  "event": {
    "id": "compose-telegram",
    "name": "health.alert",
    "tags": {
      "source": "phoenix_mock"
    }
  }
}'

status_code=$(curl -sS -o /tmp/telegram_resp.json -w "%{http_code}" \
  -X POST "$DELIVERIES_URL" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d "$telegram_payload")

if [ "$status_code" != "200" ]; then
  echo "telegram delivery failed with status $status_code"
  cat /tmp/telegram_resp.json
  exit 1
fi

telegram_body=$(cat /tmp/telegram_resp.json)
case "$telegram_body" in
  *'"status":"delivered"'*) ;;
  *)
    echo "telegram delivery response missing delivered status"
    echo "$telegram_body"
    exit 1
    ;;
esac

slack_payload='{
  "delivery_attempt_key": "compose-event:slack",
  "provider": "slack",
  "destination": {
    "webhook_url": "http://slack_mock:80/webhook"
  },
  "notification": {
    "event_name": "health.alert",
    "check_type": "health",
    "severity": "alert",
    "message": "compose slack delivery",
    "service_id": 1,
    "deployment_id": 2,
    "metadata": {
      "host": "compose.local",
      "env": "test"
    }
  },
  "event": {
    "id": "compose-slack",
    "name": "health.alert",
    "tags": {
      "source": "phoenix_mock"
    }
  }
}'

status_code=$(curl -sS -o /tmp/slack_resp.json -w "%{http_code}" \
  -X POST "$DELIVERIES_URL" \
  -H "Content-Type: application/json" \
  -H "$AUTH_HEADER" \
  -d "$slack_payload")

if [ "$status_code" != "200" ]; then
  echo "slack delivery failed with status $status_code"
  cat /tmp/slack_resp.json
  exit 1
fi

slack_body=$(cat /tmp/slack_resp.json)
case "$slack_body" in
  *'"status":"delivered"'*) ;;
  *)
    echo "slack delivery response missing delivered status"
    echo "$slack_body"
    exit 1
    ;;
esac

echo "compose integration checks passed"
