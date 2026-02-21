package logger_test

import (
	"strings"
	"testing"
	"time"

	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/testutils"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func(l *logger.DefaultLogger, message string, args ...any) *logger.LogEntry
		message  string
		args     []any
		expected string
	}{
		{
			name: "Default message",
			logFunc: func(l *logger.DefaultLogger, msg string, args ...any) *logger.LogEntry {
				return l.Default(msg, args...)
			},
			message:  "Operation completed successfully",
			expected: "Operation completed successfully\n",
		},
		{
			name: "Info message with args",
			logFunc: func(l *logger.DefaultLogger, msg string, args ...any) *logger.LogEntry {
				return l.Info(msg, args...)
			},
			message:  "Application started",
			args:     []any{"port:", 8080, "version:", "1.0.0"},
			expected: "ℹ Application started port: 8080 version: 1.0.0\n",
		},
		{
			name: "Success message",
			logFunc: func(l *logger.DefaultLogger, msg string, args ...any) *logger.LogEntry {
				return l.Success(msg, args...)
			},
			message:  "Operation completed",
			expected: "✔ Operation completed\n",
		},
		{
			name: "Warning message",
			logFunc: func(l *logger.DefaultLogger, msg string, args ...any) *logger.LogEntry {
				return l.Warning(msg, args...)
			},
			message:  "Low disk space",
			expected: "⚠ Low disk space\n",
		},
		{
			name: "Error message",
			logFunc: func(l *logger.DefaultLogger, msg string, args ...any) *logger.LogEntry {
				return l.Error(msg, args...)
			},
			message:  "Failed to connect to database",
			expected: "✘ Failed to connect to database\n",
		},
		{
			name: "Indented Success message",
			logFunc: func(l *logger.DefaultLogger, msg string, args ...any) *logger.LogEntry {
				l.WithIndent(true)
				entry := l.Success(msg, args...)
				l.WithIndent(false) // Reset after the test.
				return entry
			},
			message:  "Indented operation completed",
			expected: "  ✔ Indented operation completed\n",
		},
		{
			name: "Message with attributes",
			logFunc: func(l *logger.DefaultLogger, msg string, args ...any) *logger.LogEntry {
				return l.Success(msg).WithAttrs("items", 42, "duration", "1s")
			},
			message:  "Operation completed",
			expected: "✔ Operation completed\n  - items: 42\n  - duration: 1s\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := logger.NewDefaultLogger() // Create a new logger instance per test.

			output, err := testutils.CaptureStdout(func() {
				tt.logFunc(l, tt.message, tt.args...)
			})
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			if output != tt.expected {
				t.Errorf("Unexpected output:\nGot: %q\nWant: %q", output, tt.expected)
			}
		})
	}
}

func TestLoggerToJSON(t *testing.T) {
	l := logger.NewDefaultLogger()
	l.WithTimestamp(true)

	tests := []struct {
		name        string
		logFunc     func(*logger.DefaultLogger, string, ...any) *logger.LogEntry
		message     string
		args        []any
		attributes  []any
		expectedKey string
	}{
		{
			name:        "Success with attributes",
			logFunc:     (*logger.DefaultLogger).Success,
			message:     "Operation succeeded",
			args:        nil,
			attributes:  []any{"user", "JohnDoe", "action", "login"},
			expectedKey: "attributes",
		},
		{
			name:        "Error with attributes",
			logFunc:     (*logger.DefaultLogger).Error,
			message:     "Database error",
			args:        nil,
			attributes:  []any{"code", 500, "reason", "timeout"},
			expectedKey: "attributes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := tt.logFunc(l, tt.message, tt.args...)
			entry.WithAttrs(tt.attributes...)

			jsonData, err := entry.ToJSON()
			if err != nil {
				t.Fatalf("unexpected error serializing to JSON: %v", err)
			}

			if !strings.Contains(jsonData, tt.expectedKey) {
				t.Errorf("expected JSON to contain key %q, got: %s", tt.expectedKey, jsonData)
			}
		})
	}
}

func TestLoggerWithTimestamp(t *testing.T) {
	l := logger.NewDefaultLogger()
	l.WithTimestamp(true)

	output, err := testutils.CaptureStdout(func() {
		l.Info("Timestamped log")
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	if !strings.Contains(output, "Timestamped log") {
		t.Errorf("expected output to contain the log message, got: %q", output)
	}

	// Check for a partial timestamp (e.g., the current date in "2006-01-02" format).
	if !strings.Contains(output, time.Now().Format("2006-01-02")) {
		t.Errorf("expected output to contain the current date, got: %q", output)
	}
}

func TestLogAttrsWithIndent(t *testing.T) {
	tests := []struct {
		name           string
		indentEnabled  bool
		attrs          []any
		expectedOutput string
	}{
		{
			name:          "Attributes without indentation",
			indentEnabled: false,
			attrs:         []any{"key1", "value1", "key2", "value2"},
			expectedOutput: `✔ Testing attributes
  - key1: value1
  - key2: value2
`,
		},
		{
			name:          "Attributes with indentation",
			indentEnabled: true,
			attrs:         []any{"key1", "value1", "key2", "value2"},
			expectedOutput: `  ✔ Testing attributes
    - key1: value1
    - key2: value2
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := logger.NewDefaultLogger()
			l.WithTimestamp(false)
			l.WithIndent(tt.indentEnabled)

			output, err := testutils.CaptureStdout(func() {
				l.Success("Testing attributes").WithAttrs(tt.attrs...)
			})
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			if output != tt.expectedOutput {
				t.Errorf("Unexpected output:\nGot:\n%q\nWant:\n%q", output, tt.expectedOutput)
			}
		})
	}
}
