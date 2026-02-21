package worker

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/indaco/tempo/internal/apperrors"
	"github.com/indaco/tempo/internal/processor"
	"golang.org/x/sync/errgroup"
)

/* ------------------------------------------------------------------------- */
/* Worker Pool Options & Manager                                             */
/* ------------------------------------------------------------------------- */

// WorkerPoolOptions holds configuration for the worker pool.
type WorkerPoolOptions struct {
	Context              context.Context
	InputDir             string
	OutputDir            string
	ExcludeDir           string
	MarkerName           string
	NumWorkers           int
	IsProduction         bool // If `--prod` is set, process everything
	IsForce              bool // If `--force` is set, process everything
	IsTrackExecutionTime bool
}

// WorkerPoolOption is a functional option for WorkerPoolOptions.
type WorkerPoolOption func(*WorkerPoolOptions)

// WithExcludeDir sets the directory to exclude from processing.
func WithExcludeDir(dir string) WorkerPoolOption {
	return func(o *WorkerPoolOptions) {
		o.ExcludeDir = dir
	}
}

// WithMarkerName sets the guard marker name used in .templ files.
func WithMarkerName(name string) WorkerPoolOption {
	return func(o *WorkerPoolOptions) {
		o.MarkerName = name
	}
}

// WithNumWorkers sets the number of concurrent workers.
func WithNumWorkers(n int) WorkerPoolOption {
	return func(o *WorkerPoolOptions) {
		o.NumWorkers = n
	}
}

// WithProduction enables production mode (minification).
func WithProduction(prod bool) WorkerPoolOption {
	return func(o *WorkerPoolOptions) {
		o.IsProduction = prod
	}
}

// WithForce enables force processing, ignoring modification timestamps.
func WithForce(force bool) WorkerPoolOption {
	return func(o *WorkerPoolOptions) {
		o.IsForce = force
	}
}

// WithTrackExecutionTime enables per-file execution time tracking.
func WithTrackExecutionTime(track bool) WorkerPoolOption {
	return func(o *WorkerPoolOptions) {
		o.IsTrackExecutionTime = track
	}
}

// NewWorkerPoolOptions constructs a WorkerPoolOptions with required fields and applies
// any functional options. NumWorkers defaults to runtime.NumCPU() * 2 when not set.
// Returns an error if NumWorkers is not positive after applying options.
func NewWorkerPoolOptions(ctx context.Context, inputDir, outputDir string, opts ...WorkerPoolOption) (WorkerPoolOptions, error) {
	o := WorkerPoolOptions{
		Context:    ctx,
		InputDir:   inputDir,
		OutputDir:  outputDir,
		NumWorkers: runtime.NumCPU() * 2,
	}

	for _, opt := range opts {
		opt(&o)
	}

	if o.NumWorkers <= 0 {
		return WorkerPoolOptions{}, apperrors.Wrap(fmt.Sprintf("NumWorkers must be greater than 0, got %d", o.NumWorkers))
	}

	return o, nil
}

// JobExecutionTime stores execution duration per file.
type JobExecutionTime struct {
	FilePath string
	Duration time.Duration
}

// WorkerPoolManager manages worker lifecycle.
type WorkerPoolManager struct {
	JobChan        chan Job
	ErrorsChan     chan ProcessingError
	SkippedChan    chan ProcessingError
	Metrics        *Metrics
	Factory        processor.ProcessorFactoryInterface
	InputDir       string
	OutputDir      string
	MarkerName     string
	ExecutionTimes []JobExecutionTime
	mu             sync.Mutex
}

// NewWorkerPoolManager initializes a worker pool manager.
// Paths are pre-cleaned to avoid repeated cleaning in hot loops.
func NewWorkerPoolManager(opts WorkerPoolOptions) *WorkerPoolManager {
	// Use larger buffer sizes to prevent data drops during high-throughput processing
	bufferSize := max(opts.NumWorkers*100, 100)

	// Pre-clean paths once to avoid repeated allocations in hot paths
	inputDir := filepath.Clean(opts.InputDir)
	outputDir := filepath.Clean(opts.OutputDir)

	return &WorkerPoolManager{
		JobChan:        make(chan Job, bufferSize),
		ErrorsChan:     make(chan ProcessingError, bufferSize),
		SkippedChan:    make(chan ProcessingError, bufferSize),
		Metrics:        NewMetrics(),
		Factory:        &processor.ProcessorFactory{Production: opts.IsProduction},
		InputDir:       inputDir,
		OutputDir:      outputDir,
		MarkerName:     opts.MarkerName,
		ExecutionTimes: make([]JobExecutionTime, 0, opts.NumWorkers*10),
	}
}

/* ------------------------------------------------------------------------- */
/* Worker Pool Execution                                                     */
/* ------------------------------------------------------------------------- */

// StartWorkers launches worker goroutines using `errgroup`.
func (m *WorkerPoolManager) StartWorkers(ctx context.Context, numWorkers int, trackExecution bool) error {
	if ctx == nil {
		return apperrors.Wrap("context is nil in StartWorkers")
	}

	if m.JobChan == nil || m.Factory == nil || m.Metrics == nil ||
		m.ErrorsChan == nil || m.SkippedChan == nil {
		return apperrors.Wrap("WorkerPoolManager is not properly initialized")
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(numWorkers)

	for range numWorkers {
		g.Go(func() error {
			return WorkerPool(ctx, m, trackExecution)
		})
	}

	// Ensure all workers complete before returning
	return g.Wait()
}
