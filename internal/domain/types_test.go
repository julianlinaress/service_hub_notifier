package domain

import "testing"

func TestSuccessResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		messageID        string
		responseCode     string
		providerResponse map[string]any
	}{
		{
			name:             "initializes nil provider response",
			messageID:        "123",
			responseCode:     "200",
			providerResponse: nil,
		},
		{
			name:             "keeps existing provider response",
			messageID:        "abc",
			responseCode:     "201",
			providerResponse: map[string]any{"ok": true},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := SuccessResponse(tc.messageID, tc.responseCode, tc.providerResponse)

			if got.Status != StatusDelivered {
				t.Fatalf("status = %q, want %q", got.Status, StatusDelivered)
			}

			if got.ProviderMessageID != tc.messageID {
				t.Fatalf("provider_message_id = %q, want %q", got.ProviderMessageID, tc.messageID)
			}

			if got.ProviderResponseCode != tc.responseCode {
				t.Fatalf("provider_response_code = %q, want %q", got.ProviderResponseCode, tc.responseCode)
			}

			if got.ProviderResponse == nil {
				t.Fatalf("provider_response = nil, want non-nil map")
			}
		})
	}
}

func TestFailureResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		retryable        bool
		errorCode        string
		errorMessage     string
		responseCode     string
		providerResponse map[string]any
	}{
		{
			name:             "initializes nil provider response",
			retryable:        true,
			errorCode:        ErrInvalidRequest,
			errorMessage:     "bad request",
			responseCode:     "400",
			providerResponse: nil,
		},
		{
			name:             "keeps existing provider response",
			retryable:        false,
			errorCode:        ErrInvalidDestination,
			errorMessage:     "invalid destination",
			responseCode:     "422",
			providerResponse: map[string]any{"raw": "error"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FailureResponse(tc.retryable, tc.errorCode, tc.errorMessage, tc.responseCode, tc.providerResponse)

			if got.Status != StatusFailed {
				t.Fatalf("status = %q, want %q", got.Status, StatusFailed)
			}

			if got.Retryable != tc.retryable {
				t.Fatalf("retryable = %t, want %t", got.Retryable, tc.retryable)
			}

			if got.ErrorCode != tc.errorCode {
				t.Fatalf("error_code = %q, want %q", got.ErrorCode, tc.errorCode)
			}

			if got.ErrorMessage != tc.errorMessage {
				t.Fatalf("error_message = %q, want %q", got.ErrorMessage, tc.errorMessage)
			}

			if got.ProviderResponseCode != tc.responseCode {
				t.Fatalf("provider_response_code = %q, want %q", got.ProviderResponseCode, tc.responseCode)
			}

			if got.ProviderResponse == nil {
				t.Fatalf("provider_response = nil, want non-nil map")
			}
		})
	}
}
