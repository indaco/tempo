package worker

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

/* ------------------------------------------------------------------------- */
/* Test WorkerPoolManager Context Cancellation                               */
/* ------------------------------------------------------------------------- */

func TestWorkerPoolManager_ContextCancellation(t *testing.T) {
	opts := WorkerPoolOptions{
		Context:      context.Background(),
		InputDir:     "/mock/input",
		OutputDir:    "/mock/output",
		MarkerName:   "tempo",
		NumWorkers:   2,
		IsProduction: false,
	}

	manager := NewWorkerPoolManager(opts)
	ctx, cancel := context.WithCancel(opts.Context)

	// Start workers in a separate goroutine
	done := make(chan struct{})
	go func() {
		_ = manager.StartWorkers(ctx, opts.NumWorkers, false)
		close(done)
	}()

	// Cancel the context after a short delay
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Ensure workers exit properly
	<-done
	t.Log("[DEBUG] Workers exited after context cancellation")
}

/* ------------------------------------------------------------------------- */
/* Test WorkerPoolManager Processes Jobs                                     */
/* ------------------------------------------------------------------------- */

func TestWorkerPoolManager_ProcessesJobs(t *testing.T) {
	opts := WorkerPoolOptions{
		Context:      context.Background(),
		InputDir:     "/mock/input",
		OutputDir:    "/mock/output",
		MarkerName:   "tempo",
		NumWorkers:   2,
		IsProduction: false,
	}

	manager := NewWorkerPoolManager(opts)
	ctx, cancel := context.WithCancel(opts.Context)
	defer cancel()

	// Start workers in a separate goroutine
	done := make(chan struct{})
	go func() {
		_ = manager.StartWorkers(ctx, opts.NumWorkers, false)
		close(done)
	}()

	// Simulate job processing
	manager.JobChan <- Job{InputPath: "/mock/input/button.css", OutputPath: "/mock/output/button.templ"}
	manager.JobChan <- Job{InputPath: "/mock/input/script.js", OutputPath: "/mock/output/script.templ"}

	// Close JobChan to allow workers to exit
	close(manager.JobChan)

	// Wait for workers to finish
	<-done

	// Ensure ErrorsChan and SkippedChan are closed before collecting results
	close(manager.ErrorsChan)
	close(manager.SkippedChan)

	// Collect skipped files
	skippedFiles := CollectErrors(manager.SkippedChan)

	// Validate expected skipped files
	expectedSkipped := map[string]string{
		filepath.Join(opts.InputDir, "button.css"): "Missing corresponding .templ file in output directory",
		filepath.Join(opts.InputDir, "script.js"):  "Missing corresponding .templ file in output directory",
	}

	if len(skippedFiles) != len(expectedSkipped) {
		t.Errorf("Expected %d skipped files, but got %d", len(expectedSkipped), len(skippedFiles))
	}

	for _, skipped := range skippedFiles {
		expectedReason, exists := expectedSkipped[skipped.Source]
		if !exists {
			t.Errorf("Unexpected skipped file: %s", skipped.Source)
		} else if skipped.Reason != expectedReason {
			t.Errorf("Unexpected skip reason for %s. Expected: %s, Got: %s",
				skipped.Source, expectedReason, skipped.Reason)
		}
	}

	t.Log("[DEBUG] Workers processed jobs correctly, skipped files tracked successfully")
}

func TestWorkerPoolManager_HandlesEmptyJobQueue(t *testing.T) {
	opts := WorkerPoolOptions{
		Context:      context.Background(),
		InputDir:     "/mock/input",
		OutputDir:    "/mock/output",
		MarkerName:   "tempo",
		NumWorkers:   2,
		IsProduction: false,
	}

	manager := NewWorkerPoolManager(opts)
	ctx, cancel := context.WithCancel(opts.Context)
	defer cancel()

	// Start workers in a separate goroutine
	done := make(chan struct{})
	go func() {
		_ = manager.StartWorkers(ctx, opts.NumWorkers, false)
		close(done)
	}()

	// Close job channel immediately (no jobs added)
	close(manager.JobChan)

	// Wait for workers to complete
	<-done

	t.Log("[DEBUG] Workers exited cleanly with an empty job queue")
}

func TestWorkerPoolManager_ContextTimeout(t *testing.T) {
	opts := WorkerPoolOptions{
		Context:      context.Background(),
		InputDir:     "/mock/input",
		OutputDir:    "/mock/output",
		MarkerName:   "tempo",
		NumWorkers:   2,
		IsProduction: false,
	}

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(opts.Context, 100*time.Millisecond)
	defer cancel()

	manager := NewWorkerPoolManager(opts)

	// Start workers
	done := make(chan struct{})
	go func() {
		_ = manager.StartWorkers(ctx, opts.NumWorkers, false)
		close(done)
	}()

	// Wait for workers to exit
	<-done

	t.Log("[DEBUG] Workers exited after context timeout")
}

func TestWorkerPoolManager_InitializationErrors(t *testing.T) {
	manager := &WorkerPoolManager{}

	ctx := context.Background()
	err := manager.StartWorkers(ctx, 2, false)

	if err == nil {
		t.Fatal("Expected error when starting uninitialized WorkerPoolManager, but got nil")
	}

	t.Log("[DEBUG] WorkerPoolManager initialization failed as expected")
}

/* ------------------------------------------------------------------------- */
/* Test NewWorkerPoolOptions                                                 */
/* ------------------------------------------------------------------------- */

func TestNewWorkerPoolOptions(t *testing.T) {
	ctx := context.Background()

	t.Run("defaults applied when no options given", func(t *testing.T) {
		opts, err := NewWorkerPoolOptions(ctx, "/input", "/output")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.Context != ctx {
			t.Errorf("expected context to be set")
		}
		if opts.InputDir != "/input" {
			t.Errorf("expected InputDir=/input, got %s", opts.InputDir)
		}
		if opts.OutputDir != "/output" {
			t.Errorf("expected OutputDir=/output, got %s", opts.OutputDir)
		}
		if opts.NumWorkers <= 0 {
			t.Errorf("expected positive default NumWorkers, got %d", opts.NumWorkers)
		}
	})

	t.Run("functional options override defaults", func(t *testing.T) {
		opts, err := NewWorkerPoolOptions(ctx, "/input", "/output",
			WithExcludeDir("/input/exclude"),
			WithMarkerName("custom-marker"),
			WithNumWorkers(4),
			WithProduction(true),
			WithForce(true),
			WithTrackExecutionTime(true),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if opts.ExcludeDir != "/input/exclude" {
			t.Errorf("expected ExcludeDir=/input/exclude, got %s", opts.ExcludeDir)
		}
		if opts.MarkerName != "custom-marker" {
			t.Errorf("expected MarkerName=custom-marker, got %s", opts.MarkerName)
		}
		if opts.NumWorkers != 4 {
			t.Errorf("expected NumWorkers=4, got %d", opts.NumWorkers)
		}
		if !opts.IsProduction {
			t.Errorf("expected IsProduction=true")
		}
		if !opts.IsForce {
			t.Errorf("expected IsForce=true")
		}
		if !opts.IsTrackExecutionTime {
			t.Errorf("expected IsTrackExecutionTime=true")
		}
	})

	t.Run("validation rejects zero workers", func(t *testing.T) {
		_, err := NewWorkerPoolOptions(ctx, "/input", "/output", WithNumWorkers(0))
		if err == nil {
			t.Fatal("expected error for NumWorkers=0, got nil")
		}
	})

	t.Run("validation rejects negative workers", func(t *testing.T) {
		_, err := NewWorkerPoolOptions(ctx, "/input", "/output", WithNumWorkers(-1))
		if err == nil {
			t.Fatal("expected error for NumWorkers=-1, got nil")
		}
	})
}

func TestNewWorkerPoolManager_PanicsOnInvalidNumWorkers(t *testing.T) {
	for _, numWorkers := range []int{0, -1} {
		t.Run(fmt.Sprintf("NumWorkers=%d", numWorkers), func(t *testing.T) {
			opts := WorkerPoolOptions{
				InputDir:   "/input",
				OutputDir:  "/output",
				NumWorkers: numWorkers,
			}
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic for NumWorkers=%d, got none", numWorkers)
				}
			}()
			NewWorkerPoolManager(opts)
		})
	}
}
