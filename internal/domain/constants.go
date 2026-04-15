package domain

// Delivery provider identifiers accepted by the service.
const (
	ProviderTelegram = "telegram"
	ProviderSlack    = "slack"
)

// Delivery lifecycle statuses returned by transport adapters.
const (
	StatusDelivered = "delivered"
	StatusFailed    = "failed"
)

// Normalized API error codes returned by handlers and adapters.
const (
	ErrInvalidRequest      = "invalid_request"
	ErrInvalidJSON         = "invalid_json"
	ErrInvalidDestination  = "invalid_destination"
	ErrPayloadTooLarge     = "payload_too_large"
	ErrUnsupportedProvider = "unsupported_provider"
	ErrEncoding            = "encoding_error"
	ErrRequestBuildFailed  = "request_build_failed"
	ErrProviderRequest     = "provider_request_failed"
	ErrSlackSendFailed     = "slack_send_failed"
	ErrTelegramSendFailed  = "telegram_send_failed"
)
