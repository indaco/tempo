package utils

import (
	"testing"

	"github.com/indaco/tempo/internal/testutils"
)

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
