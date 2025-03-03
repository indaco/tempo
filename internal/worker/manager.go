package worker

import (
	"context"
	"fmt"
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
	Factory        *processor.ProcessorFactory
	InputDir       string
	OutputDir      string
	MarkerName     string
	ExecutionTimes []JobExecutionTime
	mu             sync.Mutex
}

// NewWorkerPoolManager initializes a worker pool manager.
func NewWorkerPoolManager(opts WorkerPoolOptions) *WorkerPoolManager {
	return &WorkerPoolManager{
		JobChan:        make(chan Job, opts.NumWorkers*50),
		ErrorsChan:     make(chan ProcessingError, opts.NumWorkers),
		SkippedChan:    make(chan ProcessingError, opts.NumWorkers),
		Metrics:        NewMetrics(),
		Factory:        &processor.ProcessorFactory{Production: opts.IsProduction},
		InputDir:       opts.InputDir,
		OutputDir:      opts.OutputDir,
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
