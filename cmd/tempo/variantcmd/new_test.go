package variantcmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/indaco/tempo/cmd/tempo/componentcmd"
	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/templatefuncs/providers/gonameprovider"
	"github.com/indaco/tempo/internal/testhelpers"
	"github.com/indaco/tempo/internal/testutils"
	"github.com/indaco/tempo/internal/utils"
	"github.com/urfave/cli/v3"
)

func TestVariantCommand_NewSubCmd_DefaultConfig(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := testutils.SetupConfig(tempDir, nil)
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
			componentcmd.SetupComponentCommand(cliCtx),
			SetupVariantCommand(cliCtx),
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

	// Step 2: Run "component new" to test the command
	t.Run("Create new component with default config", func(t *testing.T) {
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

	// Step 3: Run "variant define" to set up the required folder structure and files
	t.Run("Define Variant Setup", func(t *testing.T) {
		_, err := testutils.SetupVariantDefine(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		// Validate that the "variant define" command generated the expected files
		expectedFiles := []string{
			filepath.Join(cfg.Paths.TemplatesDir, "component-variant", "name.templ.gotxt"),
			filepath.Join(cfg.Paths.ActionsDir, "variant.json"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	//Step 4: Run "variant new" to test the command
	t.Run("Variant with default config", func(t *testing.T) {
		output, err := testhelpers.CaptureStdout(func() {
			args := []string{
				"tempo",
				"variant",
				"new",
				"--name", "neon",
				"--component", "button",
			}
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})

		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		// Validate CLI output
		testhelpers.ValidateCLIOutput(t, output, []string{
			"✔ Templ component for the variant and asset files (CSS) have been created",
		})

		// Validate generated files
		expectedFiles := []string{
			filepath.Join(cfg.App.GoPackage, "button", "css", "variants", "neon.templ"),
			filepath.Join(cfg.App.AssetsDir, "button", "css", "variants", "neon.css"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})
}

func TestVariantCommand_NewSubCmd_WithFlags(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := testutils.SetupConfig(tempDir, func(cfg *config.Config) {
		cfg.App.GoPackage = filepath.Join(tempDir, "custom-package")
		cfg.App.AssetsDir = filepath.Join(tempDir, "custom-assets")
	})

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Write `tempo.yaml` to the current working directory
	configPath := filepath.Join(cliCtx.CWD, "tempo.yaml")
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

	// Step 1: Run "define component"
	t.Run("Define Component Setup", func(t *testing.T) {
		_, err := testutils.SetupComponentDefine(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}
	})

	// Step 2: Run "component new" to test the command
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

		//Transform name using the same logic as in the implementation
		transformedName := gonameprovider.ToGoPackageName("custom-component")
		expectedFiles := []string{
			filepath.Join(cfg.App.GoPackage, transformedName, fmt.Sprintf("%s.templ", transformedName)),
			filepath.Join(cfg.App.GoPackage, transformedName, "css", "base.templ"),
			filepath.Join(cfg.App.AssetsDir, transformedName, "css", "base.css"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	// Step 3: Run "variant define" to set up the required folder structure and files
	t.Run("Define Variant Setup", func(t *testing.T) {
		_, err := testutils.SetupVariantDefine(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		// Validate that the "variant define" command generated the expected files
		expectedFiles := []string{
			filepath.Join(cfg.Paths.TemplatesDir, "component-variant", "name.templ.gotxt"),
			filepath.Join(cfg.Paths.ActionsDir, "variant.json"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	// Step 4: Run "variant new" to test the command
	t.Run("Variant with custom flags", func(t *testing.T) {
		output, err := testhelpers.CaptureStdout(func() {
			args := []string{
				"tempo",
				"variant",
				"new",
				"--package", cfg.App.GoPackage,
				"--name", "secondary",
				"--component", "custom-component",
				"--assets", cfg.App.AssetsDir,
			}
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})

		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		// Validate CLI output
		testhelpers.ValidateCLIOutput(t, output, []string{
			"✔ Templ component for the variant and asset files (CSS) have been created",
		})

		// Transform name using the same logic as in the implementation
		transformedName := gonameprovider.ToGoPackageName("secondary")

		// Validate generated files
		expectedFiles := []string{
			filepath.Join(cfg.App.GoPackage, "custom_component", "css", "variants", fmt.Sprintf("%s.templ", transformedName)),
			filepath.Join(cfg.App.AssetsDir, "custom_component", "css", "variants", fmt.Sprintf("%s.css", transformedName)),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})
}

func TestVariantCommand_NewSubComd_DryRun(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := testutils.SetupConfig(tempDir, nil)
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
			componentcmd.SetupComponentCommand(cliCtx),
			SetupVariantCommand(cliCtx),
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

	// Run the "component new" command with --dry-run flag.
	t.Run("Component Dry Run", func(t *testing.T) {
		_, err := testhelpers.CaptureStdout(func() {
			args := []string{
				"tempo", "component", "new",
				"--name", "button",
			}
			if err := appCmd.Run(context.Background(), args); err != nil {
				t.Fatalf("Command failed: %v", err)
			}
		})
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

	})

	// Step 3: Run "variant define" to set up the required folder structure and files
	t.Run("Define Variant Setup", func(t *testing.T) {
		_, err := testutils.SetupVariantDefine(appCmd, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		// Validate that the "variant define" command generated the expected files
		expectedFiles := []string{
			filepath.Join(cfg.Paths.TemplatesDir, "component-variant", "name.templ.gotxt"),
			filepath.Join(cfg.Paths.ActionsDir, "variant.json"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	// Run the "component new" command with --dry-run flag.
	t.Run("Component Dry Run", func(t *testing.T) {
		output, err := testhelpers.CaptureStdout(func() {
			args := []string{
				"tempo", "variant", "new",
				"--name", "outline",
				"--component", "button",
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

func TestVariantCommand_NewSubCmd_DryRun_NoChanges(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := testutils.SetupConfig(tempDir, nil)
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
			componentcmd.SetupComponentCommand(cliCtx),
			SetupVariantCommand(cliCtx),
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

	// Step 2: Run "component new" to test the command
	t.Run("Create new component with default config", func(t *testing.T) {
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

	// Step 3: Run "variant define" to set up the required folder structure and files
	t.Run("Define Variant Setup", func(t *testing.T) {
		_, err := testutils.SetupVariantDefine(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		// Validate that the "variant define" command generated the expected files
		expectedFiles := []string{
			filepath.Join(cfg.Paths.TemplatesDir, "component-variant", "name.templ.gotxt"),
			filepath.Join(cfg.Paths.ActionsDir, "variant.json"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	//Step 4: Run "variant new" to test the command
	t.Run("Variant with default config", func(t *testing.T) {
		output, err := testhelpers.CaptureStdout(func() {
			args := []string{
				"tempo",
				"variant",
				"new",
				"--name", "neon",
				"--component", "button",
				"--dry-run",
			}
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatalf("Unexpected error: %v", err)
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

func TestVariantCommand_NewSubCmd_FailsOnMissingGoMod(t *testing.T) {
	tempDir := t.TempDir() // Create a temporary directory without go.mod

	cfg := testutils.SetupConfig(tempDir, nil)
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
			componentcmd.SetupComponentCommand(cliCtx),
			SetupVariantCommand(cliCtx),
		},
	}

	args := []string{"tempo", "variant", "new", "--name", "outline", "--component", "button"}

	// Try running the command
	err := appCmd.Run(context.Background(), args)

	// Validate error
	if err == nil {
		t.Fatal("Expected error due to missing go.mod, but got none")
	}

	expectedErrorMsg := "missing go.mod file. Run 'go mod init' to create one"
	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Errorf("Unexpected error message. Expected: %q, got: %q", expectedErrorMsg, err.Error())
	}
}

func TestVariantCommand_NewSubCmd_MissingActionsFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

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

	// Ensure variant.json is missing.
	os.Remove(filepath.Join(actionsDir, "variant.json"))

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Prepare CLI app
	app := &cli.Command{
		Commands: []*cli.Command{
			componentcmd.SetupComponentCommand(cliCtx),
			SetupVariantCommand(cliCtx),
		},
	}

	args := []string{
		"tempo", "variant", "new",
		"--name", "missingVariantActions",
		"--component", "someComponent",
	}
	err := app.Run(context.Background(), args)
	if err == nil {
		t.Fatalf("Expected error due to missing variant actions file, but got nil")
	}
	// Accept error messages that either match the short form or include "Missing folders:".
	if !utils.ContainsSubstring(err.Error(), "Cannot find actions folder. Did you run 'tempo variant define' before?") &&
		!utils.ContainsSubstring(err.Error(), "Missing folders:") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestVariantCommand_NewSubCmd_MissingNameFlag(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := testutils.SetupConfig(tempDir, func(cfg *config.Config) {
		cfg.App.GoPackage = filepath.Join(tempDir, "custom-package")
		cfg.App.AssetsDir = filepath.Join(tempDir, "custom-assets")
	})

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Write `tempo.yaml` to the current working directory
	configPath := filepath.Join(cliCtx.CWD, "tempo.yaml")
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

	// Step 1: Run "define component"
	t.Run("Define Component Setup", func(t *testing.T) {
		_, err := testutils.SetupComponentDefine(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}
	})

	// Step 2: Run "component new" to test the command
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

		//Transform name using the same logic as in the implementation
		transformedName := gonameprovider.ToGoPackageName("custom-component")
		expectedFiles := []string{
			filepath.Join(cfg.App.GoPackage, transformedName, fmt.Sprintf("%s.templ", transformedName)),
			filepath.Join(cfg.App.GoPackage, transformedName, "css", "base.templ"),
			filepath.Join(cfg.App.AssetsDir, transformedName, "css", "base.css"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	// Step 3: Run "variant define" to set up the required folder structure and files
	t.Run("Define Variant Setup", func(t *testing.T) {
		_, err := testutils.SetupVariantDefine(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		// Validate that the "variant define" command generated the expected files
		expectedFiles := []string{
			filepath.Join(cfg.Paths.TemplatesDir, "component-variant", "name.templ.gotxt"),
			filepath.Join(cfg.Paths.ActionsDir, "variant.json"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	// Step 4: Run "variant new" to test the command
	t.Run("Variant with custom flags", func(t *testing.T) {
		args := []string{
			"tempo",
			"variant",
			"new",
			"--package", cfg.App.GoPackage,
			"--component", "custom-component",
			"--assets", cfg.App.AssetsDir,
		}

		err := app.Run(context.Background(), args)
		if err == nil {
			t.Fatalf("Expected error due to missing --name flag, but got nil")
		}
		expectedErrorMsg := `Required flag "name" not set`
		if !utils.ContainsSubstring(err.Error(), expectedErrorMsg) {
			t.Errorf("Unexpected error message: got %q, expected to contain %q", err.Error(), expectedErrorMsg)
		}
	})
}

func TestVariantCommand_NewSubCmd_MissingFolders(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := testutils.SetupConfig(tempDir, nil)
	// Ensure one of the required folders (e.g. component-variant) is missing.
	variantDir := filepath.Join(cfg.Paths.TemplatesDir, "component-variant")
	os.RemoveAll(variantDir)

	validate := validateVariantNewPrerequisites(cfg)
	_, err := validate(context.Background(), &cli.Command{})
	if err == nil {
		t.Fatal("Expected an error due to missing folders, but got nil")
	}
	if !strings.Contains(err.Error(), "Missing folders") {
		t.Errorf("Expected error message to mention missing folders, got: %v", err)
	}
}

func TestVariantCommand_NewSubCmd_CheckComponentExists(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := testutils.SetupConfig(tempDir, nil)
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
			componentcmd.SetupComponentCommand(cliCtx),
			SetupVariantCommand(cliCtx),
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

	// Step 3: Run "variant define" to set up the required folder structure and files
	t.Run("Define Variant Setup", func(t *testing.T) {
		_, err := testutils.SetupVariantDefine(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		// Validate that the "variant define" command generated the expected files
		expectedFiles := []string{
			filepath.Join(cfg.Paths.TemplatesDir, "component-variant", "name.templ.gotxt"),
			filepath.Join(cfg.Paths.ActionsDir, "variant.json"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	t.Run("Fail Variant Creation When Component Does Not Exist", func(t *testing.T) {
		output, err := testhelpers.CaptureStdout(func() {
			args := []string{
				"tempo", "variant", "new",
				"--name", "ghost",
				"--component", "nonexistent-component",
			}
			if err := app.Run(context.Background(), args); err == nil {
				t.Fatalf("Expected error due to missing component, but got none")
			}
		})

		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		if !strings.Contains(output, "Cannot create variant: Component does not exist") {
			t.Errorf("Expected missing component error message, got: %s", output)
		}
	})
}

// !@TODO
func TestVariantCommand_NewSubCmd_CorruptedActionsFile(t *testing.T) {

}

// !@TODO
func TestVariantCommand_NewSubCmd_UnwritableDirectory(t *testing.T) {}

func TestVariantCommand_NewSubCmd_AlreadyExists_NoForce(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := testutils.SetupConfig(tempDir, nil)
	cfg.TempoRoot = filepath.Join(tempDir, ".tempo-files")
	cfg.App.GoPackage = filepath.Join(tempDir, "custom-package")
	cfg.App.AssetsDir = filepath.Join(tempDir, "custom-assets")
	// Create required actions folder and a dummy variant actions file.
	actionsDir := filepath.Join(cfg.TempoRoot, "actions")
	if err := os.MkdirAll(actionsDir, 0755); err != nil {
		t.Fatalf("Failed to create actions folder: %v", err)
	}

	// Write tempo.yaml to simulate "tempo init"
	configPath := filepath.Join(tempDir, "tempo.yaml")
	if err := testutils.WriteConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	variantActionsPath := filepath.Join(actionsDir, "variant.json")
	if err := os.WriteFile(variantActionsPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create variant actions file: %v", err)
	}

	// Create required template folders so prerequisites pass.
	componentTemplateDir := filepath.Join(cfg.Paths.TemplatesDir, "component")
	variantTemplateDir := filepath.Join(cfg.Paths.TemplatesDir, "component-variant")
	if err := os.MkdirAll(componentTemplateDir, 0755); err != nil {
		t.Fatalf("Failed to create component template directory: %v", err)
	}
	if err := os.MkdirAll(variantTemplateDir, 0755); err != nil {
		t.Fatalf("Failed to create variant template directory: %v", err)
	}

	// Simulate that the component already exists.
	componentName := "testcomp"
	goPackagePath := cfg.App.GoPackage
	componentPath := filepath.Join(goPackagePath, componentName)
	if err := os.MkdirAll(componentPath, 0755); err != nil {
		t.Fatalf("Failed to create component folder: %v", err)
	}
	// Expected variant output file.
	variantName := "testvar"
	variantOutputPath := filepath.Join(goPackagePath, componentName, "css", "variants", variantName+".templ")
	if err := os.MkdirAll(filepath.Dir(variantOutputPath), 0755); err != nil {
		t.Fatalf("Failed to create variant output folder: %v", err)
	}
	// Create a dummy variant file.
	dummyContent := []byte("dummy content")
	if err := os.WriteFile(variantOutputPath, dummyContent, 0644); err != nil {
		t.Fatalf("Failed to create dummy variant file: %v", err)
	}

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Prepare CLI app
	app := &cli.Command{
		Commands: []*cli.Command{
			componentcmd.SetupComponentCommand(cliCtx),
			SetupVariantCommand(cliCtx),
		},
	}
	// Provide flags without --force.
	args := []string{
		"tempo", "variant", "new",
		"--name", variantName,
		"--component", componentName,
	}
	err := app.Run(context.Background(), args)
	// Expect nil error because the command should return early.
	if err != nil {
		t.Fatalf("Expected nil error when variant exists without --force, got: %v", err)
	}
	// Verify that the existing variant file was not overwritten.
	contents, err := os.ReadFile(variantOutputPath)
	if err != nil {
		t.Fatalf("Failed to read variant output file: %v", err)
	}
	if string(contents) != string(dummyContent) {
		t.Errorf("Expected variant file to remain unchanged when --force is not provided")
	}
}

func TestVariantCommand_NewSubCmd_ComponentDoesNotExist(t *testing.T) {
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	cfg := testutils.SetupConfig(tempDir, nil)

	// Ensure required folders exist to pass validation
	componentTemplateDir := filepath.Join(cfg.Paths.TemplatesDir, "component")
	variantTemplateDir := filepath.Join(cfg.Paths.TemplatesDir, "component-variant")
	if err := os.MkdirAll(componentTemplateDir, 0755); err != nil {
		t.Fatalf("Failed to create component template directory: %v", err)
	}
	if err := os.MkdirAll(variantTemplateDir, 0755); err != nil {
		t.Fatalf("Failed to create variant template directory: %v", err)
	}

	// Write tempo.yaml to simulate "tempo init"
	configPath := filepath.Join(tempDir, "tempo.yaml")
	if err := testutils.WriteConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Ensure actions folder and variant.json exist
	actionsFile := filepath.Join(cfg.Paths.ActionsDir, "variant.json")
	if err := os.MkdirAll(cfg.Paths.ActionsDir, 0755); err != nil {
		t.Fatalf("Failed to create actions directory: %v", err)
	}
	if err := os.WriteFile(actionsFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create empty actions file: %v", err)
	}

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Prepare CLI app
	app := &cli.Command{
		Commands: []*cli.Command{
			componentcmd.SetupComponentCommand(cliCtx),
			SetupVariantCommand(cliCtx),
		},
	}
	args := []string{
		"tempo", "variant", "new",
		"--name", "outline",
		"--component", "nonexistent-component", // This component does not exist
	}

	err := app.Run(context.Background(), args)
	if err == nil {
		t.Fatalf("Expected error due to missing component, but got nil")
	}

	expectedErrorMsg := "Cannot create variant: Component does not exist"
	if !utils.ContainsSubstring(err.Error(), expectedErrorMsg) {
		t.Errorf("Unexpected error message: got %q, expected to contain %q", err.Error(), expectedErrorMsg)
	}
}

func TestVariantCommand_NewSubCmd_validateVariantNewPrerequisites(t *testing.T) {
	tempDir := t.TempDir()
	cfg := testutils.SetupConfig(tempDir, nil)

	componentPath := filepath.Join(cfg.Paths.TemplatesDir, "component")
	variantPath := filepath.Join(cfg.Paths.TemplatesDir, "component-variant")

	tests := []struct {
		name           string
		missingFolders map[string]string
		expectedParts  []string
	}{
		{
			name: "Component directory missing",
			missingFolders: map[string]string{
				"Component Directory": componentPath,
			},
			expectedParts: []string{
				"Have you run 'tempo component define' or 'tempo variant define' to set up your components?",
				"Missing folders:\n  - Component Directory:",
			},
		},
		{
			name: "Both component and variant directories missing",
			missingFolders: map[string]string{
				"Component Directory": componentPath,
				"Variant Directory":   variantPath,
			},
			expectedParts: []string{
				"Have you run 'tempo component define' or 'tempo variant define' to set up your components?",
				"Missing folders:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			validate := validateVariantNewPrerequisites(cfg)
			_, err := validate(context.Background(), &cli.Command{})

			if len(tt.missingFolders) == 0 {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected error containing %q, but got nil", tt.expectedParts)
				}

				for _, expected := range tt.expectedParts {
					if !utils.ContainsSubstring(err.Error(), expected) {
						t.Errorf("Expected error message to contain %q, but got: %q", expected, err.Error())
					}
				}
			}
		})
	}
}

func TestVariantCommand_NewSubCmd_Func_createBaseTemplateData_DefaultValues(t *testing.T) {
	tempDir := t.TempDir()

	cfg := testutils.SetupConfig(tempDir, nil)

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
	if data.Force != false {
		t.Errorf("Expected default Force to be false, got: %v", data.Force)
	}
	if data.DryRun != false {
		t.Errorf("Expected default DryRun to be false, got: %v", data.DryRun)
	}
}

func TestVariantCommand_NewSubCmd_Func_resolveActionFilePath(t *testing.T) {
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
