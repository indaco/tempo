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
	for _, tt := range []struct {
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
	} {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir() // Create a fresh test directory
			baseFolder := filepath.Join(tempDir, tt.baseFolder)

			// Create go.mod inside tempDir (the correct working directory)
			if err := testutils.CreateModFile(tempDir); err != nil {
				t.Fatalf("Failed to create go.mod file: %v", err)
			}

			// Set the working directory to tempDir
			cliCtx := &app.AppContext{
				Logger: logger.NewDefaultLogger(),
				Config: config.DefaultConfig(),
				CWD:    tempDir, // This ensures go.mod is found
			}

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

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: config.DefaultConfig(),
		CWD:    tempDir,
	}

	// Create an existing config file
	configFile := filepath.Join(tempDir, "tempo.yaml")
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

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

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
	configFile := filepath.Join(tempDir, "tempo.yaml")
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
	defer func() {
		if err := os.Chmod(configFile, 0644); err != nil {
			t.Errorf("Failed to restore permissions on %s: %v", configFile, err)
		}
	}()

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
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg, err := prepareConfig(tempDir, "tempo-root", "templates", "actions")
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

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

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

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
	configFile := filepath.Join(tempDir, "tempo.yaml")
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	// Ensure that DefaultTemplateExtensions were written
	for _, ext := range config.DefaultTemplateExtensions {
		expectedLine := fmt.Sprintf("    # - %s\n", ext)
		if !strings.Contains(string(content), expectedLine) {
			t.Errorf("Expected default template extension %s not found in config file", ext)
		}
	}
}

func TestWriteConfigFile_WithFunctionProviders(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	configFile := filepath.Join(tempDir, "tempo.yaml")

	cfg, err := prepareConfig(tempDir, "tempo-root", "templates", "actions")
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}

	cfg.Templates.FunctionProviders = []config.TemplateFuncProvider{
		{Name: "default", Type: "path", Value: "./providers/default"},
		{Name: "custom", Type: "url", Value: "https://github.com/user/custom-provider.git"},
	}

	err = writeConfigFile(configFile, cfg)
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

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

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

func TestInitCommand_FailsOnMissingGoMod(t *testing.T) {
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

	// Ensure go.mod does not exist
	goModPath := filepath.Join(tempDir, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		if err := os.Remove(goModPath); err != nil {
			t.Fatalf("Failed to remove existing go.mod file: %v", err)
		}
	}

	// Run `init`, expecting an error due to missing go.mod
	args := []string{"tempo", "init", "--base-folder", tempDir}
	err := app.Run(context.Background(), args)

	if err == nil {
		t.Fatal("Expected error due to missing go.mod file, but got none")
	}

	expectedErr := "missing go.mod file"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Unexpected error message. Got: %s, Want substring: %s", err.Error(), expectedErr)
	}
}

func TestValidateInitPrerequisites_FailsOnGoModStatError(t *testing.T) {
	tempDir := t.TempDir() // Create a fresh test directory

	// Step 1: Create a subdirectory for go.mod
	restrictedDir := filepath.Join(tempDir, "restricted")
	if err := os.Mkdir(restrictedDir, 0755); err != nil {
		t.Fatalf("Failed to create restricted directory: %v", err)
	}

	// Step 2: Create go.mod inside the directory BEFORE restricting access
	goModPath := filepath.Join(restrictedDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module example.com/myproject\n\ngo 1.23\n"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Step 3: Restrict access AFTER go.mod is created
	if err := os.Chmod(restrictedDir, 0000); err != nil {
		t.Fatalf("Failed to restrict directory permissions: %v", err)
	}
	defer func() {
		if err := os.Chmod(restrictedDir, 0755); err != nil {
			t.Errorf("Failed to restore permissions on %s: %v", restrictedDir, err)
		}
	}()

	// Step 4: Run validation
	err := validateInitPrerequisites(restrictedDir, filepath.Join(tempDir, "tempo.yaml"))

	// Step 5: Ensure we get the expected "error checking go.mod file" error
	if err == nil {
		t.Fatal("Expected error, but got none")
	}

	expectedErr := "error checking go.mod file"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Unexpected error message. Got: %s, Want substring: %s", err.Error(), expectedErr)
	}
}
