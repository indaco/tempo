package apperrors

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/fatih/color"
)

func TestErrorChain(t *testing.T) {
	baseErr := errors.New("template execution failed")
	wrappedErr := Wrap("failed to render file", baseErr)
	finalErr := Wrap("component creation failed", wrappedErr)

	// Test the error message
	if finalErr.Error() != "component creation failed" {
		t.Errorf("Unexpected error message: got %q, want %q", finalErr.Error(), "component creation failed")
	}

	// Test the cause of the error
	cause := errors.Unwrap(finalErr)
	if cause.Error() != "failed to render file" {
		t.Errorf("Unexpected cause: got %q, want %q", cause.Error(), "failed to render file")
	}

	// Test the base error
	base := errors.Unwrap(cause)
	if base.Error() != "template execution failed" {
		t.Errorf("Unexpected base error: got %q, want %q", base.Error(), "template execution failed")
	}
}

func TestWrap(t *testing.T) {
	baseErr := errors.New("an error occurred")
	wrappedErr := Wrap("something went wrong", baseErr)

	if wrappedErr.Error() != "something went wrong" {
		t.Errorf("Unexpected error message: got %q, want %q", wrappedErr.Error(), "something went wrong")
	}

	if errors.Unwrap(wrappedErr) != baseErr {
		t.Errorf("Unexpected cause: got %v, want %v", errors.Unwrap(wrappedErr), baseErr)
	}
}

func TestWrapWithFormatting(t *testing.T) {
	baseErr := errors.New("an error occurred")
	wrappedErr := Wrap("failed to process '%s'", baseErr, "file.txt")

	if wrappedErr.Error() != "failed to process 'file.txt'" {
		t.Errorf("Unexpected error message: got %q, want %q", wrappedErr.Error(), "failed to process 'file.txt'")
	}

	if errors.Unwrap(wrappedErr) != baseErr {
		t.Errorf("Unexpected cause: got %v, want %v", errors.Unwrap(wrappedErr), baseErr)
	}
}

func TestLogErrorChain(t *testing.T) {
	baseErr := errors.New("template execution failed")
	wrappedErr := Wrap("failed to render file", baseErr)
	finalErr := Wrap("component creation failed", wrappedErr)

	var buf bytes.Buffer
	color.Output = &buf // Redirect color output to the buffer

	LogErrorChain(finalErr)

	got := buf.String()
	want := `X Something went wrong:
  -> component creation failed
  -> failed to render file
  -> template execution failed
`

	if got != want {
		t.Errorf("Unexpected log output:\nGot:\n%s\nWant:\n%s", got, want)
	}
}

func TestLogErrorChainNil(t *testing.T) {

	var buf bytes.Buffer
	color.Output = &buf // Redirect color output to the buffer

	LogErrorChain(nil)

	got := buf.String()
	want := "" // Expect no output

	if got != want {
		t.Errorf("Unexpected log output for nil error:\nGot:\n%s\nWant:\n%s", got, want)
	}
}

func TestNewTempoError(t *testing.T) {
	baseErr := errors.New("an error occurred")
	customErr := NewTempoError("failed operation", baseErr)

	if customErr.Message != "failed operation" {
		t.Errorf("Unexpected message: got %q, want %q", customErr.Message, "failed operation")
	}

	if customErr.Cause != baseErr {
		t.Errorf("Unexpected cause: got %v, want %v", customErr.Cause, baseErr)
	}
}

// TestWithCode has been removed as the Code field is no longer part of TempoError

func TestWithAttrs(t *testing.T) {
	err := NewTempoError("operation failed", nil)
	err.Attrs = make(map[string]any)
	err = err.WithAttr("key1", "value1")
	err = err.WithAttr("key2", 42)

	if err.Attrs["key1"] != "value1" {
		t.Errorf("Unexpected Attrs for key1: got %v, want %v", err.Attrs["key1"], "value1")
	}

	if err.Attrs["key2"] != 42 {
		t.Errorf("Unexpected Attrs for key2: got %v, want %v", err.Attrs["key2"], 42)
	}
}

func TestLogErrorChainWithMetadata(t *testing.T) {
	baseErr := NewTempoError("file not found", nil)

	metadataErr := NewTempoError("operation failed", baseErr).
		WithAttr("key1", "value1").
		WithAttr("key2", 42)

	var buf bytes.Buffer
	color.Output = &buf // Redirect color output to the buffer

	LogErrorChainWithAttrs(metadataErr)

	// Explicitly capture the output from the buffer
	got := buf.String()
	want := `X Something went wrong:
  - Message: operation failed
    Attrs:
      key1: value1
      key2: 42
  - Message: file not found
`

	if got != want {
		t.Errorf("Unexpected log output:\nGot:\n%s\nWant:\n%s", got, want)
	}
}

func TestTempoErrorToJSON(t *testing.T) {
	tests := []struct {
		name       string
		tempoError *TempoError
		expected   string
		expectErr  bool
	}{
		{
			name: "Valid TempoError with cause and attributes",
			tempoError: &TempoError{
				Message: "Internal Server Error",
				Cause:   errors.New("database connection failed"),
				Attrs: map[string]any{
					"retryable": true,
					"endpoint":  "/api/v1/resource",
				},
			},
			expected:  `{"message":"Internal Server Error","cause":"database connection failed","attrs":{"endpoint":"/api/v1/resource","retryable":true}}`,
			expectErr: false,
		},
		{
			name: "Valid TempoError without cause",
			tempoError: &TempoError{
				Message: "Not Found",
				Attrs: map[string]any{
					"resource": "user",
				},
			},
			expected:  `{"message":"Not Found","attrs":{"resource":"user"}}`,
			expectErr: false,
		},
		{
			name: "Valid TempoError with nil attributes",
			tempoError: &TempoError{
				Message: "Bad Request",
				Cause:   errors.New("invalid input"),
			},
			expected:  `{"message":"Bad Request","cause":"invalid input"}`,
			expectErr: false,
		},
		{
			name:       "Nil TempoError",
			tempoError: nil,
			expected:   "",
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the ToJSON method
			output, err := tt.tempoError.ToJSON()

			// Validate error expectation
			if (err != nil) != tt.expectErr {
				t.Fatalf("Unexpected error state: got %v, want error: %v", err, tt.expectErr)
			}

			// Validate JSON output if no error is expected
			if !tt.expectErr {
				var got map[string]any
				var expected map[string]any

				if err := json.Unmarshal(output, &got); err != nil {
					t.Fatalf("Failed to unmarshal output JSON: %v", err)
				}
				if err := json.Unmarshal([]byte(tt.expected), &expected); err != nil {
					t.Fatalf("Failed to unmarshal expected JSON: %v", err)
				}

				// Use reflect.DeepEqual to compare the maps
				if !reflect.DeepEqual(got, expected) {
					t.Errorf("Unexpected JSON output:\nGot: %s\nWant: %s", string(output), tt.expected)
				}
			}
		})

		t.Run("Nil TempoError", func(t *testing.T) {
			var err *TempoError = nil
			_, jsonErr := err.ToJSON()
			if jsonErr == nil || jsonErr.Error() != "cannot marshal nil TempoError" {
				t.Errorf("Expected error for nil TempoError, got: %v", jsonErr)
			}
		})
	}
}

func TestStringToAnySlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []any
	}{
		{
			name:     "Convert non-empty string slice",
			input:    []string{"one", "two", "three"},
			expected: []any{"one", "two", "three"},
		},
		{
			name:     "Convert single-element slice",
			input:    []string{"hello"},
			expected: []any{"hello"},
		},
		{
			name:     "Convert empty slice",
			input:    []string{},
			expected: []any{},
		},
		{
			name:     "Convert slice with special characters",
			input:    []string{"$#@!", "hello", "world"},
			expected: []any{"$#@!", "hello", "world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringToAnySlice(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestAnyToStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []any
		expected []string
	}{
		{
			name:     "Convert mixed types",
			input:    []any{"one", 2, 3.14, true},
			expected: []string{"one", "2", "3.14", "true"},
		},
		{
			name:     "Convert single-element slice",
			input:    []any{42},
			expected: []string{"42"},
		},
		{
			name:     "Convert empty slice",
			input:    []any{},
			expected: []string{},
		},
		{
			name:     "Convert slice with special characters",
			input:    []any{"$#@!", "hello", "world"},
			expected: []string{"$#@!", "hello", "world"},
		},
		{
			name:     "Convert slice with nil values",
			input:    []any{nil, "test", nil},
			expected: []string{"<nil>", "test", "<nil>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := anyToStringSlice(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestErrorsIs validates that errors.Is works correctly with TempoError chain.
func TestErrorsIs(t *testing.T) {
	// Create a base sentinel error
	baseErr := errors.New("base error")

	// Wrap it with TempoError
	wrappedErr := Wrap("wrapped error", baseErr)

	// Double wrap
	doubleWrapped := Wrap("double wrapped", wrappedErr)

	// Test errors.Is finds the base error in the chain
	if !errors.Is(wrappedErr, baseErr) {
		t.Error("errors.Is should find baseErr in wrappedErr")
	}

	if !errors.Is(doubleWrapped, baseErr) {
		t.Error("errors.Is should find baseErr in doubleWrapped")
	}

	// Test that unrelated error is not found
	unrelatedErr := errors.New("unrelated")
	if errors.Is(doubleWrapped, unrelatedErr) {
		t.Error("errors.Is should not find unrelatedErr in doubleWrapped")
	}
}

// TestErrorsAs validates that errors.As works correctly with TempoError chain.
func TestErrorsAs(t *testing.T) {
	// Create a TempoError
	tempoErr := NewTempoError("tempo error", nil)

	// Wrap it with another TempoError
	wrappedErr := Wrap("wrapped error", tempoErr)

	// Test errors.As finds the inner TempoError
	var target *TempoError
	if !errors.As(wrappedErr, &target) {
		t.Error("errors.As should find TempoError in wrappedErr")
	}

	// The first match should be the outer error
	if target.Message != "wrapped error" {
		t.Errorf("Expected 'wrapped error', got %q", target.Message)
	}

	// Test with standard error wrapped inside
	stdErr := errors.New("standard error")
	wrappedStd := Wrap("wrapped std", stdErr)

	var tempoTarget *TempoError
	if !errors.As(wrappedStd, &tempoTarget) {
		t.Error("errors.As should find TempoError wrapper")
	}

	// Verify Unwrap chain works
	unwrapped := errors.Unwrap(tempoTarget)
	if unwrapped != stdErr {
		t.Errorf("Expected unwrapped error to be stdErr, got %v", unwrapped)
	}
}
