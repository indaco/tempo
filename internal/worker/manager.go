package worker

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

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
		return fmt.Errorf("context is nil in StartWorkers")
	}

	if m.JobChan == nil || m.Factory == nil || m.Metrics == nil ||
		m.ErrorsChan == nil || m.SkippedChan == nil {
		return fmt.Errorf("WorkerPoolManager is not properly initialized")
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
