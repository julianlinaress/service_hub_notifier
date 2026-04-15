package httpclient

import (
	"net/http"
	"time"
)

// New creates an HTTP client configured with the provided timeout.
func New(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}
