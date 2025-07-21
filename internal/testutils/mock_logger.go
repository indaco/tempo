package testutils

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/indaco/tempo/internal/logger"
)

// MockLogger is a mock implementation of the LoggerInterface for testing purposes.
type MockLogger struct {
	Logs             []string
	indentEnabled    bool
	timestampEnabled bool
}

// log is a helper that formats the log message and appends it to Logs.
func (m *MockLogger) log(level, icon, message string) *logger.LogEntry {
	formatted := message
	if icon != "" {
		formatted = fmt.Sprintf("%s %s", icon, message)
	}
	if m.indentEnabled {
		formatted = "    " + formatted
	}
	if m.timestampEnabled {
		ts := time.Now().Format(time.RFC3339)
		formatted = fmt.Sprintf("%s %s", ts, formatted)
	}
	m.Logs = append(m.Logs, formatted)
	// Return a dummy LogEntry.
	return &logger.LogEntry{}
}

// Default logs a default message.
func (m *MockLogger) Default(message string, args ...any) *logger.LogEntry {
	return m.log("default", "", fmt.Sprintf(message, args...))
}

// Info logs an informational message.
func (m *MockLogger) Info(message string, args ...any) *logger.LogEntry {
	return m.log("info", "ℹ", fmt.Sprintf(message, args...))
}

// Success logs a success message.
func (m *MockLogger) Success(message string, args ...any) *logger.LogEntry {
	return m.log("success", "✔", fmt.Sprintf(message, args...))
}

// Warning logs a warning message.
func (m *MockLogger) Warning(message string, args ...any) *logger.LogEntry {
	return m.log("warning", "⚠", fmt.Sprintf(message, args...))
}

// Error logs an error message.
func (m *MockLogger) Error(message string, args ...any) *logger.LogEntry {
	return m.log("error", "✘", fmt.Sprintf(message, args...))
}

// Hint logs a hint message.
func (m *MockLogger) Hint(message string, args ...any) *logger.LogEntry {
	return m.log("hint", "💡", fmt.Sprintf(message, args...))
}

// Blank prints a blank line.
func (m *MockLogger) Blank() {
	m.Logs = append(m.Logs, "")
	_, _ = fmt.Fprintln(color.Output)
}

// WithTimestamp enables or disables the inclusion of timestamps in log messages.
func (m *MockLogger) WithTimestamp(enabled bool) {
	m.timestampEnabled = enabled
}

// WithIndent enables or disables indentation for log messages.
func (m *MockLogger) WithIndent(enabled bool) {
	m.indentEnabled = enabled
}

// Reset clears all stored log messages.
func (m *MockLogger) Reset() {
	m.Logs = []string{}
}
