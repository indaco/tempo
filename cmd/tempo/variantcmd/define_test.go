package variantcmd

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/indaco/tempo/cmd/tempo/componentcmd"
	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/testhelpers"
	"github.com/indaco/tempo/internal/testutils"
	"github.com/urfave/cli/v3"
)

func TestVariantCommand_Define(t *testing.T) {
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
			componentcmd.SetupComponentCommand(cliCtx),
			SetupVariantCommand(cliCtx),
		},
	}
	// Redirect CLI output to buffer
	var buf bytes.Buffer
	app.Writer = &buf

	t.Run("Component normal run", func(t *testing.T) {
		_, err := testhelpers.CaptureStdout(func() {
			args := []string{"tempo", "component", "define"}
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
		_, err := testhelpers.CaptureStdout(func() {
			args := []string{"tempo", "variant", "define"}
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

// /* ------------------------------------------------------------------------- */
// /* Test VariantCommand_Define Existing Entity                                        */
// /* ------------------------------------------------------------------------- */

func TestVariantCommand_Define_AlreadyExists(t *testing.T) {
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
			componentcmd.SetupComponentCommand(cliCtx),
			SetupVariantCommand(cliCtx),
		},
	}
	// Redirect CLI output to buffer
	var buf bytes.Buffer
	app.Writer = &buf

	t.Run("Component normal run", func(t *testing.T) {
		_, err := testhelpers.CaptureStdout(func() {
			args := []string{"tempo", "component", "define"}
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
		_, err := testhelpers.CaptureStdout(func() {
			args := []string{"tempo", "variant", "define"}
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
		output, err := testutils.SetupVariantDefine(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		testhelpers.ValidateCLIOutput(t, output, []string{
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
