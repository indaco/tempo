package synccmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/errors"
	"github.com/indaco/tempo/internal/helpers"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/resolver"
	"github.com/indaco/tempo/internal/utils"
	"github.com/indaco/tempo/internal/worker"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

/* ------------------------------------------------------------------------- */
/* Command Setup                                                             */
/* ------------------------------------------------------------------------- */

// SetupSyncCommand initializes the "process" CLI command.
func SetupSyncCommand(cmdCtx *app.AppContext) *cli.Command {
	return &cli.Command{
		Name:                   "sync",
		Usage:                  "Process & sync asset files into templ components",
		UsageText:              "tempo sync [options]",
		UseShortOptionHandling: true,
		Flags:                  getFlags(),
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			return ctx, app.IsTempoProject(cmdCtx.CWD)
		},
		Action: runSyncCommand(cmdCtx),
	}
}

/* ------------------------------------------------------------------------- */
/* Flag Generation                                                           */
/* ------------------------------------------------------------------------- */

// getFlags defines CLI flags.
func getFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "input",
			Aliases: []string{"i"},
			Usage:   "The directory containing asset files (e.g., CSS, JS) to be processed (default: assets)",
		},
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "The directory containing the .templ component files where assets will be injected (default: components)",
		},
		&cli.StringFlag{
			Name:    "exclude",
			Aliases: []string{"e"},
			Usage:   "Subfolder (relative to input directory) to exclude from the processing",
		},
		&cli.StringFlag{
			Name:    "workers",
			Aliases: []string{"w"},
			Usage:   "Number of concurrent workers",
		},
		&cli.BoolFlag{
			Name:    "prod",
			Aliases: []string{"p"},
			Usage:   "Enable production mode, minifying the injected content",
		},
		&cli.BoolFlag{
			Name:    "force",
			Aliases: []string{"f"},
			Usage:   "Force processing of all files, ignoring modification timestamps",
		},
		&cli.StringFlag{
			Name:    "summary",
			Aliases: []string{"s"},
			Usage:   "Summary format: compact, long, json, none (default: compact)",
		},
		&cli.BoolFlag{
			Name:  "verbose",
			Usage: "Show detailed information in the summary report",
		},
		&cli.BoolFlag{
			Name:  "track-time",
			Usage: "Display execution time per processed file.",
		},
		&cli.StringFlag{
			Name:    "report-file",
			Aliases: []string{"rf"},
			Usage:   "Export summary to a JSON file",
		},
	}
}

/* ------------------------------------------------------------------------- */
/* Command Runner                                                            */
/* ------------------------------------------------------------------------- */

func runSyncCommand(cmdCtx *app.AppContext) func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		helpers.EnableLoggerIndentation(cmdCtx.Logger)

		// Step 1: Get flag values
		opts, summaryOpts, err := resolveSyncFlags(ctx, cmd, cmdCtx)
		if err != nil {
			return err
		}

		// Step 2: Check prerequisites
		if err := validateSyncPrerequisites(opts.InputDir, opts.OutputDir); err != nil {
			return err
		}

		// Step 3: Run file processing
		cmdCtx.Logger.Info("Processing files...")
		if err := runWorkerPool(cmdCtx, opts, summaryOpts); err != nil {
			return errors.Wrap("failed processing files", err)
		} else {
			cmdCtx.Logger.Success("Processing completed successfully without errors.")
		}
		helpers.ResetLogger(cmdCtx.Logger)

		return nil
	}
}

/* ------------------------------------------------------------------------- */
/* Prerequisites Validation                                                  */
/* ------------------------------------------------------------------------- */

// validateSyncPrerequisites checks prerequisites for the "run" command, including:
// - Existence of the input folder
// - Existence of the output folder
func validateSyncPrerequisites(inputDir, outputDir string) error {
	foldersToCheck := map[string]string{
		"input_dir":  inputDir,
		"output_dir": outputDir,
	}

	missingFolders, err := utils.CheckMissingFolders(foldersToCheck)
	if err != nil {
		return err
	}

	if len(missingFolders) > 0 {
		return helpers.BuildMissingFoldersError(
			missingFolders,
			"",
			[]string{},
		)
	}

	return nil
}

/* ------------------------------------------------------------------------- */
/* Worker Pool Execution                                                     */
/* ------------------------------------------------------------------------- */

// runWorkerPool initializes and manages the worker pool.
func runWorkerPool(
	cmdCtx *app.AppContext,
	opts worker.WorkerPoolOptions,
	summaryOpts *worker.SummaryOptions,
) error {
	cacheFile := filepath.Join(cmdCtx.CWD, ".tempo-lastrun")
	lastRunTimestamp := getLastRunTimestamp(cacheFile)

	// Initialize worker pool manager
	manager := worker.NewWorkerPoolManager(opts)

	// Ensure all required fields are properly initialized
	if manager.JobChan == nil || manager.ErrorsChan == nil || manager.SkippedChan == nil || manager.Metrics == nil {
		return errors.Wrap("WorkerPoolManager initialization failed: missing required fields")
	}

	var (
		skippedFiles    []worker.ProcessingError
		collectedErrors []worker.ProcessingError
		skipMu          sync.Mutex
		errMu           sync.Mutex
		g               errgroup.Group
	)

	// Drain skipped files and store them
	g.Go(func() error {
		for skip := range manager.SkippedChan {
			skipMu.Lock()
			skippedFiles = append(skippedFiles, skip)
			skipMu.Unlock()

			manager.Metrics.IncrementSkippedFile()
		}
		return nil
	})

	// Drain errors and store them
	g.Go(func() error {
		for err := range manager.ErrorsChan {
			errMu.Lock()
			collectedErrors = append(collectedErrors, err)
			errMu.Unlock()
		}
		return nil
	})

	// Queue files for processing before closing job channel & starting workers
	if err := queueFilesForProcessing(opts, manager, lastRunTimestamp); err != nil {
		return errors.Wrap("Failed to queue files", err)
	}

	// Close job channel before starting workers
	close(manager.JobChan)

	// Start workers
	err := manager.StartWorkers(opts.Context, opts.NumWorkers, opts.IsTrackExecutionTime)
	if err != nil {
		return errors.Wrap("error starting the WorkerPoolManager", err)
	}

	// Close channels after all workers finish
	close(manager.ErrorsChan)
	close(manager.SkippedChan)

	// Wait for error and skipped file processing to complete
	if err := g.Wait(); err != nil {
		return errors.Wrap("Failed while collecting skipped/errors", err)
	}

	// Use the stored skipped files
	manager.Metrics.SkippedFiles = len(skippedFiles)

	if err := saveLastRunTimestamp(cacheFile); err != nil {
		return errors.Wrap("Failed to update last run timestamp", err)
	}

	// Handle Summary
	if err := handleSummary(cmdCtx.Logger, manager, collectedErrors, skippedFiles, summaryOpts); err != nil {
		return err
	}

	return nil
}

/* ------------------------------------------------------------------------- */
/* File Processing & Error Handling                                          */
/* ------------------------------------------------------------------------- */

// queueFilesForProcessing walks through the input directory and enqueues jobs.
func queueFilesForProcessing(
	opts worker.WorkerPoolOptions,
	manager *worker.WorkerPoolManager,
	lastRunTimestamp int64,
) error {
	return filepath.WalkDir(opts.InputDir, func(source string, d os.DirEntry, err error) error {
		if err != nil {
			handleError(manager, source, err)
			return nil
		}

		absPath, err := filepath.Abs(source)
		if err != nil {
			handleError(manager, source, err)
			return nil
		}

		if shouldExcludeDir(opts.ExcludeDir, absPath) || isExcludedFile(absPath) {
			handleSkip(manager.SkippedChan, worker.SkippedFile{
				Source:    source,
				Dest:      "", // No expected output file
				InputDir:  opts.InputDir,
				OutputDir: opts.OutputDir,
				Reason:    "Excluded by user or system file",
				SkipType:  worker.SkipExcluded,
			})
			return nil
		}

		outputFilePath := utils.RebasePathToOutput(source, opts.InputDir, opts.OutputDir)
		if !d.IsDir() && shouldProcessFile(source, outputFilePath, opts, lastRunTimestamp, manager) {
			if !enqueueJob(manager, source, outputFilePath) {
				handleSkip(manager.SkippedChan, worker.SkippedFile{
					Source:    source,
					Dest:      outputFilePath,
					InputDir:  opts.InputDir,
					OutputDir: opts.OutputDir,
					Reason:    "Job queue is full. Increase workers.",
					SkipType:  worker.SkipQueueFull,
				})
			}
		}
		return nil
	})
}

/* ------------------------------------------------------------------------- */
/* Helper Functions                                                          */
/* ------------------------------------------------------------------------- */

func resolveSyncFlags(
	ctx context.Context,
	cmd *cli.Command,
	cmdCtx *app.AppContext,
) (worker.WorkerPoolOptions, *worker.SummaryOptions, error) {
	inputDir, err := resolver.ResolveString(
		cmd.String("input"),
		cmdCtx.Config.App.AssetsDir,
		"input folder",
		config.DefaultAssetsDir,
		nil,
	)
	if err != nil {
		return worker.WorkerPoolOptions{}, nil, err
	}

	outputDir, err := resolver.ResolveString(
		cmd.String("output"),
		cmdCtx.Config.App.GoPackage,
		"output folder",
		config.DefaultGoPackage,
		nil,
	)
	if err != nil {
		return worker.WorkerPoolOptions{}, nil, err
	}

	excludeDir := cmd.String("exclude")
	isProd := cmd.Bool("prod")
	isForce := cmd.Bool("force")
	isTrackExecutionTime := cmd.Bool("track-time")

	numWorkers, err := resolver.ResolveInt(cmd.String("workers"), cmdCtx.Config.Processor.Workers, "workers")
	if err != nil {
		return worker.WorkerPoolOptions{}, nil, err
	}

	// Worker pool options
	opts := worker.WorkerPoolOptions{
		Context:              ctx,
		InputDir:             inputDir,
		OutputDir:            outputDir,
		ExcludeDir:           excludeDir,
		MarkerName:           cmdCtx.Config.Templates.GuardMarker,
		NumWorkers:           numWorkers,
		IsProduction:         isProd,
		IsForce:              isForce,
		IsTrackExecutionTime: isTrackExecutionTime,
	}

	// Summary options
	summaryFormat, err := resolver.ResolveString(
		cmd.String("summary"),
		cmdCtx.Config.Processor.SummaryFormat,
		"summary",
		config.DefaultSummaryFormat,
		[]string{"compact", "long", "json", "none"},
	)
	if err != nil {
		return worker.WorkerPoolOptions{}, nil, err
	}

	reportFile := cmd.String("report-file")
	isVerboseSummary := cmd.Bool("verbose")

	summaryOpts := &worker.SummaryOptions{
		Format:     worker.SummaryFormat(summaryFormat),
		ReportFile: reportFile,
		IsVerbose:  isVerboseSummary,
	}

	return opts, summaryOpts, nil
}

func handleSummary(
	logger logger.LoggerInterface,
	manager *worker.WorkerPoolManager,
	processingErrors []worker.ProcessingError,
	skippedFiles []worker.ProcessingError,
	summaryOpts *worker.SummaryOptions,
) error {
	// Skip summary if format is "none"
	if summaryOpts.Format == worker.FormatNone {
		return nil
	}

	// Generate summary
	summary, err := manager.Metrics.SummaryAsString(processingErrors, skippedFiles, summaryOpts)
	if err != nil {
		return errors.Wrap("Failed to generate summary", err)
	}

	// Print summary
	logger.Default(summary)

	// Handle Summary Export to JSON File
	if summaryOpts.ReportFile != "" {
		if err := manager.Metrics.ToJSONFile(processingErrors, skippedFiles, summaryOpts.ReportFile); err != nil {
			return errors.Wrap("Failed to export summary to JSON file", err)
		}
	}

	return nil
}

// handleError sends errors to the error channel without blocking.
func handleError(manager *worker.WorkerPoolManager, path string, err error) {
	select {
	case manager.ErrorsChan <- worker.FormatError(path, err):
	default: // Avoid blocking if the channel is closed
	}
}

// handleSkip sends skip reasons to the skipped channel.
func handleSkip(ch chan<- worker.ProcessingError, skipped worker.SkippedFile) {
	select {
	case ch <- worker.FormatSkipReason(skipped):
	default: // Prevent blocking
	}
}

// shouldExcludeDir checks if the current path should be excluded.
func shouldExcludeDir(excludeDir, absPath string) bool {
	return excludeDir != "" && strings.HasPrefix(absPath, excludeDir)
}

// shouldProcessFile decides whether the file should be processed or skipped.
func shouldProcessFile(source, dest string, opts worker.WorkerPoolOptions, lastRunTimestamp int64, manager *worker.WorkerPoolManager) bool {
	if isExcludedFile(source) {
		handleSkip(manager.SkippedChan, worker.SkippedFile{
			Source:    source,
			Dest:      dest,
			InputDir:  opts.InputDir,
			OutputDir: opts.OutputDir,
			Reason:    "Excluded system or hidden file",
			SkipType:  worker.SkipExcluded,
		})
		return false
	}

	lastModified, err := getFileLastModifiedTime(source)
	if err != nil {
		handleError(manager, source, err)
		return false
	}

	if !opts.IsProduction && !opts.IsForce && lastModified < lastRunTimestamp {
		handleSkip(manager.SkippedChan, worker.SkippedFile{
			Source:    source,
			Dest:      dest,
			InputDir:  opts.InputDir,
			OutputDir: opts.OutputDir,
			Reason:    "File unchanged since last run",
			SkipType:  worker.SkipUnchangedFile,
		})
		return false
	}
	return true
}

// enqueueJob attempts to enqueue a job and returns success status.
func enqueueJob(manager *worker.WorkerPoolManager, inputPath, outputPath string) bool {
	select {
	case manager.JobChan <- worker.Job{InputPath: inputPath, OutputPath: outputPath}:
		return true
	default:
		return false
	}
}

// isExcludedFile checks if a file should be ignored (e.g., system files like .DS_Store).
func isExcludedFile(path string) bool {
	base := filepath.Base(path)
	excludedFiles := map[string]bool{
		".DS_Store": true,
		"Thumbs.db": true,
	}

	return excludedFiles[base]
}
