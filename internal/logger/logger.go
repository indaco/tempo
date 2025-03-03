package logger

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

/* ------------------------------------------------------------------------- */
/* INTERFACES                                                                */
/* ------------------------------------------------------------------------- */

// LoggerInterface defines the interface for structured logging.
type LoggerInterface interface {
	Default(message string, args ...any) *LogEntry
	Info(message string, args ...any) *LogEntry
	Success(message string, args ...any) *LogEntry
	Warning(message string, args ...any) *LogEntry
	Error(message string, args ...any) *LogEntry
	Hint(message string, args ...any) *LogEntry
	Blank()
	WithTimestamp(enabled bool)
	WithIndent(enabled bool)
	Reset()
}

/* ------------------------------------------------------------------------- */
/* TYPES & VARS                                                              */
/* ------------------------------------------------------------------------- */

// LogEntry represents a single log message entry.
type LogEntry struct {
	level     string         // Log level (info, success, warning, error)
	icon      string         // Icon associated with the log level
	message   string         // Main log message
	attrs     []KeyValue     // Attributes stored in insertion order
	timestamp *time.Time     // Optional timestamp for the log entry
	logger    *DefaultLogger // Reference to the logger instance
	mu        sync.Mutex     // Mutex for concurrent attribute updates
}

// KeyValue represents a key-value pair for attributes.
type KeyValue struct {
	Key   string
	Value any
}

// DefaultLogger is the default implementation of the LoggerInterface.
type DefaultLogger struct {
	indentEnabled    bool
	timestampEnabled bool
	mu               sync.Mutex
}

// levels holds the log levels and their associated icons.
var levels = map[string]string{
	"default": "",
	"info":    "ℹ",
	"success": "✔",
	"warning": "⚠",
	"error":   "✘",
	"hint":    "💡",
}

// Define styles for log levels
var styleFuncs = map[string]func(string) string{
	"hint":    styleWrapper(color.New(color.Faint).Sprint), // Faint text for hints
	"default": func(s string) string { return s },          // No styling for default
}

// Define color functions for levels
var colorFuncs = map[string]func(string) string{
	"info":    styleWrapper(color.New(color.FgBlue, color.Bold).Sprint),
	"success": styleWrapper(color.New(color.FgGreen, color.Bold).Sprint),
	"warning": styleWrapper(color.New(color.FgYellow, color.Bold).Sprint),
	"error":   styleWrapper(color.New(color.FgRed, color.Bold).Sprint),
}

/* ------------------------------------------------------------------------- */
/* PUBLIC METHODS                                                            */
/* ------------------------------------------------------------------------- */

// NewDefaultLogger creates and returns a new instance of DefaultLogger.
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{}
}

// Reset resets the DefaultLogger to its default state.
func (l *DefaultLogger) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.indentEnabled = false
	l.timestampEnabled = false
}

// Default creates a default-level log entry.
func (l *DefaultLogger) Default(message string, args ...any) *LogEntry {
	return l.createLogEntry("default", message, args...)
}

// Info creates an info-level log entry.
func (l *DefaultLogger) Info(message string, args ...any) *LogEntry {
	return l.createLogEntry("info", message, args...)
}

// Success creates a success-level log entry.
func (l *DefaultLogger) Success(message string, args ...any) *LogEntry {
	return l.createLogEntry("success", message, args...)
}

// Warning creates a warning-level log entry.
func (l *DefaultLogger) Warning(message string, args ...any) *LogEntry {
	return l.createLogEntry("warning", message, args...)
}

// Error creates an error-level log entry.
func (l *DefaultLogger) Error(message string, args ...any) *LogEntry {
	return l.createLogEntry("error", message, args...)
}

// Hint creates an hint-level log entry.
func (l *DefaultLogger) Hint(message string, args ...any) *LogEntry {
	return l.createLogEntry("hint", message, args...)
}

// Blank prints a blank line
func (l *DefaultLogger) Blank() {
	fmt.Fprintln(color.Output)
}

// WithIndent enables or disables message indentation.
func (l *DefaultLogger) WithIndent(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.indentEnabled = enabled
}

// WithTimestamp enables or disables timestamps for log entries.
func (l *DefaultLogger) WithTimestamp(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.timestampEnabled = enabled
}

/* ------------------------------------------------------------------------- */
/* HELPER METHODS                                                            */
/* ------------------------------------------------------------------------- */

// createLogEntry initializes a new LogEntry with the given level and message.
func (l *DefaultLogger) createLogEntry(level, message string, args ...any) *LogEntry {
	icon, ok := levels[level]
	if !ok {
		icon = "?" // Default icon for unknown levels
	}

	// Get the style function based on log level
	styleFunc, exists := styleFuncs[level]
	if !exists {
		styleFunc = styleWrapper(color.New(color.Bold).Sprint) // Default to bold for other levels
	}

	formattedMessage := styleFunc(message)

	// White (non-bold) formatting for the arguments
	plain := styleWrapper(color.New(color.FgHiWhite).Sprint)
	if len(args) > 0 {
		// Ensure args are properly formatted and appended to message
		formattedArgs := []string{}
		for i := range args {
			formattedArgs = append(formattedArgs, fmt.Sprint(args[i]))
		}
		formattedMessage += " " + plain(strings.Join(formattedArgs, " "))
	}

	entry := &LogEntry{
		level:   level,
		icon:    icon,
		message: formattedMessage,
		attrs:   []KeyValue{},
		logger:  l,
	}

	if l.timestampEnabled {
		now := time.Now()
		entry.timestamp = &now
	}

	entry.log()
	return entry
}

// styleWrapper wraps a `Sprint` function to match the `func(string) string` signature.
func styleWrapper(sprintFunc func(a ...any) string) func(string) string {
	return func(input string) string {
		return sprintFunc(input)
	}
}

// indentEnabled checks if indentation is enabled for the log entry.
func (e *LogEntry) indentEnabled() bool {
	return e.logger != nil && e.logger.indentEnabled
}

// log prints the log entry to the console.
func (e *LogEntry) log() {
	output := color.Output

	// Get the style function based on log level
	style, ok := colorFuncs[e.level]
	if !ok {
		style = styleWrapper(color.New(color.FgWhite).Sprint)
	}

	var sb strings.Builder

	// Add timestamp if available
	if e.timestamp != nil {
		sb.WriteString(e.timestamp.Format("2006-01-02 15:04:05") + " ")
	}

	// Add styled icon
	if e.icon != "" {
		sb.WriteString(style(e.icon) + " ")
	}

	// Add the main message
	sb.WriteString(e.message)

	// Apply indentation if enabled
	message := sb.String()
	if e.indentEnabled() {
		message = "  " + strings.ReplaceAll(message, "\n", "\n  ")
	}

	// Print the message
	fmt.Fprintln(output, message)
}

// WithAttrs adds attributes to the log entry and prints the entry.
func (e *LogEntry) WithAttrs(attrs ...any) *LogEntry {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(attrs)%2 != 0 {
		fmt.Println("Invalid attributes: must be key-value pairs")
		return e
	}

	newAttrs := make([]KeyValue, 0, len(attrs)/2)
	for i := 0; i < len(attrs); i += 2 {
		key, ok := attrs[i].(string)
		if !ok {
			fmt.Println("Invalid attribute key: must be a string")
			continue
		}
		newAttrs = append(newAttrs, KeyValue{Key: key, Value: attrs[i+1]})
	}

	e.attrs = append(e.attrs, newAttrs...)
	e.logAttrs()
	return e
}

// logAttrs logs the attributes in a structured format.
func (e *LogEntry) logAttrs() {
	if len(e.attrs) == 0 {
		return
	}

	output := color.Output
	argColor := color.New(color.Faint).SprintFunc()

	for _, attr := range e.attrs {
		line := fmt.Sprintf("  - %s: %v", argColor(attr.Key), attr.Value)
		if e.indentEnabled() {
			line = "  " + line
		}
		fmt.Fprintln(output, line)
	}
}

// ToJSON converts the log entry into a JSON string.
func (e *LogEntry) ToJSON() (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	data := map[string]any{
		"level":      e.level,
		"message":    e.message,
		"icon":       e.icon,
		"attributes": e.attrs,
	}

	if e.timestamp != nil {
		data["timestamp"] = e.timestamp.Format(time.RFC3339)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
