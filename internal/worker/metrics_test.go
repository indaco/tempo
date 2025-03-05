package worker

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/indaco/tempo/internal/testhelpers"
)

// TestMetrics_IncrementCounters ensures that metrics are updated correctly.
func TestMetrics_IncrementCounters(t *testing.T) {
	m := NewMetrics()

	// Simulate processing
	m.IncrementFile()
	m.IncrementFile()
	m.IncrementDirectory()
	m.IncrementError()
	m.IncrementSkippedFile()
	m.IncrementSkippedFile()

	// Validate results
	if m.FilesProcessed != 2 {
		t.Errorf("Expected 2 files processed, got %d", m.FilesProcessed)
	}
	if m.DirectoriesProcessed != 1 {
		t.Errorf("Expected 1 directory processed, got %d", m.DirectoriesProcessed)
	}
	if m.ErrorsEncountered != 1 {
		t.Errorf("Expected 1 error encountered, got %d", m.ErrorsEncountered)
	}
	if m.SkippedFiles != 2 {
		t.Errorf("Expected 2 skipped files, got %d", m.SkippedFiles)
	}
}

// TestMetrics_ExportToJSON validates that the exported JSON includes all fields correctly.
func TestMetrics_ExportToJSON(t *testing.T) {
	m := NewMetrics()
	m.IncrementFile()
	m.IncrementDirectory()
	m.IncrementError()
	m.IncrementSkippedFile()

	errors := []ProcessingError{
		{FilePath: "error1.js", Message: "Syntax error"},
		{FilePath: "error2.css", Message: "Invalid property"},
	}

	skippedFiles := []ProcessingError{
		{FilePath: "ignored.txt", Reason: "Unsupported file type", SkipType: SkipUnsupportedFile},
		{FilePath: "ignored.json", Reason: "Unsupported file type", SkipType: SkipUnsupportedFile},
	}

	// Create temp file
	tempFile, err := os.CreateTemp("", "metrics_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Export metrics to JSON
	if err := m.ToJSONFile(errors, skippedFiles, tempFile.Name()); err != nil {
		t.Fatalf("Failed to export metrics to JSON: %v", err)
	}

	// Read and validate JSON file
	data, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp JSON file: %v", err)
	}

	var parsed struct {
		Metrics struct {
			FilesProcessed       int `json:"files_processed"`
			DirectoriesProcessed int `json:"directories_processed"`
			ErrorsEncountered    int `json:"errors_encountered"`
			SkippedFiles         int `json:"skipped_files"`
		} `json:"metrics"`
		Errors       []ProcessingError            `json:"errors"`
		SkippedFiles map[string][]ProcessingError `json:"skipped_files"`
	}

	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Validate values
	if parsed.Metrics.FilesProcessed != 1 {
		t.Errorf("Expected 1 file processed, got %d", parsed.Metrics.FilesProcessed)
	}
	if parsed.Metrics.DirectoriesProcessed != 1 {
		t.Errorf("Expected 1 directory processed, got %d", parsed.Metrics.DirectoriesProcessed)
	}
	if parsed.Metrics.ErrorsEncountered != 1 {
		t.Errorf("Expected 1 error encountered, got %d", parsed.Metrics.ErrorsEncountered)
	}
	if parsed.Metrics.SkippedFiles != 1 {
		t.Errorf("Expected 1 skipped file, got %d", parsed.Metrics.SkippedFiles)
	}
	if len(parsed.Errors) != 2 {
		t.Errorf("Expected 2 errors in JSON, got %d", len(parsed.Errors))
	}
	if len(parsed.SkippedFiles["unsupported_file"]) != 2 {
		t.Errorf("Expected 2 unsupported skipped files, got %d", len(parsed.SkippedFiles["unsupported_file"]))
	}
}

// TestMetrics_PrintSummary ensures the summary output contains the expected values.
func TestMetrics_PrintSummary(t *testing.T) {
	m := NewMetrics()
	m.IncrementFile()
	m.IncrementDirectory()
	m.IncrementError()
	m.IncrementSkippedFile()

	errors := []ProcessingError{
		{FilePath: "error1.js", Message: "Syntax error"},
		{FilePath: "error2.css", Message: "Invalid property"},
	}

	skippedFiles := []ProcessingError{
		{FilePath: "ignored.txt", Reason: "Unsupported file type", SkipType: SkipUnsupportedFile},
		{FilePath: "ignored.json", Reason: "Unsupported file type", SkipType: SkipUnsupportedFile},
		{FilePath: "output/styles.css", Reason: "Mismatched output structure", SkipType: SkipMismatchedPath},
	}

	// Capture the output
	output, err := testhelpers.CaptureStdout(func() {
		m.PrintSummary(errors, skippedFiles, true)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Expected output messages
	expectedMessages := []string{
		"📋 Processing Summary:",
		"- Total files processed: 1",
		"- Total directories processed: 1",
		"- Total errors encountered: 1",
		"- Total skipped files: 1",
		"Elapsed time:",
		"📌 Skipped Files Breakdown:",
		"• Reason: Unsupported File Types",
		"- file: ignored.txt",
		"- file: ignored.json",
		"• Reason: Mismatched Output Structure",
		"- file: output/styles.css",
	}

	// Validate output
	testhelpers.ValidateCLIOutput(t, output, expectedMessages)
}

// TestMetrics_Reset ensures all metrics are correctly reset.
func TestMetrics_Reset(t *testing.T) {
	m := NewMetrics()

	// Simulate processing
	m.IncrementFile()
	m.IncrementDirectory()
	m.IncrementError()
	m.IncrementSkippedFile()

	// Ensure metrics are incremented
	if m.FilesProcessed != 1 || m.DirectoriesProcessed != 1 || m.ErrorsEncountered != 1 || m.SkippedFiles != 1 {
		t.Fatalf("Metrics not incremented properly before reset")
	}

	// Perform reset
	m.Reset()

	// Ensure all values are reset
	if m.FilesProcessed != 0 {
		t.Errorf("Expected FilesProcessed to be 0 after reset, got %d", m.FilesProcessed)
	}
	if m.DirectoriesProcessed != 0 {
		t.Errorf("Expected DirectoriesProcessed to be 0 after reset, got %d", m.DirectoriesProcessed)
	}
	if m.ErrorsEncountered != 0 {
		t.Errorf("Expected ErrorsEncountered to be 0 after reset, got %d", m.ErrorsEncountered)
	}
	if m.SkippedFiles != 0 {
		t.Errorf("Expected SkippedFiles to be 0 after reset, got %d", m.SkippedFiles)
	}
	if m.ElapsedTime != "" {
		t.Errorf("Expected ElapsedTime to be empty after reset, got %s", m.ElapsedTime)
	}
	if m.StartTime.After(time.Now()) {
		t.Errorf("Expected StartTime to be reset to a recent value")
	}
}

func TestSummaryAsText(t *testing.T) {
	metrics := &Metrics{
		FilesProcessed:       10,
		DirectoriesProcessed: 2,
		ErrorsEncountered:    1,
		SkippedFiles:         3,
		ElapsedTime:          "2.345s",
	}

	skippedFiles := []ProcessingError{
		{FilePath: "assets/styles.css", Reason: "Unsupported file type", SkipType: SkipUnsupportedFile},
		{FilePath: "output/script.js", Reason: "Mismatched output structure", SkipType: SkipMismatchedPath},
		{FilePath: "input/template.templ", Reason: "Unchanged file", SkipType: SkipUnchangedFile},
	}

	expectedOutput := `
📋 Processing Summary:
  - Total files processed: 10
  - Total directories processed: 2
  - Total errors encountered: 1
  - Total skipped files: 3
  - Elapsed time: 2.345s

📌 Skipped Files Breakdown:

  • Reason: Unsupported File Types
    (Hint: Only CSS and JS files are supported. Ensure your files have valid extensions and are placed correctly.)
    - file: assets/styles.css

  • Reason: Mismatched Output Structure
    (Hint: The input folder structure must mirror the output structure. Check your file paths or adjust the configuration.)
    - file: output/script.js

  • Reason: Unchanged Files
    (Hint: These files haven't changed since the last run. Use '--force' to process them anyway if needed.)
    - file: input/template.templ

✘ Some errors occurred. Check logs for details.
`
	result := metrics.summaryAsText(skippedFiles, true, false)

	if strings.TrimSpace(result) != strings.TrimSpace(expectedOutput) {
		t.Errorf("Expected summary output:\n%s\nGot:\n%s", expectedOutput, result)
	}
}
