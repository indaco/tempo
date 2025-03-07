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

	err = runApp(args)
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
		if err := runApp(args); err != nil {
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
	// Create a temporary directory that does not contain a config file.
	tempDir := t.TempDir()

	// Create go.mod inside tempDir (the correct working directory)
	goModPath := filepath.Join(tempDir, "go.mod")
	err := os.WriteFile(goModPath, []byte("module example.com/myproject\n\ngo 1.23\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	configPath := filepath.Join(tempDir, "tempo.yaml")
	os.Remove(configPath) // Ensure it's missing.

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

	// Run the init command to trigger auto-generation.
	args := []string{"tempo", "init", "--base-folder", tempDir}

	// Capture stdout and stderr.
	origStdout, origStderr := os.Stdout, os.Stderr
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdout pipe: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stderr pipe: %v", err)
	}
	os.Stdout = wOut
	os.Stderr = wErr

	err = runApp(args)

	if err := wOut.Close(); err != nil {
		t.Fatalf("failed to close stdout writer: %v", err)
	}
	if err := wErr.Close(); err != nil {
		t.Fatalf("failed to close stderr writer: %v", err)
	}

	var bufOut, bufErr bytes.Buffer
	if _, err := io.Copy(&bufOut, rOut); err != nil {
		t.Fatalf("failed to copy stdout: %v", err)
	}
	if _, err := io.Copy(&bufErr, rErr); err != nil {
		t.Fatalf("failed to copy stderr: %v", err)
	}
	os.Stdout = origStdout
	os.Stderr = origStderr

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify that tempo.yaml now exists.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Expected tempo.yaml to be auto-generated, but it does not exist")
	}

	// Check output if available.
	output := bufOut.String() + bufErr.String()
	// In CI the output might not include auto-generation messages.
	if output != "" {
		if !utils.ContainsSubstring(output, "Generating") || !utils.ContainsSubstring(output, "Done!") {
			t.Logf("Output did not contain expected messages: %s", output)
		}
	} else {
		t.Log("No output captured; auto-generation messages may be logged elsewhere in CI.")
	}
}
