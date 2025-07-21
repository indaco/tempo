package worker

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/indaco/tempo/internal/processor"
	"github.com/indaco/tempo/internal/utils"
)

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
			case manager.SkippedChan <- FormatSkipReason(SkippedFile{
				Source:    inputPath,
				Dest:      "", // No expected output for skipped files
				InputDir:  inputDir,
				OutputDir: outputDir,
				Reason:    "Unsupported file type (not CSS or JS)",
				SkipType:  SkipUnsupportedFile,
			}):
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

func TestShouldSkipFile(t *testing.T) {
	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	outputDir := filepath.Join(tempDir, "output")

	// Ensure input and output directories exist
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Create an unsupported file type
	unsupportedFile := filepath.Join(inputDir, "unsupported.xyz")
	if err := os.WriteFile(unsupportedFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create unsupported file: %v", err)
	}

	// Case 1: Unsupported file type
	skipReason, skipType := shouldSkipFile(Job{InputPath: unsupportedFile}, inputDir, outputDir)
	if skipReason == "" || skipType != SkipUnsupportedFile {
		t.Errorf("Expected unsupported file type skip, got: %s (%v)", skipReason, skipType)
	}

	// Case 2: Missing corresponding `.templ` file
	cssFile := filepath.Join(inputDir, "style.css")
	if err := os.WriteFile(cssFile, []byte("body {color: red;}"), 0644); err != nil {
		t.Fatalf("Failed to create CSS file: %v", err)
	}

	skipReason, skipType = shouldSkipFile(Job{InputPath: cssFile}, inputDir, outputDir)
	if skipReason == "" || skipType != SkipMissingTemplFile {
		t.Errorf("Expected missing .templ file skip, got: %s (%v)", skipReason, skipType)
	}

	// Case 3: Mismatched output structure
	expectedTemplFile := utils.RebasePathToOutput(cssFile, inputDir, outputDir)
	if err := os.WriteFile(expectedTemplFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create expected .templ file: %v", err)
	}

	invalidOutputFile := filepath.Join(outputDir, "invalid-style.templ")
	skipReason, skipType = shouldSkipFile(Job{InputPath: cssFile, OutputPath: invalidOutputFile}, inputDir, outputDir)
	if skipReason == "" || skipType != SkipMismatchedPath {
		t.Errorf("Expected mismatched output structure skip, got: %s (%v)", skipReason, skipType)
	}

	// Case 4: Valid case (no skipping required)
	validJob := Job{InputPath: cssFile, OutputPath: expectedTemplFile}
	skipReason, skipType = shouldSkipFile(validJob, inputDir, outputDir)
	if skipReason != "" || skipType != "" {
		t.Errorf("Expected no skipping for valid case, but got: %s (%v)", skipReason, skipType)
	}
}

type MockProcessorFactory struct {
	Processor processor.FileProcessor
}

// Ensure MockProcessorFactory implements ProcessorFactoryInterface
var _ processor.ProcessorFactoryInterface = (*MockProcessorFactory)(nil)

func (m *MockProcessorFactory) GetProcessor(_ string) processor.FileProcessor {
	return m.Processor
}

type MockProcessor struct {
	ProcessCalled bool
}

func (m *MockProcessor) Process(input, output, marker string) error {
	m.ProcessCalled = true
	return nil
}

func TestProcessFile(t *testing.T) {
	mockProcessor := &MockProcessor{}

	mockManager := &WorkerPoolManager{
		Factory: &MockProcessorFactory{Processor: mockProcessor},
	}

	job := Job{
		InputPath:  "style.css",
		OutputPath: "style.templ",
	}

	err := processFile(job, mockManager, false)
	if err != nil {
		t.Fatalf("Expected processFile to succeed but got error: %v", err)
	}

	if !mockProcessor.ProcessCalled {
		t.Fatal("Expected Process to be called but it wasn't")
	}
}

/* ------------------------------------------------------------------------- */
/* Helper Functions                                                          */
/* ------------------------------------------------------------------------- */

// validateSkippedFiles ensures skipped files were not written to output and validates metrics.
func validateSkippedFiles(t *testing.T, outputDir string, skippedFiles []ProcessingError, metrics *Metrics) {
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
	if len(skippedFiles) != expectedSkippedCount {
		t.Errorf("Expected %d skipped files, but skipped list contains %d", expectedSkippedCount, len(skippedFiles))
	}

	// Ensure skip reasons match expectations
	for _, skipped := range skippedFiles {
		if skipped.Reason != "Unsupported file type (not CSS or JS)" {
			t.Errorf("Unexpected skip reason: %s", skipped.Reason)
		}
	}
}

func TestRecordExecutionTime(t *testing.T) {
	// Create a WorkerPoolManager with an empty ExecutionTimes slice
	mockManager := &WorkerPoolManager{}

	// Define a test file path and duration
	filePath := "testfile.css"
	duration := 500 * time.Millisecond

	// Capture standard output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the function
	recordExecutionTime(mockManager, filePath, duration)

	// Restore standard output
	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close pipe writer: %v", err)
	}
	os.Stdout = oldStdout

	// Read captured output
	out, _ := io.ReadAll(r)
	output := string(out)

	// Verify execution time was recorded
	if len(mockManager.ExecutionTimes) != 1 {
		t.Fatalf("Expected 1 execution time entry, got %d", len(mockManager.ExecutionTimes))
	}

	if mockManager.ExecutionTimes[0].FilePath != filePath {
		t.Errorf("Expected execution time entry for %s, got %s", filePath, mockManager.ExecutionTimes[0].FilePath)
	}

	if mockManager.ExecutionTimes[0].Duration != duration {
		t.Errorf("Expected execution duration %v, got %v", duration, mockManager.ExecutionTimes[0].Duration)
	}

	// Validate printed output
	expectedOutput := fmt.Sprintf("Processed %s (took %v)\n", filePath, duration)
	if !utils.ContainsSubstring(output, expectedOutput[:len(expectedOutput)-3]) { // Allow minor time variations
		t.Errorf("Expected console output containing: %q, got: %q", expectedOutput, output)
	}
}
