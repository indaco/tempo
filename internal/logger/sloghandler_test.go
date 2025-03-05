package logger

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/indaco/tempo/internal/testhelpers"
)

func TestLogHandler(t *testing.T) {
	loggerInstance := NewDefaultLogger()
	handler := NewLogHandler(loggerInstance)
	slogLogger := slog.New(handler)

	tests := []struct {
		name           string
		logFunc        func()
		expectedOutput string
		indent         bool
	}{
		{
			name: "Info log without attributes",
			logFunc: func() {
				slogLogger.Info("Application started")
			},
			expectedOutput: `ℹ Application started`,
			indent:         false,
		},
		{
			name: "Warning log with attributes",
			logFunc: func() {
				slogLogger.Warn("Low disk space",
					slog.String("disk", "C:"),
					slog.Int("available_gb", 5),
				)
			},
			expectedOutput: `⚠ Low disk space
  - disk: C:
  - available_gb: 5`,
			indent: false,
		},
		{
			name: "Error log with attributes",
			logFunc: func() {
				slogLogger.Error("Unexpected crash",
					slog.String("module", "auth"),
					slog.String("reason", "panic"),
				)
			},
			expectedOutput: `✘ Unexpected crash
  - module: auth
  - reason: panic`,
			indent: false,
		},
		{
			name: "Default log level (debug equivalent)",
			logFunc: func() {
				slogLogger.Debug("Debugging log")
			},
			expectedOutput: `Debugging log`,
			indent:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := testhelpers.CaptureStdout(tt.logFunc)
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			// Trim spaces for comparison.
			output = strings.TrimSpace(output)
			expected := strings.TrimSpace(tt.expectedOutput)

			if output != expected {
				t.Errorf("Unexpected output:\nGot: %q\nWant: %q", output, expected)
			}
		})
	}
}

func TestLogHandlerWithAttrs(t *testing.T) {
	loggerInstance := NewDefaultLogger()
	handler := NewLogHandler(loggerInstance)
	slogLogger := slog.New(handler)

	tests := []struct {
		name           string
		logFunc        func()
		expectedOutput string
		indent         bool
	}{
		{
			name: "Log message with attributes",
			logFunc: func() {
				slogLogger.Info("Starting application",
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
			output, err := testhelpers.CaptureStdout(tt.logFunc)
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			output = strings.TrimSpace(output)
			expected := strings.TrimSpace(tt.expectedOutput)

			if output != expected {
				t.Errorf("Unexpected output:\nGot: %q\nWant: %q", output, expected)
			}
		})
	}
}

func TestLogHandlerWithGroup(t *testing.T) {
	loggerInstance := NewDefaultLogger()
	handler := NewLogHandler(loggerInstance)

	// Create a grouped handler.
	groupedHandler := handler.WithGroup("database")
	slogLogger := slog.New(groupedHandler)

	output, err := testhelpers.CaptureStdout(func() {
		slogLogger.Warn("Connection slow",
			slog.String("latency", "500ms"),
		)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	expectedOutput := `⚠ Connection slow
  - database.latency: 500ms`

	output = strings.TrimSpace(output)
	expectedOutput = strings.TrimSpace(expectedOutput)

	if output != expectedOutput {
		t.Errorf("Unexpected output:\nGot: %q\nWant: %q", output, expectedOutput)
	}
}

func TestGroupedLogHandlerWithAttrs(t *testing.T) {
	loggerInstance := NewDefaultLogger()
	handler := NewLogHandler(loggerInstance)

	// Create a grouped handler and add additional attributes.
	groupedHandler := handler.WithGroup("database")
	newHandler := groupedHandler.WithAttrs([]slog.Attr{
		slog.String("query", "SELECT * FROM users"),
		slog.Int("execution_time", 120),
	})
	slogLogger := slog.New(newHandler)

	output, err := testhelpers.CaptureStdout(func() {
		slogLogger.Info("Query executed")
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	expectedOutput := `ℹ Query executed
  - database.query: SELECT * FROM users
  - database.execution_time: 120`

	output = strings.TrimSpace(output)
	expectedOutput = strings.TrimSpace(expectedOutput)

	if output != expectedOutput {
		t.Errorf("Unexpected output:\nGot: %q\nWant: %q", output, expectedOutput)
	}
}

func TestGroupedLogHandlerWithNestedGroups(t *testing.T) {
	loggerInstance := NewDefaultLogger()
	handler := NewLogHandler(loggerInstance)

	// Create a nested grouped handler.
	groupedHandler := handler.WithGroup("database").WithGroup("connection")
	slogLogger := slog.New(groupedHandler)

	output, err := testhelpers.CaptureStdout(func() {
		slogLogger.Warn("Slow response",
			slog.String("latency", "500ms"),
		)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	expectedOutput := `⚠ Slow response
  - database.connection.latency: 500ms`

	output = strings.TrimSpace(output)
	expectedOutput = strings.TrimSpace(expectedOutput)

	if output != expectedOutput {
		t.Errorf("Unexpected output:\nGot: %q\nWant: %q", output, expectedOutput)
	}
}
