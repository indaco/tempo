package initcmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/testhelpers"
	"github.com/indaco/tempo/internal/testutils"
	"github.com/urfave/cli/v3"
)

func TestInitCommand(t *testing.T) {
	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: config.DefaultConfig(),
		CWD:    t.TempDir(),
	}
	tests := []struct {
		name             string
		baseFolder       string
		expectedConfig   bool
		expectedFilePath string
	}{
		{
			name:             "Default base folder",
			baseFolder:       ".tempo-files",
			expectedConfig:   true,
			expectedFilePath: ".tempo-files/tempo.yaml",
		},
		{
			name:             "Custom base folder",
			baseFolder:       "custom-tempo",
			expectedConfig:   true,
			expectedFilePath: "custom-tempo/tempo.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			baseFolder := filepath.Join(tempDir, tt.baseFolder)

			// Run the init command
			app := &cli.Command{}
			app.Commands = []*cli.Command{
				SetupInitCommand(cliCtx),
			}

			output, err := testhelpers.CaptureStdout(func() {
				args := []string{"tempo", "init", "--base-folder", baseFolder}
				err := app.Run(context.Background(), args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			})
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			testhelpers.ValidateCLIOutput(t, output, []string{
				"ℹ Generating",
				"✔ Done! Customize it to match your project needs.",
			})

			// Ensure config file created
			expectedFiles := []string{
				filepath.Join(tempDir, tt.expectedFilePath),
			}

			testutils.ValidateGeneratedFiles(t, expectedFiles)
		})
	}
}

func TestInitCommand_FailsOnExistingConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "tempo.yaml")
	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: config.DefaultConfig(),
		CWD:    t.TempDir(),
	}

	// Create an existing config file
	err := os.WriteFile(configFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	app := &cli.Command{}
	app.Commands = []*cli.Command{
		SetupInitCommand(cliCtx),
	}

	args := []string{"tempo", "init", "--base-folder", tempDir}
	err = app.Run(context.Background(), args)

	if err == nil {
		t.Fatal("Expected error, but got none")
	}

	expectedErr := "already exists"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Unexpected error message. Got: %s, Want substring: %s", err.Error(), expectedErr)
	}
}
