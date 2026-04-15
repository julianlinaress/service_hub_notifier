package logger

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"
)

func TestLoggerInfo(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	l := &Logger{base: log.New(buffer, "", 0)}

	l.Info("delivery_completed", map[string]any{"provider": "telegram"})

	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buffer.Bytes()), &entry); err != nil {
		t.Fatalf("unmarshal log entry: %v", err)
	}

	if entry["level"] != "info" {
		t.Fatalf("level = %v, want %q", entry["level"], "info")
	}

	if entry["message"] != "delivery_completed" {
		t.Fatalf("message = %v, want %q", entry["message"], "delivery_completed")
	}

	if entry["provider"] != "telegram" {
		t.Fatalf("provider = %v, want %q", entry["provider"], "telegram")
	}

	if _, ok := entry["ts"]; !ok {
		t.Fatalf("ts field missing")
	}
}

func TestLoggerEncodingFallback(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	l := &Logger{base: log.New(buffer, "", 0)}

	l.Info("will_fail", map[string]any{"bad": make(chan int)})

	out := buffer.String()
	if out == "" {
		t.Fatalf("expected fallback output")
	}

	if !bytes.Contains(buffer.Bytes(), []byte("logger_encoding_failed")) {
		t.Fatalf("expected logger_encoding_failed in output, got %q", out)
	}
}
