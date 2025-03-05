package helpers

import (
	"strings"
	"testing"

	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/testhelpers"
)

func TestEnableLoggerIndentation(t *testing.T) {
	// Create a new logger.
	log := logger.NewDefaultLogger()

	// By default, indentation is off. Call our helper to enable it.
	EnableLoggerIndentation(log)

	// Capture the output of a log message.
	output, err := testhelpers.CaptureStdout(func() {
		log.Success("Test message")
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// The logger prints the message using Fprintln, so the output should end with a newline.
	// With indentation enabled, the output should start with two spaces.
	if !strings.HasPrefix(output, "  ") {
		t.Errorf("Expected output to be indented, got: %q", output)
	}
}

func TestResetLogger(t *testing.T) {
	// Create a new logger.
	log := logger.NewDefaultLogger()

	// Enable indentation.
	EnableLoggerIndentation(log)
	outputIndented, err := testhelpers.CaptureStdout(func() {
		log.Success("Indented message")
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}
	if !strings.HasPrefix(outputIndented, "  ") {
		t.Errorf("Expected indented output, got: %q", outputIndented)
	}

	// Now reset the logger.
	ResetLogger(log)
	outputReset, err := testhelpers.CaptureStdout(func() {
		log.Success("Non indented message")
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// After reset, the output should not have the extra indentation.
	if strings.HasPrefix(outputReset, "  ") {
		t.Errorf("Expected output not to be indented after reset, got: %q", outputReset)
	}
}
