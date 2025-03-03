package testutils

import (
	"errors"
	"fmt"
)

// MockLogger is a mock implementation of the Logger interface for testing purposes.
type MockLogger struct {
	Logs          []string
	indentEnabled bool
}

func (m *MockLogger) Default(message string, args ...any) {
	m.Logs = append(m.Logs, fmt.Sprintf(message, args...))
}

func (m *MockLogger) Success(message string, args ...any) {
	m.Logs = append(m.Logs, fmt.Sprintf("✔ "+message, args...))
}

func (m *MockLogger) SuccessColored(message string, args ...any) {
	m.Logs = append(m.Logs, fmt.Sprintf("[Colored] ✔ "+message, args...))
}

func (m *MockLogger) Info(message string, args ...any) {
	m.Logs = append(m.Logs, fmt.Sprintf("ℹ "+message, args...))
}

func (m *MockLogger) InfoColored(message string, args ...any) {
	m.Logs = append(m.Logs, fmt.Sprintf("[Colored] ℹ "+message, args...))
}

func (m *MockLogger) Warning(message string, args ...any) {
	m.Logs = append(m.Logs, fmt.Sprintf("⚠ "+message, args...))
}

func (m *MockLogger) WarningColored(message string, args ...any) {
	m.Logs = append(m.Logs, fmt.Sprintf("[Colored] ⚠ "+message, args...))
}

func (m *MockLogger) Error(message string, args ...any) {
	m.Logs = append(m.Logs, fmt.Sprintf("✘ "+message, args...))
}

func (m *MockLogger) ErrorColored(message string, args ...any) {
	m.Logs = append(m.Logs, fmt.Sprintf("[Colored] ✘ "+message, args...))
}

func (m *MockLogger) Errorf(message string, args ...any) error {
	formattedMessage := fmt.Sprintf("✘ "+message, args...)
	m.Logs = append(m.Logs, formattedMessage)
	return errors.New(formattedMessage)
}

func (m *MockLogger) ErrorfColored(message string, args ...any) error {
	formattedMessage := fmt.Sprintf("[Colored] ✘ "+message, args...)
	m.Logs = append(m.Logs, formattedMessage)
	return errors.New(formattedMessage)
}

// SetIndent enables or disables message indentation.
func (m *MockLogger) SetIndent(enabled bool) {
	m.indentEnabled = enabled
}
