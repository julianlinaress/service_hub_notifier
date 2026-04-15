package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// Logger emits structured JSON log entries.
type Logger struct {
	base *log.Logger
}

// New creates a JSON logger that writes to stdout.
func New() *Logger {
	return &Logger{base: log.New(os.Stdout, "", 0)}
}

// Info emits an info-level structured log entry.
func (l *Logger) Info(message string, fields map[string]any) {
	l.log("info", message, fields)
}

// Error emits an error-level structured log entry.
func (l *Logger) Error(message string, fields map[string]any) {
	l.log("error", message, fields)
}

func (l *Logger) log(level string, message string, fields map[string]any) {
	entry := map[string]any{
		"ts":      time.Now().UTC().Format(time.RFC3339Nano),
		"level":   level,
		"message": message,
	}

	for key, value := range fields {
		entry[key] = value
	}

	encoded, err := json.Marshal(entry)
	if err != nil {
		l.base.Printf(`{"ts":"%s","level":"error","message":"logger_encoding_failed","error":"%s"}`,
			time.Now().UTC().Format(time.RFC3339Nano),
			err.Error(),
		)
		return
	}

	l.base.Print(fmt.Sprintf("%s", encoded))
}
