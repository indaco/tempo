package helpers

import (
	"strings"
	"testing"
)

func TestBuildMissingFoldersError(t *testing.T) {
	tests := []struct {
		name           string
		missingFolders []string
		contextMsg     string
		helpCommands   []string
		expectedError  string
	}{
		{
			name:           "Single Missing Folder",
			missingFolders: []string{"  - Input directory: ./input"},
			contextMsg:     "Please ensure all required folders exist.",
			helpCommands:   []string{"tempo define -h", "tempo create -h"},
			expectedError: `oops! It looks like some required folders are missing.

Please ensure all required folders exist.

Missing folders:
  - Input directory: ./input

💡 Need help? Run:
  - tempo define -h
  - tempo create -h`,
		},
		{
			name:           "Multiple Missing Folders Without Help",
			missingFolders: []string{"  - Input directory: ./input", "  - Output directory: ./output"},
			contextMsg:     "Run setup to create missing folders.",
			helpCommands:   nil,
			expectedError: `oops! It looks like some required folders are missing.

Run setup to create missing folders.

Missing folders:
  - Input directory: ./input
  - Output directory: ./output`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := BuildMissingFoldersError(tt.missingFolders, tt.contextMsg, tt.helpCommands)
			if err == nil {
				t.Fatalf("Expected an error but got nil")
			}

			actualError := err.Error()

			// Normalize line endings for cross-platform testing
			expected := strings.ReplaceAll(tt.expectedError, "\r\n", "\n")
			actual := strings.ReplaceAll(actualError, "\r\n", "\n")

			if actual != expected {
				t.Errorf("Expected error message:\n%s\nGot:\n%s", expected, actual)
			}
		})
	}
}
