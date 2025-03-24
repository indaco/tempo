package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/testhelpers"
	"github.com/indaco/tempo/internal/utils"
	"github.com/indaco/tempo/internal/version"
)

func TestNewCLIFields(t *testing.T) {
	// Setup a dummy AppContext.
	// Note: If config.Config has required fields, you might want to initialize them.
	dummyCfg := &config.Config{}
	dummyLogger := logger.NewDefaultLogger()
	cliCtx := &app.AppContext{
		Logger: dummyLogger,
		Config: dummyCfg,
		CWD:    utils.GetCWD(),
	}

	// Build the CLI command.
	cmd := newCLI(cliCtx)

	// Verify basic fields.
	if cmd.Name != appName {
		t.Errorf("expected command name %q, got %q", appName, cmd.Name)
	}
	expectedVersion := fmt.Sprintf("v%s", version.GetVersion())
	if cmd.Version != expectedVersion {
		t.Errorf("expected version %q, got %q", expectedVersion, cmd.Version)
	}
	if cmd.Usage != usage {
		t.Errorf("expected usage %q, got %q", usage, cmd.Usage)
	}
	if cmd.Description != description {
		t.Errorf("expected description %q, got %q", description, cmd.Description)
	}

	// Verify that the expected subcommands are present.
	expectedSubcommands := []string{"init", "component", "variant", "register", "sync"}
	if len(cmd.Commands) != len(expectedSubcommands) {
		t.Errorf("expected %d subcommands, got %d", len(expectedSubcommands), len(cmd.Commands))
	}
	for _, expected := range expectedSubcommands {
		found := false
		for _, sub := range cmd.Commands {
			if sub.Name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q not found", expected)
		}
	}
}

// TestRunApp_Version simply verifies that running with "--version" returns the version.
func TestRunApp_Version(t *testing.T) {
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Use the "--version" flag.
	args := []string{"tempo", "--version"}

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	err = runCLI(args)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to copy output: %v", err)
	}
	os.Stdout = origStdout

	output := buf.String()
	if !utils.ContainsSubstring(output, "tempo version") {
		t.Errorf("Expected version string in output, got: %s", output)
	}
}

// TestRunApp_Help verifies that running the app with "--help" produces help output that contains "USAGE:" and "COMMANDS:".
func TestRunApp_Help(t *testing.T) {
	// Create a temporary directory and change into it.
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Prepare arguments for help.
	args := []string{"tempo", "--help"}

	// Capture stdout.
	_, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	output, err := testhelpers.CaptureStdout(func() {
		if err := runCLI(args); err != nil {
			t.Fatalf("runApp returned error: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture output: %v", err)
	}
	testhelpers.ValidateCLIOutput(t, output, []string{"USAGE:", "COMMANDS:"})
}

// TestRunApp_InitAutoGen verifies that when no config file exists, the "init" command auto-generates one.
func TestRunApp_InitAutoGen(t *testing.T) {
	tempDir, origDir := setupTempDir(t)
	defer restoreWorkingDir(t, origDir)

	configPath := filepath.Join(tempDir, "tempo.yaml")
	os.Remove(configPath) // Ensure it's missing.

	// Run the command and capture output
	output, err := testhelpers.CaptureStdout(func() {
		_ = runCLI([]string{"tempo", "init", "--base-folder", tempDir})
	})

	if err != nil {
		t.Fatalf("Failed to capture output: %v", err)
	}

	// Verify that tempo.yaml now exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Expected tempo.yaml to be auto-generated, but it does not exist")
	}

	// Validate output messages
	testhelpers.ValidateCLIOutput(t, output, []string{"Generating", "Done!"})
}

/* ------------------------------------------------------------------------- */
/* ERROR CASES                                                               */
/* ------------------------------------------------------------------------- */
func TestRunCLI_LoadConfigError(t *testing.T) {
	tmp := t.TempDir()

	// Create an unreadable tempo.yaml file
	configPath := filepath.Join(tmp, "tempo.yaml")
	if err := os.WriteFile(configPath, []byte("path: ./broken"), 0000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(configPath, 0644) // Restore permissions so temp dir can be deleted
	})

	// Switch to temp dir
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	// Run CLI to trigger config loading
	err = runCLI([]string{"tempo", "init"})
	if err == nil {
		t.Fatal("expected error from LoadConfig, got nil")
	}
	if !utils.ContainsSubstring(err.Error(), "failed to read config file:") {
		t.Errorf("unexpected error: %v", err)
	}
}

/* ------------------------------------------------------------------------- */
/* HELPERS                                                                   */
/* ------------------------------------------------------------------------- */

// setupTempDir initializes a temporary directory and returns it along with the original working directory.
func setupTempDir(t *testing.T) (string, string) {
	t.Helper()
	tempDir := t.TempDir()

	goModPath := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module example.com/myproject\n\ngo 1.23\n"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	return tempDir, origDir
}

// restoreWorkingDir restores the original working directory after the test.
func restoreWorkingDir(t *testing.T, origDir string) {
	t.Helper()
	if err := os.Chdir(origDir); err != nil {
		t.Fatalf("Failed to restore working directory: %v", err)
	}
}
