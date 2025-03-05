package testhelpers

import (
	"bytes"
	"os"

	"github.com/fatih/color"
)

// CaptureStdout captures all writes to os.Stdout during the execution of the provided function.
// It returns the captured output as a string and restores os.Stdout to its original state afterward.
func CaptureStdout(f func()) (string, error) {
	color.NoColor = true // Disable colors for testing

	// Save the original stdout and color output
	origStdout := os.Stdout
	origColorOutput := color.Output

	// Create a pipe to capture stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}
	os.Stdout = w
	color.Output = w // Redirect color output to the pipe

	// Create a channel to capture the output asynchronously
	outputChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		outputChan <- buf.String()
	}()

	// Execute the function
	f()

	// Close the pipe and restore stdout and color output
	w.Close()
	os.Stdout = origStdout
	color.Output = origColorOutput

	// Retrieve the captured output
	output := <-outputChan
	return output, nil
}
