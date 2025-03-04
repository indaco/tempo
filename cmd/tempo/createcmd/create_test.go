package createcmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/indaco/tempo/cmd/tempo/definecmd"
	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/templatefuncs/providers/gonameprovider"
	"github.com/indaco/tempo/internal/utils"
	"github.com/indaco/tempo/testutils"
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

func TestCreateCommandComponent_DefaultConfig(t *testing.T) {
	tempDir := t.TempDir()
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
			definecmd.SetupDefineCommand(cliCtx),
			SetupCreateCommand(cliCtx),
		},
	}

	// Step 1: Run "define component" to set up the required folder structure and files
	t.Run("Define Component Setup", func(t *testing.T) {
		_, err := testutils.SetupDefineComponent(app, t)
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
		output, err := testutils.CaptureStdout(func() {
			args := []string{
				"tempo", "new", "component",
				"--name", "button",
			}
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})

		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		testutils.ValidateCLIOutput(t, output, []string{
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

func TestCreateCommandComponent_WithFlags(t *testing.T) {
	tempDir := t.TempDir()
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
			definecmd.SetupDefineCommand(cliCtx),
			SetupCreateCommand(cliCtx),
		},
	}

	// Step 1: Run "define component"
	t.Run("Define Component Setup", func(t *testing.T) {
		_, err := testutils.SetupDefineComponent(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}
	})

	// Step 2: Run "new component" to test the command
	t.Run("Component with configs by flags", func(t *testing.T) {
		output, err := testutils.CaptureStdout(func() {
			args := []string{
				"tempo",
				"new",
				"component",
				"--module", "custom-module",
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

		testutils.ValidateCLIOutput(t, output, []string{
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

func TestCreateCommandVariant_DefaultConfig(t *testing.T) {
	tempDir := t.TempDir()
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
			definecmd.SetupDefineCommand(cliCtx),
			SetupCreateCommand(cliCtx),
		},
	}

	// Step 1: Run "define component" to set up the required folder structure and files
	t.Run("Define Component", func(t *testing.T) {
		_, err := testutils.SetupDefineComponent(app, t)
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
	t.Run("Create new component with default config", func(t *testing.T) {
		output, err := testutils.CaptureStdout(func() {
			args := []string{
				"tempo", "new", "component",
				"--name", "button",
			}
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})

		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		testutils.ValidateCLIOutput(t, output, []string{
			"✔ Templ component files have been created",
		})

		expectedFiles := []string{
			filepath.Join(cfg.App.GoPackage, "button", "button.templ"),
			filepath.Join(cfg.App.GoPackage, "button", "css", "base.templ"),
			filepath.Join(cfg.App.AssetsDir, "button", "css", "base.css"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	// Step 3: Run "define variant" to set up the required folder structure and files
	t.Run("Define Variant Setup", func(t *testing.T) {
		_, err := testutils.SetupDefineVariant(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		// Validate that the "define variant" command generated the expected files
		expectedFiles := []string{
			filepath.Join(cfg.Paths.TemplatesDir, "component-variant", "name.templ.gotxt"),
			filepath.Join(cfg.Paths.ActionsDir, "variant.json"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	//Step 4: Run "new variant" to test the command
	t.Run("Variant with default config", func(t *testing.T) {
		output, err := testutils.CaptureStdout(func() {
			args := []string{
				"tempo",
				"new",
				"variant",
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
		testutils.ValidateCLIOutput(t, output, []string{
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

func TestCreateCommandVariant_WithFlags(t *testing.T) {
	tempDir := t.TempDir()
	// Prepare configuration
	cfg := setupConfig(tempDir, func(cfg *config.Config) {
		cfg.App.GoPackage = filepath.Join(tempDir, "custom-package")
		cfg.App.AssetsDir = filepath.Join(tempDir, "custom-assets")
	})
	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    t.TempDir(),
	}

	// Write `tempo.yaml` to the current working directory
	configPath := filepath.Join(cliCtx.CWD, "tempo.yaml")
	if err := testutils.WriteConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to create mock config file: %v", err)
	}

	// Prepare CLI app
	app := &cli.Command{
		Commands: []*cli.Command{
			definecmd.SetupDefineCommand(cliCtx),
			SetupCreateCommand(cliCtx),
		},
	}

	// Step 1: Run "define component"
	t.Run("Define Component Setup", func(t *testing.T) {
		_, err := testutils.SetupDefineComponent(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}
	})

	// Step 2: Run "new component" to test the command
	t.Run("Component with configs by flags", func(t *testing.T) {
		output, err := testutils.CaptureStdout(func() {
			args := []string{
				"tempo",
				"new",
				"component",
				"--module", "custom-module",
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

		testutils.ValidateCLIOutput(t, output, []string{
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

	// Step 3: Run "define variant" to set up the required folder structure and files
	t.Run("Define Variant Setup", func(t *testing.T) {
		_, err := testutils.SetupDefineVariant(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		// Validate that the "define variant" command generated the expected files
		expectedFiles := []string{
			filepath.Join(cfg.Paths.TemplatesDir, "component-variant", "name.templ.gotxt"),
			filepath.Join(cfg.Paths.ActionsDir, "variant.json"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	// Step 4: Run "new variant" to test the command
	t.Run("Variant with custom flags", func(t *testing.T) {
		output, err := testutils.CaptureStdout(func() {
			args := []string{
				"tempo",
				"new",
				"variant",
				"--module", "custom-module",
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
		testutils.ValidateCLIOutput(t, output, []string{
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

func TestHandleEntityExistence(t *testing.T) {
	tests := []struct {
		name       string
		entityType string
		entityName string
		outputPath string
		force      bool
		shouldWarn bool
	}{
		{"Component Exists Without Force", "component", "button", "/mock/path/button", false, true},
		{"Component Exists With Force", "component", "button", "/mock/path/button", true, false},
		{"Variant Exists Without Force", "variant", "outline", "/mock/path/button/css/variants/outline.templ", false, true},
		{"Variant Exists With Force", "variant", "outline", "/mock/path/button/css/variants/outline.templ", true, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := testutils.CaptureStdout(func() {
				logger := logger.NewDefaultLogger()
				handleEntityExistence(tc.entityType, tc.entityName, tc.outputPath, tc.force, logger)
			})

			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			if tc.shouldWarn {
				if !strings.Contains(output, "Use '--force' to overwrite it.") {
					t.Errorf("Expected warning message, got: %s", output)
				}
			} else {
				if !strings.Contains(output, "Overwriting due to '--force' flag.") {
					t.Errorf("Expected overwrite message, got: %s", output)
				}
			}
		})
	}
}

func TestCreateComponent_CheckComponentExists(t *testing.T) {
	tempDir := t.TempDir()
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
			definecmd.SetupDefineCommand(cliCtx),
			SetupCreateCommand(cliCtx),
		},
	}

	// Step 1: Run "define component" to set up the required folder structure and files
	t.Run("Define Component", func(t *testing.T) {
		_, err := testutils.SetupDefineComponent(app, t)
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
		output, err := testutils.CaptureStdout(func() {
			args := []string{
				"tempo", "new", "component",
				"--name", "button",
			}
			if err := app.Run(context.Background(), args); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})

		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		testutils.ValidateCLIOutput(t, output, []string{
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
		output, err := testutils.CaptureStdout(func() {
			args := []string{
				"tempo", "new", "component",
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

func TestCreateVariant_CheckComponentExists(t *testing.T) {
	tempDir := t.TempDir()
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
			definecmd.SetupDefineCommand(cliCtx),
			SetupCreateCommand(cliCtx),
		},
	}

	// Step 1: Run "define component" to set up the required folder structure and files
	t.Run("Define Component", func(t *testing.T) {
		_, err := testutils.SetupDefineComponent(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		expectedFiles := []string{
			filepath.Join(cfg.Paths.TemplatesDir, "component", "templ", "component.templ.gotxt"),
			filepath.Join(cfg.Paths.ActionsDir, "component.json"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	// Step 3: Run "define variant" to set up the required folder structure and files
	t.Run("Define Variant Setup", func(t *testing.T) {
		_, err := testutils.SetupDefineVariant(app, t)
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		// Validate that the "define variant" command generated the expected files
		expectedFiles := []string{
			filepath.Join(cfg.Paths.TemplatesDir, "component-variant", "name.templ.gotxt"),
			filepath.Join(cfg.Paths.ActionsDir, "variant.json"),
		}
		testutils.ValidateGeneratedFiles(t, expectedFiles)
	})

	t.Run("Fail Variant Creation When Component Does Not Exist", func(t *testing.T) {
		output, err := testutils.CaptureStdout(func() {
			args := []string{
				"tempo", "new", "variant",
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

func TestValidateCreateComponentPrerequisites(t *testing.T) {
	tempDir := t.TempDir()
	cfg := setupConfig(tempDir, nil)

	t.Run("Valid component prerequisites", func(t *testing.T) {
		// Create the required templates directory
		componentTemplateDir := filepath.Join(cfg.Paths.TemplatesDir, "component")
		if err := os.MkdirAll(componentTemplateDir, 0755); err != nil {
			t.Fatalf("Failed to create component template directory: %v", err)
		}

		validate := validateCreateComponentPrerequisites(cfg)
		_, err := validate(context.Background(), &cli.Command{})

		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
	})

	t.Run("Missing component templates directory", func(t *testing.T) {
		// Ensure the templates directory does not exist
		componentTemplateDir := filepath.Join(cfg.Paths.TemplatesDir, "component")
		os.RemoveAll(componentTemplateDir)

		validate := validateCreateComponentPrerequisites(cfg)
		_, err := validate(context.Background(), &cli.Command{})

		if err == nil {
			t.Fatal("Expected an error due to missing component templates directory, but got none")
		}

		if !strings.Contains(err.Error(), "Missing folders") {
			t.Errorf("Expected missing folders error message, got: %v", err)
		}
	})
}

// TestValidateCreateVariantPrerequisites_MissingFolders tests that when required folders are missing,
// validateCreateVariantPrerequisites returns an error containing "Missing folders".
func TestValidateCreateVariantPrerequisites_MissingFolders(t *testing.T) {
	tempDir := t.TempDir()
	cfg := setupConfig(tempDir, nil)
	// Ensure one of the required folders (e.g. component-variant) is missing.
	variantDir := filepath.Join(cfg.Paths.TemplatesDir, "component-variant")
	os.RemoveAll(variantDir)

	validate := validateCreateVariantPrerequisites(cfg)
	_, err := validate(context.Background(), &cli.Command{})
	if err == nil {
		t.Fatal("Expected an error due to missing folders, but got nil")
	}
	if !strings.Contains(err.Error(), "Missing folders") {
		t.Errorf("Expected error message to mention missing folders, got: %v", err)
	}
}

// TestCreateCommandComponent_DryRun tests that when the --dry-run flag is provided,
// the command logs the dry run message.
func TestCreateCommandComponent_DryRun(t *testing.T) {
	tempDir := t.TempDir()
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
			definecmd.SetupDefineCommand(cliCtx),
			SetupCreateCommand(cliCtx),
		},
	}

	// Setup the required folders for "define component".
	t.Run("Define Component Setup", func(t *testing.T) {
		_, err := testutils.SetupDefineComponent(appCmd, t)
		if err != nil {
			t.Fatalf("SetupDefineComponent failed: %v", err)
		}
		// Allow some time for filesystem operations.
		time.Sleep(500 * time.Millisecond)
	})

	// Run the "new component" command with --dry-run flag.
	t.Run("Component Dry Run", func(t *testing.T) {
		output, err := testutils.CaptureStdout(func() {
			args := []string{
				"tempo", "new", "component",
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

// createTestComponentCmd returns the CLI command for creating a component.
func createTestComponentCmd(cliCtx *app.AppContext) *cli.Command {
	return SetupCreateCommand(cliCtx).Commands[0] // component subcommand
}

// createTestVariantCmd returns the CLI command for creating a variant.
func createTestVariantCmd(cliCtx *app.AppContext) *cli.Command {
	return SetupCreateCommand(cliCtx).Commands[1] // variant subcommand
}

// TestCreateComponent_MissingActionsFile tests that if the required component actions file
// is missing, the command returns an error with the expected message.
func TestCreateComponent_MissingActionsFile(t *testing.T) {
	tempDir := t.TempDir()
	// Setup config but do not create component.json in the actions folder.
	cfg := config.DefaultConfig()
	cfg.TempoRoot = filepath.Join(tempDir, ".tempo-files")
	cfg.App.GoPackage = filepath.Join(tempDir, "custom-package")
	cfg.App.AssetsDir = filepath.Join(tempDir, "custom-assets")
	actionsDir := filepath.Join(cfg.TempoRoot, "actions")
	if err := os.MkdirAll(actionsDir, 0755); err != nil {
		t.Fatalf("Failed to create actions folder: %v", err)
	}
	// Ensure component.json is missing.
	os.Remove(filepath.Join(actionsDir, "component.json"))

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	cmd := createTestComponentCmd(cliCtx)
	args := []string{
		"tempo", "new", "component",
		"--name", "missingActions",
	}
	err := cmd.Run(context.Background(), args)
	if err == nil {
		t.Fatalf("Expected error due to missing actions file, but got nil")
	}
	// Accept error messages that either mention the specific phrase or list missing folders.
	if !utils.ContainsSubstring(err.Error(), "Cannot find actions folder. Did you run 'tempo define component' before?") &&
		!utils.ContainsSubstring(err.Error(), "Missing folders:") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestCreateVariant_MissingActionsFile tests that if the required variant actions file is missing,
// the command returns an error with an expected substring.
func TestCreateVariant_MissingActionsFile(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TempoRoot = filepath.Join(tempDir, ".tempo-files")
	cfg.App.GoPackage = filepath.Join(tempDir, "custom-package")
	cfg.App.AssetsDir = filepath.Join(tempDir, "custom-assets")
	actionsDir := filepath.Join(cfg.TempoRoot, "actions")
	if err := os.MkdirAll(actionsDir, 0755); err != nil {
		t.Fatalf("Failed to create actions folder: %v", err)
	}
	// Ensure variant.json is missing.
	os.Remove(filepath.Join(actionsDir, "variant.json"))

	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	cmd := createTestVariantCmd(cliCtx)
	args := []string{
		"tempo", "new", "variant",
		"--name", "missingVariantActions",
		"--component", "someComponent",
	}
	err := cmd.Run(context.Background(), args)
	if err == nil {
		t.Fatalf("Expected error due to missing variant actions file, but got nil")
	}
	// Accept error messages that either match the short form or include "Missing folders:".
	if !utils.ContainsSubstring(err.Error(), "Cannot find actions folder. Did you run 'tempo define variant' before?") &&
		!utils.ContainsSubstring(err.Error(), "Missing folders:") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestCreateVariant_AlreadyExists_NoForce tests that if the variant file already exists and --force is not provided,
// the command returns early without overwriting the file.
func TestCreateVariant_AlreadyExists_NoForce(t *testing.T) {
	tempDir := t.TempDir()
	cfg := config.DefaultConfig()
	cfg.TempoRoot = filepath.Join(tempDir, ".tempo-files")
	cfg.App.GoPackage = filepath.Join(tempDir, "custom-package")
	cfg.App.AssetsDir = filepath.Join(tempDir, "custom-assets")
	// Create required actions folder and a dummy variant actions file.
	actionsDir := filepath.Join(cfg.TempoRoot, "actions")
	if err := os.MkdirAll(actionsDir, 0755); err != nil {
		t.Fatalf("Failed to create actions folder: %v", err)
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

	cmd := createTestVariantCmd(cliCtx)
	// Provide flags without --force.
	args := []string{
		"tempo", "new", "variant",
		"--name", variantName,
		"--component", componentName,
	}
	err := cmd.Run(context.Background(), args)
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
