package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/indaco/tempo/internal/testhelpers"
	"github.com/indaco/tempo/internal/utils"
)

// TestRunApp_Version simply verifies that running with "--version" returns the version.
func TestRunApp_Version(t *testing.T) {
	// Create a temporary directory.
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

// setupTempDir initializes a temporary directory and returns it along with the original working directory.
func setupTempDir(t *testing.T) (string, string) {
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
	if err := os.Chdir(origDir); err != nil {
		t.Fatalf("Failed to restore working directory: %v", err)
	}
}
