package app

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/tempo/internal/utils"
)

func TestIsTempoProject(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(tempDir string)
		overrideFunc  func()
		expectedError string
	}{
		{
			name: "Valid tempo.yaml exists with go.mod",
			setupFunc: func(tempDir string) {
				// Create tempo.yaml
				filePath := filepath.Join(tempDir, "tempo.yaml")
				if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
					t.Fatalf("failed to create tempo.yaml: %v", err)
				}

				// Create go.mod
				goModPath := filepath.Join(tempDir, "go.mod")
				if err := os.WriteFile(goModPath, []byte("module test"), 0644); err != nil {
					t.Fatalf("failed to create go.mod: %v", err)
				}
			},
			expectedError: "",
		},
		{
			name: "Valid tempo.yml exists with go.mod",
			setupFunc: func(tempDir string) {
				// Create tempo.yml
				filePath := filepath.Join(tempDir, "tempo.yml")
				if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
					t.Fatalf("failed to create tempo.yml: %v", err)
				}

				// Create go.mod
				goModPath := filepath.Join(tempDir, "go.mod")
				if err := os.WriteFile(goModPath, []byte("module test"), 0644); err != nil {
					t.Fatalf("failed to create go.mod: %v", err)
				}
			},
			expectedError: "",
		},
		{
			name: "No config file exists but go.mod is present",
			setupFunc: func(tempDir string) {
				// Create go.mod
				goModPath := filepath.Join(tempDir, "go.mod")
				if err := os.WriteFile(goModPath, []byte("module test"), 0644); err != nil {
					t.Fatalf("failed to create go.mod: %v", err)
				}
			},
			expectedError: "no config file found; checked: [tempo.yaml tempo.yml]. Run 'tempo init' first",
		},
		{
			name: "No go.mod file",
			setupFunc: func(tempDir string) {
				// Do not create go.mod file
			},
			expectedError: "missing go.mod file. Run 'go mod init' to create one",
		},
		{
			name: "Error when checking for file existence",
			setupFunc: func(tempDir string) {
				// Create go.mod
				goModPath := filepath.Join(tempDir, "go.mod")
				if err := os.WriteFile(goModPath, []byte("module test"), 0644); err != nil {
					t.Fatalf("failed to create go.mod: %v", err)
				}
			},
			overrideFunc: func() {
				// Override the FileOrDirExistsFunc function to simulate an error.
				utils.FileOrDirExistsFunc = func(path string) (bool, bool, error) {
					return false, false, errors.New("mocked error")
				}
			},
			expectedError: "error checking config file",
		},
	}

	// Save the original function and restore it after the tests.
	origFunc := utils.FileOrDirExistsFunc
	defer func() {
		utils.FileOrDirExistsFunc = origFunc
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for this test.
			tempDir := t.TempDir()
			if tt.setupFunc != nil {
				tt.setupFunc(tempDir)
			}
			// If an override is provided, store the original function and restore after the test.
			var origFunc func(string) (bool, bool, error)
			if tt.overrideFunc != nil {
				origFunc = utils.FileOrDirExistsFunc
				tt.overrideFunc()
				defer func() {
					utils.FileOrDirExistsFunc = origFunc
				}()
			}
			err := IsTempoProject(tempDir)
			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("expected nil error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.expectedError)
				} else if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got: %v", tt.expectedError, err)
				}
			}
		})
	}
}
