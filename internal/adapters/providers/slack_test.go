package providers

import (
	"strings"
	"testing"
)

func TestParseSlackBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want map[string]any
	}{
		{
			name: "json payload",
			body: `{"ok":true,"channel":"alerts"}`,
			want: map[string]any{
				"ok":      true,
				"channel": "alerts",
			},
		},
		{
			name: "text payload",
			body: "ok",
			want: map[string]any{
				"raw": "ok",
			},
		},
		{
			name: "empty payload",
			body: "\n\t",
			want: map[string]any{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := parseSlackBody(strings.NewReader(tc.body))
			if len(got) != len(tc.want) {
				t.Fatalf("len(got) = %d, want %d", len(got), len(tc.want))
			}

			for key, wantVal := range tc.want {
				if got[key] != wantVal {
					t.Fatalf("got[%q] = %#v, want %#v", key, got[key], wantVal)
				}
			}
		})
	}
}
