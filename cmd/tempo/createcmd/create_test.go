package createcmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/tempo/cmd/tempo/definecmd"
	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/templatefuncs/providers/gonameprovider"
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
