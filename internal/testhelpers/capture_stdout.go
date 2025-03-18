package testhelpers

import (
	"bytes"
	"os"

	"github.com/fatih/color"
)

// CaptureStdout captures all writes to os.Stdout and os.Stderr during the execution of the provided function.
// It returns the combined captured output as a string and restores os.Stdout and os.Stderr to their original states.
func CaptureStdout(f func()) (string, error) {
	color.NoColor = true // Disable colors for testing

	// Save original stdout, stderr, and color output
	origStdout, origStderr := os.Stdout, os.Stderr
	origColorOutput := color.Output

	// Create pipes to capture stdout and stderr
	rOut, wOut, err := os.Pipe()
	if err != nil {
		return "", err
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		return "", err
	}

	// Redirect output
	os.Stdout, os.Stderr = wOut, wErr
	color.Output = wOut // Redirect color output to the pipe

	// Capture output concurrently
	outputChan := make(chan string)
	go func() {
		var bufOut, bufErr bytes.Buffer
		_, _ = bufOut.ReadFrom(rOut)
		_, _ = bufErr.ReadFrom(rErr)
		outputChan <- bufOut.String() + bufErr.String()
	}()

	// Execute the function
	f()

	// Close pipes and restore output
	wOut.Close()
	wErr.Close()
	os.Stdout, os.Stderr = origStdout, origStderr
	color.Output = origColorOutput

	// Retrieve captured output
	output := <-outputChan
	return output, nil
}
