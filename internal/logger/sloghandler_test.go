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
			name: "Info log without attributes",
			logFunc: func() {
				logger.Info("Application started")
			},
			expectedOutput: `ℹ Application started`,
			indent:         false,
		},
		{
			name: "Warning log with attributes",
			logFunc: func() {
				logger.Warn("Low disk space",
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
				logger.Error("Unexpected crash",
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
				logger.Debug("Debugging log")
			},
			expectedOutput: `Debugging log`, // No icon for default level
			indent:         false,
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

func TestLogHandlerWithAttrs(t *testing.T) {
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

func TestLogHandlerWithGroup(t *testing.T) {
	loggerInstance := NewDefaultLogger()
	handler := NewLogHandler(loggerInstance)

	// Use WithGroup to create a grouped handler
	groupedHandler := handler.WithGroup("database")
	logger := slog.New(groupedHandler)

	output, err := testutils.CaptureStdout(func() {
		logger.Warn("Connection slow",
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

	// Create a grouped handler
	groupedHandler := handler.WithGroup("database")

	// Create a new handler with additional attributes
	newHandler := groupedHandler.WithAttrs([]slog.Attr{
		slog.String("query", "SELECT * FROM users"),
		slog.Int("execution_time", 120),
	})

	logger := slog.New(newHandler)

	output, err := testutils.CaptureStdout(func() {
		logger.Info("Query executed")
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

	// Create a nested grouped handler
	groupedHandler := handler.WithGroup("database").WithGroup("connection")
	logger := slog.New(groupedHandler)

	output, err := testutils.CaptureStdout(func() {
		logger.Warn("Slow response",
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
