package worker

import (
	"fmt"
	"testing"

	"github.com/indaco/tempo/internal/testhelpers"
)

func TestCollectErrors(t *testing.T) {
	errorsChan := make(chan ProcessingError, 3)
	expectedErrors := []ProcessingError{
		{FilePath: "file1.js", Message: "Syntax error"},
		{FilePath: "file2.css", Message: "Invalid property"},
		{FilePath: "file3.txt", Reason: "Unsupported file type", SkipType: SkipUnsupportedFile},
	}

	for _, err := range expectedErrors {
		errorsChan <- err
	}
	close(errorsChan)

	collectedErrors := CollectErrors(errorsChan)

	if len(collectedErrors) != len(expectedErrors) {
		t.Errorf("Expected %d errors, got %d", len(expectedErrors), len(collectedErrors))
	}

	for i, err := range collectedErrors {
		if err != expectedErrors[i] {
			t.Errorf("Mismatch at index %d: got %+v, expected %+v", i, err, expectedErrors[i])
		}
	}
}

func TestFormatError(t *testing.T) {
	err := FormatError("broken.js", fmt.Errorf("Unexpected token"))

	if err.FilePath != "broken.js" {
		t.Errorf("Expected FilePath to be 'broken.js', got %s", err.FilePath)
	}
	if err.Message != "Unexpected token" {
		t.Errorf("Expected Message to be 'Unexpected token', got %s", err.Message)
	}
}

func TestFormatSkipReason(t *testing.T) {
	skip := FormatSkipReason("invalid.txt", "Unsupported file type", SkipUnsupportedFile)

	if skip.FilePath != "invalid.txt" {
		t.Errorf("Expected FilePath to be 'invalid.txt', got %s", skip.FilePath)
	}
	if skip.Reason != "Unsupported file type" {
		t.Errorf("Expected Reason to be 'Unsupported file type', got %s", skip.Reason)
	}
	if skip.SkipType != SkipUnsupportedFile {
		t.Errorf("Expected SkipType to be '%s', got '%s'", SkipUnsupportedFile, skip.SkipType)
	}
}

func TestPrintErrors(t *testing.T) {

	errors := []ProcessingError{
		{FilePath: "file1.js", Message: "Syntax error"},
		{FilePath: "file2.css", Message: "Invalid property"},
		{FilePath: "ignored.txt", Reason: "Unsupported file type", SkipType: SkipUnsupportedFile},
	}

	// Capture output
	output, err := testhelpers.CaptureStdout(func() {
		PrintErrors(errors)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	expectedMessages := []string{
		"Errors encountered:",
		"- File: file1.js\n  Error: Syntax error",
		"- File: file2.css\n  Error: Invalid property",
		"- Skipped File: ignored.txt\n  Reason: Unsupported file type (Type: unsupported_file)",
	}

	testhelpers.ValidateCLIOutput(t, output, expectedMessages)
}
