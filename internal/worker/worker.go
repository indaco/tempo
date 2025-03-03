package worker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/indaco/tempo/internal/processor"
	"github.com/indaco/tempo/internal/utils"
)

// WorkerPool processes files concurrently and updates metrics.
func WorkerPool(ctx context.Context, m *WorkerPoolManager, trackExecution bool) error {
	for {
		select {
		case <-ctx.Done(): // Exit when context is canceled
			return nil
		case job, ok := <-m.JobChan:
			if !ok {
				return nil
			}

			if skipReason, skipType := shouldSkipFile(job, m.InputDir, m.OutputDir); skipReason != "" {
				m.Metrics.IncrementSkippedFile()
				select {
				case m.SkippedChan <- FormatSkipReason(job.InputPath, skipReason, skipType):
				default:
				}
				continue
			}

			if err := processFile(job, m, trackExecution); err != nil {
				m.Metrics.IncrementError()
				select {
				case m.ErrorsChan <- FormatError(job.InputPath, err):
				default:
				}
				continue
			}

			m.Metrics.IncrementFile()
		}
	}
}

/* ------------------------------------------------------------------------- */
/* HELPER FUNCTIONS                                                          */
/* ------------------------------------------------------------------------- */

// shouldSkipFile checks if a file should be skipped and returns the reason.
func shouldSkipFile(job Job, inputDir, outputDir string) (string, SkipType) {
	ext := filepath.Ext(job.InputPath)

	// Unsupported file type
	if processor.GetLoader(ext) == api.LoaderNone {
		return "Unsupported file type (not CSS or JS)", SkipUnsupportedFile
	}

	// Ensure output structure matches expectations
	expectedOutput := utils.RebasePathToOutput(job.InputPath, inputDir, outputDir)
	// Ensure the expected `.templ` file actually exists
	if _, err := os.Stat(expectedOutput); os.IsNotExist(err) {
		return "Missing corresponding .templ file in output directory", SkipMissingTemplFile
	}

	// Validate output path
	if !isValidOutputPath(job.OutputPath, expectedOutput) {
		return "Mismatched output structure", SkipMismatchedPath
	}

	return "", "" // No skipping required
}

// processFile processes a single job and optionally tracks execution time.
func processFile(job Job, m *WorkerPoolManager, trackExecution bool) error {
	start := time.Now()

	processor := m.Factory.GetProcessor(job.InputPath)
	err := processor.Process(job.InputPath, job.OutputPath, m.MarkerName)

	// Ensure execution time tracking is recorded
	if trackExecution {
		recordExecutionTime(m, job.InputPath, time.Since(start))
	}

	return err
}

// recordExecutionTime safely stores job execution time in WorkerPoolManager.
func recordExecutionTime(m *WorkerPoolManager, filePath string, duration time.Duration) {
	m.mu.Lock()
	m.ExecutionTimes = append(m.ExecutionTimes, JobExecutionTime{FilePath: filePath, Duration: duration})
	m.mu.Unlock()

	fmt.Printf("Processed %s (took %v)\n", filePath, duration)
}

// isValidOutputPath checks if the generated output path matches expectations.
func isValidOutputPath(actual, expected string) bool {
	return actual == expected
}
