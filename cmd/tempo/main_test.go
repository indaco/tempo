package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/tempo/testutils"
)

// TestRunApp_Version verifies that running the app with "--version" prints the version string.
func TestRunApp_Version(t *testing.T) {
	// Create a temporary directory and change into it.
	tempDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Prepare arguments to show version.
	args := []string{"tempo", "--version"}

	// Capture stdout.
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = runApp(args)
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tempo version") {
		t.Errorf("Expected version string in output, got: %s", output)
	}
}

// TestRunApp_Help tests the top-level help output.
func TestRunApp_Help(t *testing.T) {
	tempDir := t.TempDir()
	// Change into tempDir.
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tempDir)

	args := []string{"tempo", "--help"}

	output, err := testutils.CaptureStdout(func() {
		if err := runApp(args); err != nil {
			t.Fatalf("runApp returned error: %v", err)
		}
	})
	if err != nil {
		t.Fatalf("Failed to capture output: %v", err)
	}
	testutils.ValidateCLIOutput(t, output, []string{"USAGE:", "COMMANDS:"})
}

// TestRunApp_InitAutoGen verifies that when no config file exists, the "init" command auto-generates one.
func TestRunApp_InitAutoGen(t *testing.T) {
	// Create a temporary directory that does not contain a config file.
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "tempo.yaml")
	// Ensure the config file does not exist.
	os.Remove(configPath)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer os.Chdir(origDir)
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Run the init command to trigger auto-generation.
	// The --base-folder flag tells the init command where to create tempo.yaml.
	args := []string{"tempo", "init", "--base-folder", tempDir}

	// Capture both stdout and stderr.
	oldOut, oldErr := os.Stdout, os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	err = runApp(args)

	wOut.Close()
	wErr.Close()
	var bufOut, bufErr bytes.Buffer
	io.Copy(&bufOut, rOut)
	io.Copy(&bufErr, rErr)
	os.Stdout = oldOut
	os.Stderr = oldErr

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify that the config file now exists.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Expected tempo.yaml to be auto-generated, but it does not exist")
	}

	// In CI the log output may be suppressed; if available, check for expected substrings.
	output := bufOut.String() + bufErr.String()
	if output != "" {
		// Optionally, check for messages like "Generating tempo.yaml" and "Done! Customize it to match your project needs."
		if !strings.Contains(output, "Generating") || !strings.Contains(output, "Done!") {
			t.Logf("Output did not contain expected messages: %s", output)
		}
	} else {
		t.Log("No output captured; auto-generation messages may be logged elsewhere in CI.")
	}
}
