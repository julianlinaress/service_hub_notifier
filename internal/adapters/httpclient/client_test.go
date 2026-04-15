package httpclient

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "positive timeout",
			timeout: 5 * time.Second,
		},
		{
			name:    "zero timeout",
			timeout: 0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client := New(tc.timeout)
			if client.Timeout != tc.timeout {
				t.Fatalf("timeout = %s, want %s", client.Timeout, tc.timeout)
			}
		})
	}
}
