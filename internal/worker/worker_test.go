package worker

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/indaco/tempo/internal/utils"
)

/* ------------------------------------------------------------------------- */
/* Test WorkerPool Skipped & Processed Files                                 */
/* ------------------------------------------------------------------------- */

// TestWorkerPool_SkippedAndProcessedFiles ensures:
// - CSS & JS files are processed correctly
// - Non-CSS/JS files are skipped
// - Metrics reflect the skipped & processed files correctly
func TestWorkerPool_SkippedAndProcessedFiles(t *testing.T) {
	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	outputDir := filepath.Join(tempDir, "output")

	// Step 1: Create input folder structure with various file types
	files := map[string]string{
		"button.css":     ".button { color: blue; }",
		"script.js":      "console.log('Hello World');",
		"ignored.txt":    "This should be skipped",
		"ignored.json":   `{"key": "value"}`,
		"ignored.html":   "<p>Hello</p>",
		"ignored.random": "Random file type",
	}

	for filename, content := range files {
		filePath := filepath.Join(inputDir, filename)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("Failed to create directories: %v", err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}
	}

	// Step 2: Create expected ".templ" output files
	expectedOutputFiles := map[string]string{
		"button.templ": "",
		"script.templ": "",
	}

	for filename, content := range expectedOutputFiles {
		filePath := filepath.Join(outputDir, filename)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			t.Fatalf("Failed to create directories for templ files: %v", err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create templ file %s: %v", filename, err)
		}
	}

	// Step 3: Setup WorkerPool Manager with Timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	opts := WorkerPoolOptions{
		Context:      ctx,
		InputDir:     inputDir,
		OutputDir:    outputDir,
		MarkerName:   "tempo",
		NumWorkers:   2,
		IsProduction: false,
	}

	manager := NewWorkerPoolManager(opts)

	// Step 3: Start WorkerPool in a separate goroutine
	done := make(chan struct{})
	go func() {
		_ = manager.StartWorkers(ctx, opts.NumWorkers, false)
		close(done)
	}()

	// Step 4: Add jobs to queue
	for filename := range files {
		inputPath := filepath.Join(inputDir, filename)

		// Only CSS/JS should have corresponding .templ files
		if filepath.Ext(filename) == ".css" || filepath.Ext(filename) == ".js" {
			outputPath := filepath.Join(outputDir, utils.ToTemplFilename(filename))
			manager.JobChan <- Job{InputPath: inputPath, OutputPath: outputPath}
		} else {
			// Ensure skipped files are properly counted
			select {
			case <-ctx.Done():
				return

			case manager.SkippedChan <- FormatSkipReason(inputPath, "Unsupported file type (not CSS or JS)", SkipUnsupportedFile):
			}
			manager.Metrics.IncrementSkippedFile()
		}
	}

	// Step 5: Close job channel after sending all jobs
	close(manager.JobChan)

	// Step 6: Ensure worker pool exits correctly
	select {
	case <-done:
		t.Log("[DEBUG] WorkerPool completed successfully")
	case <-time.After(5 * time.Second):
		t.Fatal("[ERROR] WorkerPool did not complete within timeout")
	}

	// Step 7: Collect Skipped Files BEFORE Validating
	collectedSkipped := CollectErrors(manager.SkippedChan)

	// Step 8: Validate skipped files
	validateSkippedFiles(t, outputDir, collectedSkipped, manager.Metrics)
}

/* ------------------------------------------------------------------------- */
/* Helper Functions                                                          */
/* ------------------------------------------------------------------------- */

// validateSkippedFiles ensures skipped files were not written to output and validates metrics.
func validateSkippedFiles(t *testing.T, outputDir string, collectedSkipped []ProcessingError, metrics *Metrics) {
	expectedSkippedFiles := []string{"ignored.txt", "ignored.json", "ignored.html", "ignored.random"}

	for _, file := range expectedSkippedFiles {
		outputPath := filepath.Join(outputDir, file+".templ")
		if _, err := os.Stat(outputPath); err == nil {
			t.Errorf("File %s should have been skipped but was found in output", outputPath)
		}
	}

	// Validate Skipped File Metrics
	expectedSkippedCount := len(expectedSkippedFiles)
	if metrics.SkippedFiles != expectedSkippedCount {
		t.Errorf("Expected %d skipped files, but got %d", expectedSkippedCount, metrics.SkippedFiles)
	}

	// Validate skipped file processing reasons
	if len(collectedSkipped) != expectedSkippedCount {
		t.Errorf("Expected %d skipped files, but skipped list contains %d", expectedSkippedCount, len(collectedSkipped))
	}

	// Ensure skip reasons match expectations
	for _, skipEntry := range collectedSkipped {
		if skipEntry.Reason != "Unsupported file type (not CSS or JS)" {
			t.Errorf("Unexpected skip reason: %s", skipEntry.Reason)
		}
	}
}
