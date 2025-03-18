package componentcmd

import (
	"bytes"
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

// Test DefineCommand Normal Execution
func TestComponentCommand_DefineSubComd(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	// Prepare configuration
	cfg := config.DefaultConfig()
	testutils.PrepareTestConfig(cfg, tempDir)

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Write `tempo.yaml` in the current working directory
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
	// Redirect CLI output to buffer
	var buf bytes.Buffer
	app.Writer = &buf

	t.Run("Component normal run", func(t *testing.T) {
		output, err := testutils.SetupComponentDefine(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		testhelpers.ValidateCLIOutput(t, output, []string{
			"✔ Templates for the component and assets (CSS and JS) have been created",
			"✔ Tempo action file for 'component' has been created",
		})

		// Ensure no files are created
		expectedFiles := []string{
			filepath.Join(cfg.Paths.ActionsDir, "component.json"),
			filepath.Join(cfg.Paths.TemplatesDir, "component", "templ", "component.templ.gotxt"),
		}

		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

}

// Test Define Command DryRun
func TestComponentCommand_DefineSubComd_WithDryRun(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	// Prepare configuration
	cfg := config.DefaultConfig()
	testutils.PrepareTestConfig(cfg, tempDir)

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Write `tempo.yaml` in the current working directory
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
	// Redirect CLI output to buffer
	var buf bytes.Buffer
	app.Writer = &buf

	t.Run("Component dry-run", func(t *testing.T) {
		args := []string{"tempo", "component", "define", "--dry-run"}
		err := app.Run(context.Background(), args)
		// Validate error expectation
		if err != nil {
			t.Fatalf("Unexpected error state: %v", err)
		}

		// Validate CLI execution
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Ensure no files are created
		expectedFiles := []string{
			filepath.Join(cfg.Paths.ActionsDir, "component.json"),
			filepath.Join(cfg.Paths.TemplatesDir, "component", "templ", "component.templ.gotxt"),
		}
		for _, path := range expectedFiles {
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Errorf("Expected file should not exist in dry-run mode: %s", path)
			}
		}
	})
}

// Test DefineCommand Existing Entity
func TestComponentCommand_DefineSubComd_AlreadyExists(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	// Prepare configuration
	cfg := config.DefaultConfig()
	testutils.PrepareTestConfig(cfg, tempDir)

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Write `tempo.yaml` in the current working directory
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
	// Redirect CLI output to buffer
	var buf bytes.Buffer
	app.Writer = &buf

	t.Run("Component normal run", func(t *testing.T) {
		output, err := testutils.SetupComponentDefine(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		testhelpers.ValidateCLIOutput(t, output, []string{
			"✔ Templates for the component and assets (CSS and JS) have been created",
			"✔ Tempo action file for 'component' has been created",
		})

		// Ensure no files are created
		expectedFiles := []string{
			filepath.Join(cfg.Paths.ActionsDir, "component.json"),
			filepath.Join(cfg.Paths.TemplatesDir, "component", "templ", "component.templ.gotxt"),
		}

		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	t.Run("Component already exists", func(t *testing.T) {
		output, err := testutils.SetupComponentDefine(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		testhelpers.ValidateCLIOutput(t, output, []string{
			"⚠ Templates for 'component' already exist.",
			"  Use '--force' to overwrite them. Any changes will be lost.",
			"  - path: " + cfg.Paths.TemplatesDir + "/component",
		})

		// Ensure no files are created
		expectedFiles := []string{
			filepath.Join(cfg.Paths.ActionsDir, "component.json"),
			filepath.Join(cfg.Paths.TemplatesDir, "component", "templ", "component.templ.gotxt"),
		}

		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

}

// Test Handling When No Config File Exists
func TestComponentCommand_DefineSubComd_NoConfigFile(t *testing.T) {
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

	app := &cli.Command{
		Commands: []*cli.Command{
			SetupComponentCommand(cliCtx),
		},
	}

	t.Run("No Config File", func(t *testing.T) {
		args := []string{"tempo", "component", "define"}
		err := app.Run(context.Background(), args)
		if err == nil {
			t.Fatal("Expected error due to missing config file, but got nil")
		}

		expectedError := "no config file found; checked"
		if !testutils.Contains(err.Error(), expectedError) {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

func TestComponentCommand_DefineSubCmd_PermissionError(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := config.DefaultConfig()
	testutils.PrepareTestConfig(cfg, tempDir)

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Write valid `tempo.yaml`
	configPath := filepath.Join(tempDir, "tempo.yaml")
	if err := testutils.WriteConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to create mock config file: %v", err)
	}

	app := &cli.Command{
		Commands: []*cli.Command{
			SetupComponentCommand(cliCtx),
		},
	}

	// Ensure the directory exists before changing permissions
	if err := os.MkdirAll(cfg.Paths.ActionsDir, 0755); err != nil {
		t.Fatalf("Failed to set up test directory: %v", err)
	}

	// Restrict permissions (simulate permission error)
	if err := os.Chmod(cfg.Paths.ActionsDir, 0000); err != nil {
		t.Fatalf("Failed to set restricted permissions: %v", err)
	}

	// Ensure permissions are restored after the test
	defer func() {
		_ = os.Chmod(cfg.Paths.ActionsDir, 0755)
	}()

	t.Run("Permission Error", func(t *testing.T) {
		args := []string{"tempo", "component", "define"}
		err := app.Run(context.Background(), args)

		if err == nil {
			t.Fatal("Expected permission error, but got nil")
		}

		expectedError := "permission denied"
		if !strings.Contains(err.Error(), expectedError) {
			t.Errorf("Expected permission denied error, got: %v", err)
		}
	})
}
