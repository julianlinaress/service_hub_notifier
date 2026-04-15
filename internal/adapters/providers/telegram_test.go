package providers

import (
	"errors"
	"strings"
	"testing"
)

func TestSanitizeProviderError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     error
		want    string
		wantNot string
	}{
		{
			name:    "redacts telegram token in url",
			err:     errors.New("Post \"https://api.telegram.org/bot123456:ABC/sendMessage\": dial tcp timeout"),
			want:    "Post \"https://api.telegram.org/bot<redacted>/sendMessage\": dial tcp timeout",
			wantNot: "bot123456:ABC",
		},
		{
			name:    "keeps non telegram errors",
			err:     errors.New("dial tcp 10.0.0.1:443: connect: refused"),
			want:    "dial tcp 10.0.0.1:443: connect: refused",
			wantNot: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := sanitizeProviderError(tc.err)
			if got != tc.want {
				t.Fatalf("sanitizeProviderError() = %q, want %q", got, tc.want)
			}

			if tc.wantNot != "" && strings.Contains(got, tc.wantNot) {
				t.Fatalf("sanitizeProviderError() leaked token: %q", got)
			}
		})
	}
}
