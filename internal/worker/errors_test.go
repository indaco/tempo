package worker

import (
	"fmt"
	"testing"

	"github.com/indaco/tempo/internal/testhelpers"
	"github.com/indaco/tempo/internal/utils"
)

func TestCollectErrors(t *testing.T) {
	// Test case 1: errorsChan with values
	t.Run("With values in channel", func(t *testing.T) {
		errorsChan := make(chan ProcessingError, 3)
		expectedErrors := []ProcessingError{
			{Source: "file1.js", Message: "Syntax error"},
			{Source: "file2.css", Message: "Invalid property"},
			{Source: "file3.txt", Reason: "Unsupported file type", SkipType: SkipUnsupportedFile},
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
	})

	// Test case 2: errorsChan is nil
	t.Run("Nil errorsChan", func(t *testing.T) {
		var errorsChan <-chan ProcessingError = nil

		collectedErrors := CollectErrors(errorsChan)

		if len(collectedErrors) != 0 {
			t.Errorf("Expected 0 errors, got %d", len(collectedErrors))
		}
	})
}

func TestFormatError(t *testing.T) {
	err := FormatError("broken.js", fmt.Errorf("Unexpected token"))

	if err.Source != "broken.js" {
		t.Errorf("Expected FilePath to be 'broken.js', got %s", err.Source)
	}
	if err.Message != "Unexpected token" {
		t.Errorf("Expected Message to be 'Unexpected token', got %s", err.Message)
	}
}

func TestFormatSkipReason(t *testing.T) {
	//cwd := utils.GetCWD()

	tests := []struct {
		name        string
		skipped     SkippedFile
		expSource   string
		expDest     string
		expReason   string
		expSkipType SkipType
	}{
		{
			name: "Standard case",
			skipped: SkippedFile{
				Source:    "assets/js/script.js",
				Dest:      "components/js/script.templ",
				InputDir:  "assets",
				OutputDir: "components",
				Reason:    "Missing corresponding .templ file in output directory",
				SkipType:  SkipMissingTemplFile,
			},
			expSource:   "assets/js/script.js",
			expDest:     "components/js/script.templ",
			expReason:   "Missing corresponding .templ file in output directory",
			expSkipType: SkipMissingTemplFile,
		},
		{
			name: "OutputDir is the current working directory (cwd)",
			skipped: SkippedFile{
				Source:    "assets/js/script.js",
				Dest:      "js/script.templ",
				InputDir:  "assets",
				OutputDir: utils.GetCWD(),
				Reason:    "Missing corresponding .templ file in output directory",
				SkipType:  SkipMissingTemplFile,
			},
			expSource:   "assets/js/script.js",
			expDest:     "js/script.templ",
			expReason:   "Missing corresponding .templ file in output directory",
			expSkipType: SkipMissingTemplFile,
		},
		{
			name: "OutputDir is the current working directory (.)",
			skipped: SkippedFile{
				Source:    "assets/js/script.js",
				Dest:      "js/script.templ",
				InputDir:  "assets",
				OutputDir: ".",
				Reason:    "Missing corresponding .templ file in output directory",
				SkipType:  SkipMissingTemplFile,
			},
			expSource:   "assets/js/script.js",
			expDest:     "js/script.templ",
			expReason:   "Missing corresponding .templ file in output directory",
			expSkipType: SkipMissingTemplFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skip := FormatSkipReason(tt.skipped)

			if skip.Source != tt.expSource {
				t.Errorf("Expected Source to be '%s', got '%s'", tt.expSource, skip.Source)
			}
			if skip.Dest != tt.expDest {
				t.Errorf("Expected Dest to be '%s', got '%s'", tt.expDest, skip.Dest)
			}
			if skip.Reason != tt.expReason {
				t.Errorf("Expected Reason to be '%s', got '%s'", tt.expReason, skip.Reason)
			}
			if skip.SkipType != tt.expSkipType {
				t.Errorf("Expected SkipType to be '%s', got '%s'", tt.expSkipType, skip.SkipType)
			}
		})
	}
}

func TestPrintErrors(t *testing.T) {
	// Test case 1: errors slice is not empty
	t.Run("No errors", func(t *testing.T) {
		errors := []ProcessingError{
			{Source: "file1.js", Message: "Syntax error"},
			{Source: "file2.css", Message: "Invalid property"},
			{Source: "ignored.txt", Reason: "Unsupported file type", SkipType: SkipUnsupportedFile},
		}

		// Capture output when there are errors
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
	})

	// Test case 2:  errors slice is empty
	t.Run("No errors", func(t *testing.T) {
		// Capture output when there are no errors
		output, err := testhelpers.CaptureStdout(func() {
			PrintErrors([]ProcessingError{}) // Empty slice
		})
		if err != nil {
			t.Fatalf("Failed to capture stdout: %v", err)
		}

		// Validate that no output is produced
		if output != "" {
			t.Errorf("Expected no output, but got: %s", output)
		}
	})
}
