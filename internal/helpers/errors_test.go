package helpers

import (
	"strings"
	"testing"
)

func TestBuildMissingFoldersError(t *testing.T) {
	tests := []struct {
		name           string
		missingFolders map[string]string
		contextMsg     string
		helpCommands   []string
		expectedError  string
	}{
		{
			name: "Single Missing Folder",
			missingFolders: map[string]string{
				"Input directory": "./input",
			},
			contextMsg:   "Please ensure all required folders exist.",
			helpCommands: []string{"tempo component -h", "tempo variant -h"},
			expectedError: `oops! It looks like some required folders are missing.

Please ensure all required folders exist.

Missing folders:
  - Input directory: ./input

💡 Need help? Run:
  - tempo component -h
  - tempo variant -h`,
		},
		{
			name: "Multiple Missing Folders Without Help",
			missingFolders: map[string]string{
				"Input directory":  "./input",
				"Output directory": "./output",
			},
			contextMsg:   "Run setup to create missing folders.",
			helpCommands: nil,
			expectedError: `oops! It looks like some required folders are missing.

Run setup to create missing folders.

Missing folders:
  - Input directory: ./input
  - Output directory: ./output`,
		},
		{
			name:           "No Missing Folders",
			missingFolders: map[string]string{},
			contextMsg:     "This message should not appear.",
			helpCommands:   []string{"tempo setup"},
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := BuildMissingFoldersError(tt.missingFolders, tt.contextMsg, tt.helpCommands)

			// If no folders are missing, err should be nil
			if len(tt.missingFolders) == 0 {
				if err != nil {
					t.Fatalf("Expected nil error but got: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("Expected an error but got nil")
			}

			actualError := err.Error()

			// Normalize line endings for cross-platform testing
			expected := strings.ReplaceAll(tt.expectedError, "\r\n", "\n")
			actual := strings.ReplaceAll(actualError, "\r\n", "\n")

			if actual != expected {
				t.Errorf("Expected error message:\n%q\nGot:\n%q", expected, actual)
			}
		})
	}
}
