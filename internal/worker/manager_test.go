package worker

import (
	"context"
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
		"button.css": "Missing corresponding .templ file in output directory",
		"script.js":  "Missing corresponding .templ file in output directory",
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
