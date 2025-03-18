package componentcmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/templatefuncs/providers/gonameprovider"
	"github.com/indaco/tempo/internal/testhelpers"
	"github.com/indaco/tempo/internal/testutils"
	"github.com/indaco/tempo/internal/utils"
	"github.com/urfave/cli/v3"
)

func setupConfig(tempDir string, overrides func(cfg *config.Config)) *config.Config {
	cfg := config.DefaultConfig()
	cfg.TempoRoot = filepath.Join(tempDir, ".tempo-files")
	cfg.App.GoPackage = filepath.Join(tempDir, "custom-package")
	cfg.App.AssetsDir = filepath.Join(tempDir, "custom-assets")
	cfg.Paths.TemplatesDir = filepath.Join(cfg.TempoRoot, "templates")
	cfg.Paths.ActionsDir = filepath.Join(cfg.TempoRoot, "actions")

	if overrides != nil {
		overrides(cfg)
	}
	return cfg
}

func TestComponentCommand_NewSubCmd_DefaultConfig(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := setupConfig(tempDir, nil)

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Write `tempo.yaml` to the current working directory
	configPath := filepath.Join(tempDir, "tempo.yaml")

	if err := testutils.WriteConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to create mock config file: %v", err)
	}

	// Prepare CLI app
	app := &cli.Command{
		Commands: []*cli.Command{
			SetupComponentCommand(cliCtx),
		},
	}

	// Step 1: Run "define component" to set up the required folder structure and files
	t.Run("Define Component Setup", func(t *testing.T) {
		_, err := testutils.SetupComponentDefine(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		expectedFiles := []string{
			filepath.Join(cfg.Paths.TemplatesDir, "component", "templ", "component.templ.gotxt"),
			filepath.Join(cfg.Paths.ActionsDir, "component.json"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	// Step 2: Run "new component" to test the command
	t.Run("Component with default config", func(t *testing.T) {
		output, err := testhelpers.CaptureStdout(func() {
			args := []string{
				"tempo", "component", "new",
				"--name", "button",
			}
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})

		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		testhelpers.ValidateCLIOutput(t, output, []string{
			"✔ Templ component files have been created",
		})

		expectedFiles := []string{
			filepath.Join(cfg.App.GoPackage, "button", "button.templ"),
			filepath.Join(cfg.App.GoPackage, "button", "css", "base.templ"),
			filepath.Join(cfg.App.AssetsDir, "button", "css", "base.css"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})
}

func TestComponentCommand_NewSubCmd_WithFlags(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := setupConfig(tempDir, nil)
	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Write `tempo.yaml`
	configPath := filepath.Join(tempDir, "tempo.yaml")
	if err := testutils.WriteConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to create mock config file: %v", err)
	}

	// Prepare CLI app
	app := &cli.Command{
		Commands: []*cli.Command{
			SetupComponentCommand(cliCtx),
		},
	}

	// Step 1: Run "define component"
	t.Run("Define Component Setup", func(t *testing.T) {
		_, err := testutils.SetupComponentDefine(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}
	})

	// Step 2: Run "new component" to test the command
	t.Run("Component with configs by flags", func(t *testing.T) {
		output, err := testhelpers.CaptureStdout(func() {
			args := []string{
				"tempo",
				"component",
				"new",
				"--package", cfg.App.GoPackage,
				"--name", "custom-component",
				"--assets", cfg.App.AssetsDir,
			}
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})

		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		testhelpers.ValidateCLIOutput(t, output, []string{
			"✔ Templ component files have been created",
		})

		// Transform name using the same logic as in the implementation
		transformedName := gonameprovider.ToGoPackageName("custom-component")
		expectedFiles := []string{
			filepath.Join(cfg.App.GoPackage, transformedName, fmt.Sprintf("%s.templ", transformedName)),
			filepath.Join(cfg.App.GoPackage, transformedName, "css", "base.templ"),
			filepath.Join(cfg.App.AssetsDir, transformedName, "css", "base.css"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})
}

func TestComponentCommand_NewSubComd_DryRun(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := setupConfig(tempDir, nil)
	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Write tempo.yaml to simulate "tempo init"
	configPath := filepath.Join(tempDir, "tempo.yaml")
	if err := testutils.WriteConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Prepare the CLI app with the create command.
	appCmd := &cli.Command{
		Commands: []*cli.Command{
			SetupComponentCommand(cliCtx),
		},
	}

	// Setup the required folders for "define component".
	t.Run("Define Component Setup", func(t *testing.T) {
		_, err := testutils.SetupComponentDefine(appCmd, t)
		if err != nil {
			t.Fatalf("SetupDefineComponent failed: %v", err)
		}
		// Allow some time for filesystem operations.
		time.Sleep(500 * time.Millisecond)
	})

	// Run the "new component" command with --dry-run flag.
	t.Run("Component Dry Run", func(t *testing.T) {
		output, err := testhelpers.CaptureStdout(func() {
			args := []string{
				"tempo", "component", "new",
				"--name", "dryrun",
				"--dry-run",
			}
			if err := appCmd.Run(context.Background(), args); err != nil {
				t.Fatalf("Command failed: %v", err)
			}
		})
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		// Check that the output contains the dry run message.
		if !strings.Contains(output, "Dry Run Mode: No changes will be made.") {
			t.Errorf("Expected dry run message in output, got: %s", output)
		}
	})
}

func TestComponentCommand_NewSubCmd_FailsOnMissingGoMod(t *testing.T) {
	tempDir := t.TempDir() // Create a temporary directory without go.mod

	cfg := setupConfig(tempDir, nil)
	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Ensure go.mod does NOT exist
	goModPath := filepath.Join(tempDir, "go.mod")
	if _, err := os.Stat(goModPath); !os.IsNotExist(err) {
		t.Fatalf("go.mod file should NOT exist for this test case")
	}

	// Prepare CLI app
	appCmd := &cli.Command{
		Commands: []*cli.Command{
			SetupComponentCommand(cliCtx),
		},
	}

	// Try running the command
	err := appCmd.Run(context.Background(), []string{"tempo", "component", "new", "--name", "button"})

	// Validate error
	if err == nil {
		t.Fatal("Expected error due to missing go.mod, but got none")
	}

	expectedErrorMsg := "missing go.mod file. Run 'go mod init' to create one"
	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Errorf("Unexpected error message. Expected: %q, got: %q", expectedErrorMsg, err.Error())
	}
}

func TestComponentCommand_NewSubCmd_MissingActionsFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	// Setup config but do not create component.json in the actions folder.
	cfg := config.DefaultConfig()
	cfg.TempoRoot = filepath.Join(tempDir, ".tempo-files")
	cfg.App.GoPackage = filepath.Join(tempDir, "custom-package")
	cfg.App.AssetsDir = filepath.Join(tempDir, "custom-assets")
	actionsDir := filepath.Join(cfg.TempoRoot, "actions")
	if err := os.MkdirAll(actionsDir, 0755); err != nil {
		t.Fatalf("Failed to create actions folder: %v", err)
	}

	// Write tempo.yaml to simulate "tempo init"
	configPath := filepath.Join(tempDir, "tempo.yaml")
	if err := testutils.WriteConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Ensure component.json is missing.
	os.Remove(filepath.Join(actionsDir, "component.json"))

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Prepare the CLI app with the create command.
	cmd := &cli.Command{
		Commands: []*cli.Command{
			SetupComponentCommand(cliCtx),
		},
	}
	args := []string{
		"tempo", "component", "new",
		"--name", "missingActions",
	}
	err := cmd.Run(context.Background(), args)
	if err == nil {
		t.Fatalf("Expected error due to missing actions file, but got nil")
	}
	// Accept error messages that either mention the specific phrase or list missing folders.
	if !utils.ContainsSubstring(err.Error(), "Cannot find actions folder. Did you run 'tempo component define' first?") &&
		!utils.ContainsSubstring(err.Error(), "Missing folders:") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestComponentCommand_NewSubCmd_MissingNameFlag(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := setupConfig(tempDir, nil)

	// Write tempo.yaml to simulate "tempo init"
	configPath := filepath.Join(tempDir, "tempo.yaml")
	if err := testutils.WriteConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Ensure required folders exist to pass validation
	componentTemplateDir := filepath.Join(cfg.Paths.TemplatesDir, "component")
	if err := os.MkdirAll(componentTemplateDir, 0755); err != nil {
		t.Fatalf("Failed to create component template directory: %v", err)
	}

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Prepare the CLI app with the create command.
	cmd := &cli.Command{
		Commands: []*cli.Command{
			SetupComponentCommand(cliCtx),
		},
	}
	args := []string{"tempo", "component", "new"} // Missing --name flag

	err := cmd.Run(context.Background(), args)
	if err == nil {
		t.Fatalf("Expected error due to missing --name flag, but got nil")
	}
	expectedErrorMsg := `Required flag "name" not set`
	if !utils.ContainsSubstring(err.Error(), expectedErrorMsg) {
		t.Errorf("Unexpected error message: got %q, expected to contain %q", err.Error(), expectedErrorMsg)
	}
}

func TestComponentCommand_NewSubCmd_CheckComponentExists(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := setupConfig(tempDir, nil)
	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Write `tempo.yaml` to the current working directory
	configPath := filepath.Join(tempDir, "tempo.yaml")
	if err := testutils.WriteConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to create mock config file: %v", err)
	}

	// Prepare CLI app
	app := &cli.Command{
		Commands: []*cli.Command{
			SetupComponentCommand(cliCtx),
		},
	}

	// Step 1: Run "define component" to set up the required folder structure and files
	t.Run("Define Component", func(t *testing.T) {
		_, err := testutils.SetupComponentDefine(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		expectedFiles := []string{
			filepath.Join(cfg.Paths.TemplatesDir, "component", "templ", "component.templ.gotxt"),
			filepath.Join(cfg.Paths.ActionsDir, "component.json"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	// Step 2: Run "new component" to test the command
	t.Run("Component with default config", func(t *testing.T) {
		output, err := testhelpers.CaptureStdout(func() {
			args := []string{
				"tempo", "component", "new",
				"--name", "button",
			}
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})

		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		testhelpers.ValidateCLIOutput(t, output, []string{
			"✔ Templ component files have been created",
		})

		expectedFiles := []string{
			filepath.Join(cfg.App.GoPackage, "button", "button.templ"),
			filepath.Join(cfg.App.GoPackage, "button", "css", "base.templ"),
			filepath.Join(cfg.App.AssetsDir, "button", "css", "base.css"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	// Step 3: Run "new component" again
	t.Run("Fail Component Creation When The Same already exists", func(t *testing.T) {
		output, err := testhelpers.CaptureStdout(func() {
			args := []string{
				"tempo", "component", "new",
				"--name", "button",
			}
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})

		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		if !strings.Contains(output, "Component 'button' already exists") {
			t.Errorf("Expected missing component error message, got: %s", output)
		}
	})
}

func TestComponentCommand_NewSubCmd_CorruptedActionsFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := setupConfig(tempDir, nil)
	// Write `tempo.yaml` to the current working directory
	configPath := filepath.Join(tempDir, "tempo.yaml")
	if err := testutils.WriteConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to create mock config file: %v", err)
	}

	// Ensure required folders exist to pass validation
	componentTemplateDir := filepath.Join(cfg.Paths.TemplatesDir, "component")
	if err := os.MkdirAll(componentTemplateDir, 0755); err != nil {
		t.Fatalf("Failed to create component template directory: %v", err)
	}

	// Ensure actions folder exists before writing the corrupted actions file
	actionsDir := cfg.Paths.ActionsDir
	if err := os.MkdirAll(actionsDir, 0755); err != nil {
		t.Fatalf("Failed to create actions directory: %v", err)
	}

	// Create a corrupted `component.json` file
	actionsFile := filepath.Join(actionsDir, "component.json")
	if err := os.WriteFile(actionsFile, []byte("INVALID_JSON"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted actions file: %v", err)
	}

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Prepare CLI app
	app := &cli.Command{
		Commands: []*cli.Command{
			SetupComponentCommand(cliCtx),
		},
	}
	args := []string{
		"tempo", "component", "new",
		"--name", "corrupted",
	}

	err := app.Run(context.Background(), args)
	if err == nil {
		t.Fatalf("Expected error due to corrupted actions file, but got nil")
	}

	// Adjust expected error message based on actual output
	expectedErrorMsg := "failed to process actions for component"
	if !utils.ContainsSubstring(err.Error(), expectedErrorMsg) {
		t.Errorf("Unexpected error message: got %q, expected to contain %q", err.Error(), expectedErrorMsg)
	}
}

func TestComponentCommand_NewSubCmd_UnwritableDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := setupConfig(tempDir, nil)

	// Write `tempo.yaml` to the current working directory
	configPath := filepath.Join(tempDir, "tempo.yaml")
	if err := testutils.WriteConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to create mock config file: %v", err)
	}

	// Ensure required folders exist to pass validation
	componentTemplateDir := filepath.Join(cfg.Paths.TemplatesDir, "component")
	if err := os.MkdirAll(componentTemplateDir, 0755); err != nil {
		t.Fatalf("Failed to create component template directory: %v", err)
	}

	// Ensure actions folder and `component.json` exist
	actionsDir := cfg.Paths.ActionsDir
	if err := os.MkdirAll(actionsDir, 0755); err != nil {
		t.Fatalf("Failed to create actions directory: %v", err)
	}
	actionsFile := filepath.Join(actionsDir, "component.json")
	if err := os.WriteFile(actionsFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create actions file: %v", err)
	}

	// Create an unwritable target directory
	componentDir := filepath.Join(cfg.App.GoPackage, "unwritable_component")
	if err := os.MkdirAll(componentDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}
	if err := os.Chmod(componentDir, 0000); err != nil {
		t.Fatalf("Failed to make directory unwritable: %v", err)
	}
	defer func() {
		if err := os.Chmod(componentDir, 0755); err != nil {
			t.Errorf("Failed to restore permissions on %s: %v", componentDir, err)
		}
	}()

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	app := &cli.Command{
		Commands: []*cli.Command{
			SetupComponentCommand(cliCtx),
		},
	}
	args := []string{
		"tempo", "component", "new",
		"--name", "unwritable_component",
		"--force", // Ensure it tries to write files
	}

	err := app.Run(context.Background(), args)
	if err == nil {
		t.Fatalf("Expected error due to unwritable directory, but got nil")
	}

	expectedErrorMsg := "failed to process actions for component"
	if !utils.ContainsSubstring(err.Error(), expectedErrorMsg) {
		t.Errorf("Unexpected error message: got %q, expected to contain %q", err.Error(), expectedErrorMsg)
	}
}

func TestComponentCommand_NewSubCmd_DryRun_NoChanges(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := setupConfig(tempDir, nil)

	// Write `tempo.yaml` to the current working directory
	configPath := filepath.Join(tempDir, "tempo.yaml")
	if err := testutils.WriteConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to create mock config file: %v", err)
	}

	// Ensure required template folders exist
	componentTemplateDir := filepath.Join(cfg.Paths.TemplatesDir, "component")
	if err := os.MkdirAll(componentTemplateDir, 0755); err != nil {
		t.Fatalf("Failed to create component template directory: %v", err)
	}

	// Ensure actions folder and `component.json` exist
	actionsDir := cfg.Paths.ActionsDir
	if err := os.MkdirAll(actionsDir, 0755); err != nil {
		t.Fatalf("Failed to create actions directory: %v", err)
	}
	actionsFile := filepath.Join(actionsDir, "component.json")
	if err := os.WriteFile(actionsFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create actions file: %v", err)
	}

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	app := &cli.Command{
		Commands: []*cli.Command{
			SetupComponentCommand(cliCtx),
		},
	}
	args := []string{
		"tempo", "component", "new",
		"--name", "dryrun_component",
		"--dry-run",
	}

	output, err := testhelpers.CaptureStdout(func() {
		err := app.Run(context.Background(), args)
		if err != nil {
			t.Fatalf("Command failed: %v", err)
		}
	})

	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Validate that the dry-run message appears
	expectedOutput := "Dry Run Mode: No changes will be made."
	if !utils.ContainsSubstring(output, expectedOutput) {
		t.Errorf("Expected output to contain %q, but got: %q", expectedOutput, output)
	}

	// Ensure processEntityActions was NOT called (no modifications should happen)
	expectedFailureMsg := "failed to process actions"
	if utils.ContainsSubstring(output, expectedFailureMsg) {
		t.Errorf("Unexpected error during dry-run: %q", output)
	}
}

func TestComponentCommand_NewSubCmd_Func_validateComponentNewPrerequisites(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := setupConfig(tempDir, nil)

	t.Run("Valid component prerequisites", func(t *testing.T) {
		// Create the required templates directory
		componentTemplateDir := filepath.Join(cfg.Paths.TemplatesDir, "component")
		if err := os.MkdirAll(componentTemplateDir, 0755); err != nil {
			t.Fatalf("Failed to create component template directory: %v", err)
		}

		validate := validateComponentNewPrerequisites(cfg)
		_, err := validate(context.Background(), &cli.Command{})

		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
	})

	t.Run("Missing component templates directory", func(t *testing.T) {
		// Ensure the templates directory does not exist
		componentTemplateDir := filepath.Join(cfg.Paths.TemplatesDir, "component")
		os.RemoveAll(componentTemplateDir)

		validate := validateComponentNewPrerequisites(cfg)
		_, err := validate(context.Background(), &cli.Command{})

		if err == nil {
			t.Fatal("Expected an error due to missing component templates directory, but got none")
		}

		if !strings.Contains(err.Error(), "Missing folders") {
			t.Errorf("Expected missing folders error message, got: %v", err)
		}
	})
}

func TestComponentCommand_NewSubCmd_Func_validateComponentNewPrerequisites_ErrorOnCheckMissingFolders(t *testing.T) {
	tempDir := t.TempDir()
	cfg := setupConfig(tempDir, nil)

	// Mock `CheckMissingFoldersFunc` to simulate a missing folder error
	originalFunc := utils.CheckMissingFoldersFunc
	defer func() { utils.CheckMissingFoldersFunc = originalFunc }() // Restore after test

	utils.CheckMissingFoldersFunc = func(folders map[string]string) ([]string, error) {
		return []string{"  - Templates directory: /mock/path/component"}, fmt.Errorf("mock error in CheckMissingFolders")
	}

	validate := validateComponentNewPrerequisites(cfg)
	_, err := validate(context.Background(), &cli.Command{})

	if err == nil {
		t.Fatal("Expected an error due to CheckMissingFolders failure, but got nil")
	}

	// Instead of checking for the mock error, check that the error message contains "Missing folders"
	expectedSubstring := "Missing folders:"
	if !strings.Contains(err.Error(), expectedSubstring) {
		t.Errorf("Expected error message to contain %q, but got %q", expectedSubstring, err.Error())
	}
}

func TestComponentCommand_NewSubCmd_Func_createBaseTemplateData_DefaultValues(t *testing.T) {
	tempDir := t.TempDir()

	cfg := setupConfig(tempDir, nil)

	appCmd := &cli.Command{
		Flags: getNewFlags(),
	}

	// Call function without setting any flags
	data, err := createBaseTemplateData(appCmd, cfg)
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}

	// Verify defaults
	if data.GoPackage != cfg.App.GoPackage {
		t.Errorf("Expected default GoPackage, got: %s", data.GoPackage)
	}
	if data.AssetsDir != cfg.App.AssetsDir {
		t.Errorf("Expected default AssetsDir, got: %s", data.AssetsDir)
	}
	if data.WithJs != cfg.App.WithJs {
		t.Errorf("Expected default WithJs value, got: %v", data.WithJs)
	}
	if data.Force != false {
		t.Errorf("Expected default Force to be false, got: %v", data.Force)
	}
	if data.DryRun != false {
		t.Errorf("Expected default DryRun to be false, got: %v", data.DryRun)
	}
}

func TestComponentCommand_NewSubCmd_Func_resolveActionFilePath(t *testing.T) {
	tempDir := t.TempDir()
	actionsDir := filepath.Join(tempDir, "actions")
	existingFile := filepath.Join(actionsDir, "existing.json")
	nonExistentFile := filepath.Join(tempDir, "nonexistent.json")
	mockError := fmt.Errorf("mock error")

	tests := []struct {
		name           string
		actionsDir     string
		actionFileFlag string
		mockFileExists func(path string) (bool, error)
		expectedErr    string
		expectedResult string
	}{
		{
			name:           "Action file found in ActionsDir",
			actionsDir:     actionsDir,
			actionFileFlag: "existing.json",
			mockFileExists: func(path string) (bool, error) {
				if path == existingFile {
					return true, nil
				}
				return false, nil
			},
			expectedResult: existingFile,
		},
		{
			name:           "Action file does not exist",
			actionsDir:     "",
			actionFileFlag: nonExistentFile,
			mockFileExists: func(path string) (bool, error) {
				return false, nil
			},
			expectedErr: "action file does not exist",
		},
		{
			name:           "Error checking resolvedPath",
			actionsDir:     actionsDir,
			actionFileFlag: "error.json",
			mockFileExists: func(path string) (bool, error) {
				if strings.Contains(path, actionsDir) {
					return false, mockError
				}
				return false, nil
			},
			expectedErr: "mock error",
		},
		{
			name:           "Error checking absolute path",
			actionsDir:     "",
			actionFileFlag: nonExistentFile,
			mockFileExists: func(path string) (bool, error) {
				if path == nonExistentFile {
					return false, mockError
				}
				return false, nil
			},
			expectedErr: "error checking action file path",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Mock the function
			utils.FileExistsFunc = tc.mockFileExists
			defer func() { utils.FileExistsFunc = utils.FileExists }() // Reset after test

			result, err := resolveActionFilePath(tc.actionsDir, tc.actionFileFlag)

			if tc.expectedErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.expectedErr) {
					t.Fatalf("Expected error %q, but got: %v", tc.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result != tc.expectedResult {
					t.Fatalf("Expected result %q, but got %q", tc.expectedResult, result)
				}
			}
		})
	}
}
