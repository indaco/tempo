package registercmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/templatefuncs/registry"
	"github.com/indaco/tempo/internal/testhelpers"
	"github.com/indaco/tempo/internal/testutils"
)

// setupConfig creates a test config similar to your "create" command tests.
func setupConfig(tempDir string, overrides func(cfg *config.Config)) *config.Config {
	cfg := config.DefaultConfig()
	cfg.TempoRoot = filepath.Join(tempDir, ".tempo-files")
	// Set other paths as needed.
	cfg.App.GoPackage = filepath.Join(tempDir, "go-package")
	cfg.App.AssetsDir = filepath.Join(tempDir, "assets")
	cfg.Paths.TemplatesDir = filepath.Join(cfg.TempoRoot, "templates")
	cfg.Paths.ActionsDir = filepath.Join(cfg.TempoRoot, "actions")
	if overrides != nil {
		overrides(cfg)
	}
	return cfg
}

// clearRegistry removes all entries from the global function registry.
func clearRegistry() {
	funcs := registry.GetRegisteredFunctions()
	for k := range funcs {
		delete(funcs, k)
	}
}

// TestRegisterCommand_Functions_Repo tests registering functions from a remote repository.
func TestRegisterCommand_Functions_Repo(t *testing.T) {
	tempDir := t.TempDir()
	cfg := setupConfig(tempDir, nil)
	// Write a valid config file (simulate "tempo init")
	configPath := filepath.Join(tempDir, "tempo.yaml")
	if err := testutils.WriteConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	appCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Clear registry for isolation.
	clearRegistry()

	// Build the CLI command.
	cmd := SetupRegisterCommand(appCtx)
	// We target the "functions" subcommand.
	// Here we provide the --url flag to register from a repository.
	args := []string{
		"register", "functions",
		"--url", "https://github.com/indaco/tempo-provider-sprig.git",
	}

	output, err := testhelpers.CaptureStdout(func() {
		if err := cmd.Run(context.Background(), args); err != nil {
			t.Fatalf("Command failed: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Verify output indicates success.
	if !strings.Contains(output, "Functions successfully registered!") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify that functions are registered.
	funcs := registry.GetRegisteredFunctions()
	if len(funcs) == 0 {
		t.Errorf("Expected functions to be registered, got 0")
	}
	// Check for a known function from the repo provider.
	if _, ok := funcs["toStrings"]; !ok {
		t.Errorf("Expected function 'toStrings' to be registered, but it was not found")
	}
}

// TestRegisterCommand_Functions_Local tests registering functions from a local provider.
func TestRegisterCommand_Functions_Local(t *testing.T) {
	tempDir := t.TempDir()
	cfg := setupConfig(tempDir, nil)
	// Write a config file to simulate "tempo init".
	configPath := filepath.Join(tempDir, "tempo.yaml")
	if err := testutils.WriteConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create a temporary local provider directory.
	providerDir := filepath.Join(tempDir, "mockprovider")
	if err := os.MkdirAll(providerDir, 0755); err != nil {
		t.Fatalf("Failed to create provider directory: %v", err)
	}
	// Write a minimal go.mod file so that the module is valid.
	goModContent := "module mockprovider\n\ngo 1.20\n"
	if err := os.WriteFile(filepath.Join(providerDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}
	// Write a minimal provider.go file.
	providerContent := `package mockprovider

import (
	"text/template"
	"github.com/indaco/tempo-api/templatefuncs"
)

type MockProvider struct{}

func (p *MockProvider) GetFunctions() template.FuncMap {
	return template.FuncMap{"localFunc": func() string { return "Hello from local!" }}
}

var Provider templatefuncs.TemplateFuncProvider = &MockProvider{}
`
	if err := os.WriteFile(filepath.Join(providerDir, "provider.go"), []byte(providerContent), 0644); err != nil {
		t.Fatalf("Failed to write provider.go: %v", err)
	}

	appCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}
	clearRegistry()

	cmd := SetupRegisterCommand(appCtx)
	// Use the --path flag to register functions from the local provider.
	args := []string{
		"register", "functions",
		"--path", providerDir,
	}

	output, err := testhelpers.CaptureStdout(func() {
		if err := cmd.Run(context.Background(), args); err != nil {
			t.Fatalf("Command failed: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}
	if !strings.Contains(output, "Functions successfully registered!") {
		t.Errorf("Expected success message, got: %s", output)
	}

	funcs := registry.GetRegisteredFunctions()
	if len(funcs) == 0 {
		t.Errorf("Expected functions to be registered, got 0")
	}
	if _, ok := funcs["localFunc"]; !ok {
		t.Errorf("Expected function 'localFunc' to be registered, but it was not found")
	}
}

// TestRegisterCommand_Functions_Both tests registering when both --url and --path are provided.
func TestRegisterCommand_Functions_Both(t *testing.T) {
	tempDir := t.TempDir()
	cfg := setupConfig(tempDir, nil)
	// Write config file to simulate "tempo init".
	configPath := filepath.Join(tempDir, "tempo.yaml")
	if err := testutils.WriteConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Setup a local provider.
	providerDir := filepath.Join(tempDir, "mockprovider")
	if err := os.MkdirAll(providerDir, 0755); err != nil {
		t.Fatalf("Failed to create provider directory: %v", err)
	}
	// Create a minimal go.mod for the local provider.
	goModContent := "module mockprovider\n\ngo 1.20\n"
	if err := os.WriteFile(filepath.Join(providerDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}
	providerContent := `package mockprovider

import (
	"text/template"
	"github.com/indaco/tempo-api/templatefuncs"
)

type MockProvider struct{}

func (p *MockProvider) GetFunctions() template.FuncMap {
	return template.FuncMap{"localFunc": func() string { return "Hello from local!" }}
}

var Provider templatefuncs.TemplateFuncProvider = &MockProvider{}
`
	if err := os.WriteFile(filepath.Join(providerDir, "provider.go"), []byte(providerContent), 0644); err != nil {
		t.Fatalf("Failed to write provider.go: %v", err)
	}

	appCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}
	clearRegistry()

	cmd := SetupRegisterCommand(appCtx)
	// When both --url and --path are provided, two providers are registered with differentiated names.
	args := []string{
		"register", "functions",
		"--url", "https://github.com/indaco/tempo-provider-sprig.git",
		"--path", providerDir,
	}

	output, err := testhelpers.CaptureStdout(func() {
		if err := cmd.Run(context.Background(), args); err != nil {
			t.Fatalf("Command failed: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}
	if !strings.Contains(output, "Functions successfully registered!") {
		t.Errorf("Expected success message, got: %s", output)
	}

	funcs := registry.GetRegisteredFunctions()
	if len(funcs) == 0 {
		t.Errorf("Expected functions to be registered, got 0")
	}
	// Check for a known function from the repo provider.
	if _, ok := funcs["toStrings"]; !ok {
		t.Errorf("Expected function 'toStrings' from repo provider to be registered, but it was not found")
	}
	// Check for the function from the local provider.
	if _, ok := funcs["localFunc"]; !ok {
		t.Errorf("Expected function 'localFunc' from local provider to be registered, but it was not found")
	}
}
