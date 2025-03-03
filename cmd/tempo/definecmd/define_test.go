package definecmd

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
	"github.com/indaco/tempo/testutils"
	"github.com/urfave/cli/v3"
)

/* ------------------------------------------------------------------------- */
/* Test Setup Define Command                                                 */
/* ------------------------------------------------------------------------- */

func TestSetupDefineCommand(t *testing.T) {
	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: config.DefaultConfig(),
		CWD:    t.TempDir(),
	}

	command := SetupDefineCommand(cliCtx)
	if command == nil {
		t.Fatal("SetupDefineCommand returned nil")
	}

	if command.Name != "define" {
		t.Errorf("Expected command name 'define', got '%s'", command.Name)
	}
}

/* ------------------------------------------------------------------------- */
/* Test DefineCommand Normal Execution                                       */
/* ------------------------------------------------------------------------- */

func TestDefineCommand_Component(t *testing.T) {

	tempDir := t.TempDir()
	// Prepare configuration
	cfg := config.DefaultConfig()
	prepareTestConfig(cfg, tempDir)

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
			SetupDefineCommand(cliCtx),
		},
	}
	// Redirect CLI output to buffer
	var buf bytes.Buffer
	app.Writer = &buf

	t.Run("Component normal run", func(t *testing.T) {
		output, err := testutils.SetupDefineComponent(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		testutils.ValidateCLIOutput(t, output, []string{
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

func TestDefineCommand_Component_WithDryRun(t *testing.T) {
	tempDir := t.TempDir()
	// Prepare configuration
	cfg := config.DefaultConfig()
	prepareTestConfig(cfg, tempDir)

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
			SetupDefineCommand(cliCtx),
		},
	}
	// Redirect CLI output to buffer
	var buf bytes.Buffer
	app.Writer = &buf

	t.Run("Component dry-run", func(t *testing.T) {
		args := []string{"tempo", "define", "component", "--dry-run"}
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

func TestDefineCommand_Variant(t *testing.T) {

	tempDir := t.TempDir()
	// Prepare configuration
	cfg := config.DefaultConfig()
	prepareTestConfig(cfg, tempDir)

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
			SetupDefineCommand(cliCtx),
		},
	}
	// Redirect CLI output to buffer
	var buf bytes.Buffer
	app.Writer = &buf

	t.Run("Component normal run", func(t *testing.T) {
		_, err := testutils.CaptureStdout(func() {
			args := []string{"tempo", "define", "component"}
			err := app.Run(context.Background(), args)
			// Validate error expectation
			if err != nil {
				t.Fatalf("Unexpected error state: %v", err)
			}
		})

		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		// Ensure no files are created
		expectedFiles := []string{
			filepath.Join(cfg.Paths.ActionsDir, "component.json"),
			filepath.Join(cfg.Paths.TemplatesDir, "component", "templ", "component.templ.gotxt"),
		}

		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	t.Run("Variant normal run", func(t *testing.T) {
		_, err := testutils.CaptureStdout(func() {
			args := []string{"tempo", "define", "variant"}
			err := app.Run(context.Background(), args)
			// Validate error expectation
			if err != nil {
				t.Fatalf("Unexpected error state: %v", err)
			}
		})

		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		// Validate generated files
		expectedFiles := []string{
			filepath.Join(cfg.Paths.ActionsDir, "variant.json"),
			filepath.Join(cfg.Paths.TemplatesDir, "component-variant", "name.templ.gotxt"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})
}

/* ------------------------------------------------------------------------- */
/* Test DefineCommand Existing Entity                                        */
/* ------------------------------------------------------------------------- */

func TestDefineCommand_Component_AlreadyExists(t *testing.T) {

	tempDir := t.TempDir()
	// Prepare configuration
	cfg := config.DefaultConfig()
	prepareTestConfig(cfg, tempDir)

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
			SetupDefineCommand(cliCtx),
		},
	}
	// Redirect CLI output to buffer
	var buf bytes.Buffer
	app.Writer = &buf

	t.Run("Component normal run", func(t *testing.T) {
		output, err := testutils.SetupDefineComponent(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		testutils.ValidateCLIOutput(t, output, []string{
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
		output, err := testutils.SetupDefineComponent(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		testutils.ValidateCLIOutput(t, output, []string{
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

func TestDefineCommand_Variant_AlreadyExists(t *testing.T) {

	tempDir := t.TempDir()
	// Prepare configuration
	cfg := config.DefaultConfig()
	prepareTestConfig(cfg, tempDir)

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
			SetupDefineCommand(cliCtx),
		},
	}
	// Redirect CLI output to buffer
	var buf bytes.Buffer
	app.Writer = &buf

	t.Run("Component normal run", func(t *testing.T) {
		_, err := testutils.CaptureStdout(func() {
			args := []string{"tempo", "define", "component"}
			err := app.Run(context.Background(), args)
			// Validate error expectation
			if err != nil {
				t.Fatalf("Unexpected error state: %v", err)
			}
		})

		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		// Ensure no files are created
		expectedFiles := []string{
			filepath.Join(cfg.Paths.ActionsDir, "component.json"),
			filepath.Join(cfg.Paths.TemplatesDir, "component", "templ", "component.templ.gotxt"),
		}

		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	t.Run("Variant normal run", func(t *testing.T) {
		_, err := testutils.CaptureStdout(func() {
			args := []string{"tempo", "define", "variant"}
			err := app.Run(context.Background(), args)
			// Validate error expectation
			if err != nil {
				t.Fatalf("Unexpected error state: %v", err)
			}
		})

		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		// Validate generated files
		expectedFiles := []string{
			filepath.Join(cfg.Paths.ActionsDir, "variant.json"),
			filepath.Join(cfg.Paths.TemplatesDir, "component-variant", "name.templ.gotxt"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	t.Run("Variant already exists", func(t *testing.T) {
		output, err := testutils.SetupDefineVariant(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		testutils.ValidateCLIOutput(t, output, []string{
			"⚠ Templates for 'variant' already exist.",
			"  Use '--force' to overwrite them. Any changes will be lost.",
			"  - path: " + cfg.Paths.TemplatesDir + "/component-variant",
		})

		// Ensure no files are created
		expectedFiles := []string{
			filepath.Join(cfg.Paths.ActionsDir, "variant.json"),
			filepath.Join(cfg.Paths.TemplatesDir, "component-variant", "name.templ.gotxt"),
		}

		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})
}

/* ------------------------------------------------------------------------- */
/* Test Handling When No Config File Exists                                  */
/* ------------------------------------------------------------------------- */

func TestDefineCommand_NoConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: config.DefaultConfig(),
		CWD:    tempDir,
	}

	app := &cli.Command{Commands: []*cli.Command{SetupDefineCommand(cliCtx)}}

	t.Run("No Config File", func(t *testing.T) {
		args := []string{"tempo", "define", "component"}
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

/* ------------------------------------------------------------------------- */
/* Test Handling of Permission Errors                                        */
/* ------------------------------------------------------------------------- */

func TestDefineCommand_PermissionError(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.DefaultConfig()
	prepareTestConfig(cfg, tempDir)

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

	app := &cli.Command{Commands: []*cli.Command{SetupDefineCommand(cliCtx)}}

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
		args := []string{"tempo", "define", "component"}
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

/* ------------------------------------------------------------------------- */
/* Test handleEntityExistence                                                */
/* ------------------------------------------------------------------------- */

func TestHandleEntityExistence(t *testing.T) {
	tests := []struct {
		name        string
		entityType  string
		outputPath  string
		force       bool
		expectedMsg []string
	}{
		{
			name:       "Entity exists without force flag",
			entityType: "component",
			outputPath: "/mock/path/to/component",
			force:      false,
			expectedMsg: []string{
				"⚠ Templates for 'component' already exist.",
				"  Use '--force' to overwrite them. Any changes will be lost.",
				"  - path: /mock/path/to/component",
			},
		},
		{
			name:       "Entity exists with force flag",
			entityType: "component",
			outputPath: "/mock/path/to/component",
			force:      true,
			expectedMsg: []string{
				"ℹ Templates for 'component' already exist.",
				"  Overwriting due to '--force' flag.",
				"  - path: /mock/path/to/component",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture logger output
			output, err := testutils.CaptureStdout(func() {
				log := logger.NewDefaultLogger()
				handleEntityExistence(tt.entityType, tt.outputPath, tt.force, log)
			})

			if err != nil {
				t.Fatalf("Failed to capture logger output: %v", err)
			}

			// Validate output
			testutils.ValidateCLIOutput(t, output, tt.expectedMsg)
		})
	}
}

/* ------------------------------------------------------------------------- */
/* Helper Function for Setup                                                 */
/* ------------------------------------------------------------------------- */

func prepareTestConfig(cfg *config.Config, tempDir string) {
	cfg.TempoRoot = filepath.Join(tempDir, ".tempo-files")
	cfg.Paths.TemplatesDir = filepath.Join(cfg.TempoRoot, "templates")
	cfg.Paths.ActionsDir = filepath.Join(cfg.TempoRoot, "actions")
}
