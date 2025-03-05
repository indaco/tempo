package initcmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/testhelpers"
	"github.com/indaco/tempo/internal/testutils"
	"github.com/indaco/tempo/internal/utils"
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

func TestInitCommand_FailsOnUnwritableConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "tempo.yaml")

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: config.DefaultConfig(),
		CWD:    tempDir,
	}

	app := &cli.Command{}
	app.Commands = []*cli.Command{
		SetupInitCommand(cliCtx),
	}

	// Step 1: Ensure the config file does NOT exist
	if _, err := os.Stat(configFile); err == nil {
		if err := os.Remove(configFile); err != nil {
			t.Fatalf("Failed to remove existing config file: %v", err)
		}
	}

	// Step 2: Create an unwritable file manually (simulating write failure)
	file, err := os.Create(configFile)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
	file.Close()

	// Make the file unwritable
	if err := os.Chmod(configFile, 0444); err != nil {
		t.Fatalf("Failed to set read-only permissions on config file: %v", err)
	}
	defer os.Chmod(configFile, 0644) // Ensure cleanup after test

	// Step 3: Run `init` again, expecting a write error
	args := []string{"tempo", "init", "--base-folder", tempDir}
	err = app.Run(context.Background(), args)

	if err == nil {
		t.Fatal("Expected error due to unwritable file, but got none")
	}

	expectedErr := "Failed to write the configuration file"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Unexpected error message. Got: %s, Want substring: %s", err.Error(), expectedErr)
	}
}

func TestPrepareConfig_DefaultExtensions(t *testing.T) {
	cfg := prepareConfig("tempo-root", "templates", "actions")

	if len(cfg.Templates.Extensions) == 0 {
		t.Fatalf("Expected extensions to have a default value, got empty slice")
	}

	expectedExtensions := config.DefaultTemplateExtensions
	if len(cfg.Templates.Extensions) != len(expectedExtensions) {
		t.Errorf("Mismatch in default extensions count: got %d, want %d", len(cfg.Templates.Extensions), len(expectedExtensions))
	}
}

func TestInitCommand_UsesDefaultTemplateExtensions(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "tempo.yaml")

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: config.DefaultConfig(),
		CWD:    tempDir,
	}

	// Override config to simulate missing extensions
	cliCtx.Config.Templates.Extensions = []string{} // Force empty slice

	app := &cli.Command{}
	app.Commands = []*cli.Command{
		SetupInitCommand(cliCtx),
	}

	args := []string{"tempo", "init", "--base-folder", tempDir}
	err := app.Run(context.Background(), args)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Read the generated config file
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Ensure that DefaultTemplateExtensions were written
	for _, ext := range config.DefaultTemplateExtensions {
		expectedLine := fmt.Sprintf("    - %s\n", ext)
		if !strings.Contains(string(content), expectedLine) {
			t.Errorf("Expected default template extension %s not found in config file", ext)
		}
	}
}

func TestWriteConfigFile_WithFunctionProviders(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "tempo.yaml")

	cfg := prepareConfig("tempo-root", "templates", "actions")
	cfg.Templates.FunctionProviders = []config.TemplateFuncProvider{
		{Name: "default", Type: "path", Value: "./providers/default"},
		{Name: "custom", Type: "url", Value: "https://github.com/user/custom-provider.git"},
	}

	err := writeConfigFile(configFile, cfg)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Read file content
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}
	configContent := string(content)

	// Ensure function providers were correctly written
	if !strings.Contains(configContent, "name: default") ||
		!strings.Contains(configContent, "value: ./providers/default") ||
		!strings.Contains(configContent, "name: custom") ||
		!strings.Contains(configContent, "value: https://github.com/user/custom-provider.git") {
		t.Errorf("Function providers were not written correctly in the config file:\n%s", configContent)
	}
}

func TestInitCommand_FailsOnConfigFileCheckError(t *testing.T) {
	tempDir := t.TempDir()

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: config.DefaultConfig(),
		CWD:    tempDir,
	}

	app := &cli.Command{}
	app.Commands = []*cli.Command{
		SetupInitCommand(cliCtx),
	}

	// Step 1: Mock `utils.FileExists` to always return an error
	utils.FileExistsFunc = func(_ string) (bool, error) {
		return false, fmt.Errorf("simulated file system error")
	}
	defer func() { utils.FileExistsFunc = utils.FileExists }() // Restore after test

	// Step 2: Run `init`, expecting an error
	args := []string{"tempo", "init", "--base-folder", tempDir}
	err := app.Run(context.Background(), args)

	if err == nil {
		t.Fatal("Expected error due to file existence check failure, but got none")
	}

	expectedErr := "Error checking configuration file"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Unexpected error message. Got: %s, Want substring: %s", err.Error(), expectedErr)
	}
}
