package worker

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/indaco/tempo/internal/processor"
)

// BenchmarkWorkerPool_Throughput measures end-to-end worker pool throughput
// by processing a batch of CSS and JS files using the passthrough processor.
func BenchmarkWorkerPool_Throughput(b *testing.B) {
	dir := b.TempDir()
	inputDir := filepath.Join(dir, "input")
	outputDir := filepath.Join(dir, "output")

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		b.Fatalf("failed to create input dir: %v", err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		b.Fatalf("failed to create output dir: %v", err)
	}

	// Create a batch of input files and matching .templ output stubs.
	const numFiles = 10
	inputFiles := make([]string, 0, numFiles)
	outputFiles := make([]string, 0, numFiles)

	markerContent := "/* [tempo] BEGIN - Do not edit! This section is auto-generated. */\n/* [tempo] END */"
	cssContent := ".btn { color: blue; padding: 8px; }"

	for i := range numFiles {
		name := filepath.Join(dir, "component")
		if err := os.MkdirAll(name, 0755); err != nil {
			b.Fatalf("failed to create dir: %v", err)
		}
		in := filepath.Join(inputDir, filepath.FromSlash(
			"component/style"+string(rune('0'+i))+".css"))
		out := filepath.Join(outputDir, filepath.FromSlash(
			"component/style"+string(rune('0'+i))+".templ"))

		if err := os.MkdirAll(filepath.Dir(in), 0755); err != nil {
			b.Fatalf("failed to create dir: %v", err)
		}
		if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
			b.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(in, []byte(cssContent), 0644); err != nil {
			b.Fatalf("failed to write input: %v", err)
		}
		if err := os.WriteFile(out, []byte(markerContent), 0644); err != nil {
			b.Fatalf("failed to write output stub: %v", err)
		}
		inputFiles = append(inputFiles, in)
		outputFiles = append(outputFiles, out)
	}

	opts := WorkerPoolOptions{
		Context:      context.Background(),
		InputDir:     inputDir,
		OutputDir:    outputDir,
		MarkerName:   "tempo",
		NumWorkers:   4,
		IsProduction: false,
	}

	b.ResetTimer()

	for b.Loop() {
		// Restore output stubs so markers are present for each iteration.
		for _, out := range outputFiles {
			if err := os.WriteFile(out, []byte(markerContent), 0644); err != nil {
				b.Fatalf("failed to restore output stub: %v", err)
			}
		}

		manager := NewWorkerPoolManager(opts)

		// Enqueue all jobs.
		for i := range numFiles {
			manager.JobChan <- Job{
				InputPath:  inputFiles[i],
				OutputPath: outputFiles[i],
			}
		}
		close(manager.JobChan)

		if err := manager.StartWorkers(opts.Context, opts.NumWorkers, false); err != nil {
			b.Fatalf("StartWorkers failed: %v", err)
		}

		close(manager.ErrorsChan)
		close(manager.SkippedChan)
	}
}

// BenchmarkNewWorkerPoolOptions measures the cost of constructing WorkerPoolOptions
// via the functional options constructor.
func BenchmarkNewWorkerPoolOptions(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for b.Loop() {
		_, _ = NewWorkerPoolOptions(ctx, "/input", "/output",
			WithNumWorkers(4),
			WithMarkerName("tempo"),
			WithProduction(false),
			WithForce(false),
		)
	}
}

// BenchmarkWorkerPoolManager_Init measures the cost of NewWorkerPoolManager.
func BenchmarkWorkerPoolManager_Init(b *testing.B) {
	opts := WorkerPoolOptions{
		Context:      context.Background(),
		InputDir:     "/mock/input",
		OutputDir:    "/mock/output",
		MarkerName:   "tempo",
		NumWorkers:   4,
		IsProduction: false,
	}

	b.ResetTimer()
	for b.Loop() {
		m := NewWorkerPoolManager(opts)
		// Immediately close channels to avoid goroutine leaks.
		close(m.JobChan)
		close(m.ErrorsChan)
		close(m.SkippedChan)
		// Drain channels.
		for range m.JobChan {
		}
		for range m.ErrorsChan {
		}
		for range m.SkippedChan {
		}
	}
}

// MockPassthroughProcessor is a no-op processor used in benchmarks that
// bypasses real file I/O to isolate worker pool scheduling overhead.
type MockPassthroughProcessor struct{}

func (m *MockPassthroughProcessor) Process(_, _, _ string) error { return nil }

type mockFactory struct{}

func (f *mockFactory) GetProcessor(_ string) processor.FileProcessor {
	return &MockPassthroughProcessor{}
}

// BenchmarkWorkerPool_SchedulingOverhead measures pure worker scheduling
// overhead using a no-op processor (no file I/O).
func BenchmarkWorkerPool_SchedulingOverhead(b *testing.B) {
	const numJobs = 100

	opts := WorkerPoolOptions{
		Context:    context.Background(),
		InputDir:   "/mock/input",
		OutputDir:  "/mock/output",
		MarkerName: "tempo",
		NumWorkers: 4,
	}

	b.ResetTimer()

	for b.Loop() {
		manager := NewWorkerPoolManager(opts)
		// Replace factory with the no-op mock.
		manager.Factory = &mockFactory{}

		for i := range numJobs {
			manager.JobChan <- Job{
				InputPath:  "/mock/input/style" + string(rune('0'+i%10)) + ".css",
				OutputPath: "/mock/output/style" + string(rune('0'+i%10)) + ".templ",
			}
		}
		close(manager.JobChan)

		_ = manager.StartWorkers(opts.Context, opts.NumWorkers, false)
		close(manager.ErrorsChan)
		close(manager.SkippedChan)
	}
}
