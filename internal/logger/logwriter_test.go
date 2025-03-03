package logger

import (
	"testing"

	"github.com/indaco/tempo/testutils"
)

func TestLogWriter(t *testing.T) {
	tests := []struct {
		name           string
		level          string
		input          string
		expectedOutput string
	}{
		{
			name:           "Info Level",
			level:          "info",
			input:          "This is an info message",
			expectedOutput: "ℹ This is an info message\n", // ℹ is the info icon
		},
		{
			name:           "Success Level",
			level:          "success",
			input:          "Operation completed successfully",
			expectedOutput: "✔ Operation completed successfully\n", // ✔ is the success icon
		},
		{
			name:           "Warning Level",
			level:          "warn",
			input:          "This is a warning",
			expectedOutput: "⚠ This is a warning\n", // ⚠ is the warning icon
		},
		{
			name:           "Error Level",
			level:          "error",
			input:          "An error occurred",
			expectedOutput: "✘ An error occurred\n", // ✘ is the error icon
		},
		{
			name:           "Default Level",
			level:          "default",
			input:          "This is a default message",
			expectedOutput: "This is a default message\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &LogWriter{Level: tt.level, Logger: NewDefaultLogger()}
			output, err := testutils.CaptureStdout(func() {
				_, err := writer.Write([]byte(tt.input))
				if err != nil {
					t.Fatalf("Failed to write on LogWriter: %v", err)
				}
			})
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			if output != tt.expectedOutput {
				t.Errorf("Unexpected output for level %s:\nGot: %q\nWant: %q", tt.level, output, tt.expectedOutput)
			}
		})
	}
}
