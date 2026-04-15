package providers

import "net/http"

const defaultTelegramAPIBaseURL = "https://api.telegram.org"

// HTTPClient describes the network dependency used by provider adapters.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// TelegramOption customizes Telegram adapter behavior.
type TelegramOption func(*TelegramAdapter)

// WithTelegramAPIBaseURL sets the Telegram API base URL.
func WithTelegramAPIBaseURL(baseURL string) TelegramOption {
	return func(adapter *TelegramAdapter) {
		adapter.apiBaseURL = baseURL
	}
}
