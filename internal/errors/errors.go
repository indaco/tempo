package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
)

/* ------------------------------------------------------------------------- */
/* INTERFACES                                                                */
/* ------------------------------------------------------------------------- */

// TempoErrorInterface defines the interface for TempoError with extended capabilities.
type TempoErrorInterface interface {
	error
	Code() int
	Attrs() map[string]any
}

/* ------------------------------------------------------------------------- */
/* STRUCTS                                                                   */
/* ------------------------------------------------------------------------- */

// TempoError represents an error with an additional context message, cause, and attributes.
type TempoError struct {
	Code    int            // Error code
	Message string         // Context message for the error
	Cause   error          // Underlying cause of the error, if any
	Attrs   map[string]any // Additional attributes for the error
}

/* ------------------------------------------------------------------------- */
/* CONSTRUCTOR FUNCTIONS                                                     */
/* ------------------------------------------------------------------------- */

// NewTempoError creates a new TempoError instance.
func NewTempoError(message string, cause error) *TempoError {
	return &TempoError{
		Message: message,
		Cause:   cause,
		Attrs:   make(map[string]any),
	}
}

func (e *TempoError) ToJSON() ([]byte, error) {
	if e == nil { // Check if the receiver is nil
		return nil, errors.New("cannot marshal nil TempoError")
	}

	var causeMessage string
	if e.Cause != nil {
		causeMessage = e.Cause.Error()
	}

	return json.Marshal(struct {
		Code    int            `json:"code"`
		Message string         `json:"message"`
		Cause   string         `json:"cause,omitempty"`
		Attrs   map[string]any `json:"attrs,omitempty"`
	}{
		Code:    e.Code,
		Message: e.Message,
		Cause:   causeMessage,
		Attrs:   e.Attrs,
	})
}

/* ------------------------------------------------------------------------- */
/* UTILITY METHODS                                                           */
/* ------------------------------------------------------------------------- */

// WithCode adds an error code to the TempoError and returns the updated error.
func (e *TempoError) WithCode(code int) *TempoError {
	e.Code = code
	return e
}

// WithAttr adds a key-value pair to the attributes and returns the updated error.
func (e *TempoError) WithAttr(key string, value any) *TempoError {
	e.Attrs[key] = value
	return e
}

/* ------------------------------------------------------------------------- */
/* INTERFACE IMPLEMENTATIONS                                                 */
/* ------------------------------------------------------------------------- */

// Error returns the context message of the TempoError.
func (e *TempoError) Error() string {
	return e.Message
}

// Unwrap returns the underlying cause of the error, if any.
func (e *TempoError) Unwrap() error {
	return e.Cause
}

/* ------------------------------------------------------------------------- */
/* WRAPPING FUNCTIONS                                                        */
/* ------------------------------------------------------------------------- */

// Wrap creates a TempoError with a given message, optional cause, and formatting arguments.
// If the first variadic argument is an error, it is treated as the cause.
func Wrap(msg string, args ...any) error {
	var cause error

	// Extract the first argument if it's an error (cause)
	if len(args) > 0 {
		if e, ok := args[0].(error); ok {
			cause = e
			args = args[1:] // Remove the cause from args
		}
	}

	// If no remaining arguments, return the raw message
	if len(args) == 0 {
		return NewTempoError(msg, cause)
	}

	// Apply faint color formatting to all arguments
	argColor := color.New(color.Faint).SprintFunc()
	coloredArgs := make([]string, len(args))
	for i, arg := range args {
		coloredArgs[i] = argColor(fmt.Sprint(arg)) // Ensure everything is a string
	}

	// Convert []string → []any for fmt.Sprintf compatibility
	coloredMsg := formatMessage(msg, stringToAnySlice(coloredArgs)...)

	return NewTempoError(coloredMsg, cause)
}

/* ------------------------------------------------------------------------- */
/* LOGGING FUNCTIONS                                                         */
/* ------------------------------------------------------------------------- */

// LogErrorChain logs an error chain to the console.
func LogErrorChain(err error) {
	if err == nil {
		return
	}

	output := color.Output
	errorColor := color.New(color.FgRed, color.Bold).SprintFunc()

	fmt.Fprintf(output, "%s\n", errorColor("✘ Something went wrong:"))
	for err != nil {
		fmt.Fprintf(output, "  %s %v\n", errorColor("→"), err)
		err = errors.Unwrap(err)
	}
}

// LogErrorChainWithAttrs logs an error chain with additional attributes in a structured format.
func LogErrorChainWithAttrs(err error) {
	output := color.Output
	errorColor := color.New(color.FgRed, color.Bold).SprintFunc()
	argColor := color.New(color.Faint).SprintFunc()

	fmt.Fprintf(output, "%s\n", errorColor("✘ Something went wrong:"))
	for err != nil {
		if tempoErr, ok := err.(*TempoError); ok {
			fmt.Fprintf(output, "  - Code: %d, Message: %s\n", tempoErr.Code, tempoErr.Message)
			if len(tempoErr.Attrs) > 0 {
				fmt.Fprintf(output, "    Attrs:\n")

				// Sort the metadata keys for consistent output
				keys := sortedKeys(tempoErr.Attrs)
				for _, key := range keys {
					fmt.Fprintf(output, "      %s: %v\n", argColor(key), tempoErr.Attrs[key])
				}
			}
		} else {
			fmt.Fprintf(output, "  - %v\n", err)
		}
		err = errors.Unwrap(err)
	}
}

/* ------------------------------------------------------------------------- */
/* HELPER FUNCTIONS                                                          */
/* ------------------------------------------------------------------------- */

// stringToAnySlice converts []string to []any
func stringToAnySlice(strs []string) []any {
	out := make([]any, len(strs))
	for i, s := range strs {
		out[i] = s
	}
	return out
}

// anyToStringSlice converts []any to []string
func anyToStringSlice(args []any) []string {
	out := make([]string, len(args))
	for i, a := range args {
		out[i] = fmt.Sprint(a)
	}
	return out
}

// formatMessage ensures safe message formatting
func formatMessage(msg string, args ...any) string {
	if strings.Contains(msg, "%") {
		return fmt.Sprintf(msg, args...)
	}
	return fmt.Sprint(msg, ": ", strings.Join(anyToStringSlice(args), " "))
}

// sortedKeys returns a sorted slice of keys from a map.
func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
