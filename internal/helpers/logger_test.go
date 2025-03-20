package helpers

import (
	"strings"
	"testing"

	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/testhelpers"
	"github.com/indaco/tempo/internal/testutils"
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

func TestLogSuccessMessages(t *testing.T) {
	tests := []struct {
		name       string
		entityType string
		expected   string
	}{
		{
			name:       "Component",
			entityType: "component",
			expected:   "✔ Templates for the component and assets (CSS and JS) have been created",
		},
		{
			name:       "Component Variant",
			entityType: "component-variant",
			expected:   "✔ Templates for the component variant and assets (CSS) have been created",
		},
		{
			name:       "Default Case",
			entityType: "unknown",
			expected:   "✔ Templates and assets have been created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &testutils.MockLogger{}
			cfg := testutils.SetupConfig(t.TempDir(), nil)

			LogSuccessMessages(tt.entityType, cfg, mockLogger)

			if mockLogger.Logs[0] != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, mockLogger.Logs[0])
			}
		})
	}
}
