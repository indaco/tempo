package logger

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/indaco/tempo/testutils"
)

func TestLogHandler(t *testing.T) {
	loggerInstance := NewDefaultLogger()
	handler := NewLogHandler(loggerInstance)
	logger := slog.New(handler)

	tests := []struct {
		name           string
		logFunc        func()
		expectedOutput string
		indent         bool
	}{
		{
			name: "Log message with attributes",
			logFunc: func() {
				logger.Info("Starting application",
					slog.String("version", "1.0.0"),
					slog.Int("port", 8080),
				)
			},
			expectedOutput: `ℹ Starting application
  - version: 1.0.0
  - port: 8080`,
			indent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			output, err := testutils.CaptureStdout(tt.logFunc)
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			// Trim spaces for cleaner comparison
			output = strings.TrimSpace(output)
			expected := strings.TrimSpace(tt.expectedOutput)

			if output != expected {
				t.Errorf("Unexpected output:\nGot: %q\nWant: %q", output, expected)
			}
		})
	}
}
