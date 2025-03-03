package cleancmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/testutils"
	"github.com/urfave/cli/v3"
)

func TestCleanCommand(t *testing.T) {
	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
	}
	tests := []struct {
		name             string
		prepareFile      bool
		prepareAsDir     bool
		expectError      bool
		expectedErrorMsg string
		expectedLogs     []string
	}{
		{
			name:         "File exists and is removed",
			prepareFile:  true,
			prepareAsDir: false,
			expectError:  false,
			expectedLogs: []string{"✔ Successfully removed .tempo-lastrun"},
		},
		{
			name:         "File does not exist",
			prepareFile:  false,
			expectError:  false,
			expectedLogs: []string{"ℹ File does not exist, nothing to clean"},
		},
		{
			name:             "Path exists but is a directory",
			prepareFile:      true,
			prepareAsDir:     true,
			expectError:      true,
			expectedErrorMsg: "exists but is a directory, not a file",
			expectedLogs:     nil, // No logs are expected because an error occurs first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Change directory to tempDir
			if err := os.Chdir(tempDir); err != nil {
				t.Fatalf("Failed to change directory to tempDir: %v", err)
			}

			filePath := filepath.Join(tempDir, ".tempo-lastrun")

			// Prepare the file or directory if needed
			if tt.prepareFile {
				if tt.prepareAsDir {
					// Create a directory instead of a file
					if err := os.Mkdir(filePath, 0755); err != nil {
						t.Fatalf("Failed to prepare test directory: %v", err)
					}
				} else {
					// Create a file
					if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
						t.Fatalf("Failed to prepare test file: %v", err)
					}
				}
			}

			// Capture CLI output
			output, err := testutils.CaptureStdout(func() {
				app := &cli.Command{}
				app.Commands = []*cli.Command{
					SetupCleanCommand(cliCtx),
				}
				args := []string{"tempo", "clean"}
				_ = app.Run(context.Background(), args)
			})

			// If an error was expected, validate the error message
			if tt.expectError && err != nil {
				if !strings.Contains(err.Error(), tt.expectedErrorMsg) {
					t.Errorf("Unexpected error message: got %q, expected to contain %q", err.Error(), tt.expectedErrorMsg)
				}
			}

			// Validate CLI output
			if !tt.expectError {
				testutils.ValidateCLIOutput(t, output, tt.expectedLogs)
			}

			// Validate file removal
			if !tt.prepareAsDir && tt.prepareFile && !tt.expectError {
				// Ensure the file was removed
				if _, err := os.Stat(filePath); !os.IsNotExist(err) {
					t.Errorf("Expected file to be removed, but it still exists")
				}
			}
		})
	}
}
