package synccmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/testhelpers"
	"github.com/indaco/tempo/internal/testutils"
	"github.com/indaco/tempo/internal/utils"
	"github.com/indaco/tempo/internal/worker"
	"github.com/urfave/cli/v3"
)

func TestSyncCommand(t *testing.T) {
	tempDir := os.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	outputDir := filepath.Join(tempDir, "output")

	// Create go.mod inside tempDir (the correct working directory)
	if err := testutils.CreateModFile(tempDir); err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	// Setup CLI config
	cfg := testutils.SetupConfig(tempDir, nil)

	// Write `tempo.yaml`
	configPath := filepath.Join(tempDir, "tempo.yaml")
	if err := testutils.WriteConfigToFile(configPath, cfg); err != nil {
		t.Fatalf("Failed to create mock config file: %v", err)
	}

	// Setup the CLI context
	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	// Ensure required folders exist before running "tempo sync"
	_ = os.MkdirAll(inputDir, 0755)
	_ = os.MkdirAll(outputDir, 0755)
	_ = os.MkdirAll(filepath.Join(cfg.Paths.TemplatesDir, "component"), 0755)
	_ = os.MkdirAll(filepath.Join(cfg.Paths.TemplatesDir, "component-variant"), 0755)
	_ = os.MkdirAll(cfg.Paths.ActionsDir, 0755)

	_ = os.MkdirAll(cfg.App.AssetsDir, 0755)
	_ = os.MkdirAll(cfg.App.GoPackage, 0755)

	tests := []struct {
		name           string
		args           []string
		expectedOutput []string
		expectedFiles  []string
	}{
		{
			name:           "Basic Execution",
			args:           []string{"tempo", "sync"},
			expectedOutput: []string{"Processing files..."},
			expectedFiles:  []string{"file1.templ", "file2.templ"},
		},
		{
			name:           "Run with Summary in JSON",
			args:           []string{"tempo", "sync", "--summary", "json"},
			expectedOutput: []string{"Processing files...", `"metrics":`, `"errors":`, `"skipped_files":`},
			expectedFiles:  []string{"file1.templ", "file2.templ"},
		},
		{
			name:           "Run with Forced Processing",
			args:           []string{"tempo", "sync", "--force"},
			expectedOutput: []string{"Processing files..."},
			expectedFiles:  []string{"file1.templ", "file2.templ"},
		},
		{
			name:           "Run with Excluded Directory",
			args:           []string{"tempo", "sync", "--exclude", "excluded"},
			expectedOutput: []string{"Processing files..."},
			expectedFiles:  []string{"file1.templ", "file2.templ"}, // Exclude file3.css
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock input files
			testFiles := []string{"file1.css", "file2.js"}
			for _, file := range testFiles {
				testutils.CreateFile(t, filepath.Join(inputDir, file), "mock content")
			}

			// Create an excluded subfolder with a file inside
			excludedDir := filepath.Join(inputDir, "excluded")
			_ = os.MkdirAll(excludedDir, 0755)
			testutils.CreateFile(t, filepath.Join(excludedDir, "file3.css"), "excluded content")

			// Ensure expected output files already exist before running `sync`
			for _, file := range tt.expectedFiles {
				outputFilePath := filepath.Join(outputDir, file)
				testutils.CreateFile(t, outputFilePath, "/* [tempo] BEGIN */\n/* [tempo] END */")
			}

			// Run the command
			app := &cli.Command{}
			app.Commands = []*cli.Command{
				SetupSyncCommand(cliCtx),
			}

			output, err := testhelpers.CaptureStdout(func() {
				err := app.Run(context.Background(), tt.args)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			})
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			// Validate CLI output
			testhelpers.ValidateCLIOutput(t, output, tt.expectedOutput)

			// Validate expected output files
			expectedOutputFiles := []string{}
			for _, file := range tt.expectedFiles {
				expectedOutputFiles = append(expectedOutputFiles, filepath.Join(outputDir, file))
			}
			testutils.ValidateGeneratedFiles(t, expectedOutputFiles)

			// Ensure excluded file was NOT processed
			excludedOutputFile := filepath.Join(outputDir, "excluded", "file3.templ")
			if _, err := os.Stat(excludedOutputFile); err == nil {
				t.Errorf("Excluded file %s was processed, but should have been skipped", excludedOutputFile)
			}
		})
	}
}

func TestSyncWorkerPool_BasicExecution(t *testing.T) {
	t.Log("[DEBUG] Starting TestRunWorkerPool_BasicExecution")

	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	outputDir := filepath.Join(tempDir, "output")

	cmdCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		CWD:    tempDir,
	}

	// Setup test files
	testFiles := []string{"file1.css", "file2.js"}
	for _, file := range testFiles {
		testutils.CreateFile(t, filepath.Join(inputDir, file), "body { color: black; }")
		testutils.CreateFile(t, filepath.Join(outputDir, file+".templ"), "/* [tempo] BEGIN */\n/* [tempo] END */")
	}

	// Increase timeout to avoid premature failures
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use WorkerPoolOptions
	opts := worker.WorkerPoolOptions{
		Context:      ctx,
		InputDir:     inputDir,
		OutputDir:    outputDir,
		MarkerName:   "tempo",
		NumWorkers:   2,
		IsProduction: false,
	}

	// Run worker pool
	err := runWorkerPool(cmdCtx, opts, &worker.SummaryOptions{})
	if err != nil {
		t.Fatalf("Worker pool execution failed: %v", err)
	}

	// Validate that output files were processed
	for _, file := range testFiles {
		outputFile := filepath.Join(outputDir, file+".templ")
		exists, _ := utils.FileExistsFunc(outputFile)
		if !exists {
			t.Errorf("Expected output file %s to exist", outputFile)
		}
	}

	t.Log("[DEBUG] Worker pool executed successfully")
}

func TestSyncWorkerPool_SummaryAsJSON(t *testing.T) {
	t.Log("[DEBUG] Starting TestRunWorkerPool_SummaryAsJSON")

	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	outputDir := filepath.Join(tempDir, "output")

	cmdCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		CWD:    tempDir,
	}

	// Setup test files
	testFiles := []string{"file1.css", "file2.js"}
	for _, file := range testFiles {
		testutils.CreateFile(t, filepath.Join(inputDir, file), "body { color: black; }")
		testutils.CreateFile(t, filepath.Join(outputDir, file+".templ"), "/* [tempo] BEGIN */\n/* [tempo] END */")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := worker.WorkerPoolOptions{
		Context:      ctx,
		InputDir:     inputDir,
		OutputDir:    outputDir,
		MarkerName:   "tempo",
		NumWorkers:   2,
		IsProduction: false,
	}

	// Capture JSON output
	output, err := testhelpers.CaptureStdout(func() {
		_ = runWorkerPool(cmdCtx, opts, &worker.SummaryOptions{Format: "json"})
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Validate JSON output
	var summaryData map[string]any
	if err := json.Unmarshal([]byte(output), &summaryData); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Check for required keys
	expectedKeys := []string{"metrics", "errors", "skipped_files"}
	for _, key := range expectedKeys {
		if _, exists := summaryData[key]; !exists {
			t.Errorf("Missing key in JSON summary: %s", key)
		}
	}

	t.Log("[DEBUG] JSON summary output validated successfully")
}

func TestSyncWorkerPool_SummaryToJSONFile(t *testing.T) {
	t.Log("[DEBUG] Starting TestRunWorkerPool_SummaryToJSONFile")

	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	outputDir := filepath.Join(tempDir, "output")
	summaryFile := filepath.Join(tempDir, "summary.json")

	cmdCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		CWD:    tempDir,
	}

	// Setup test files
	testFiles := []string{"file1.css", "file2.js"}
	for _, file := range testFiles {
		testutils.CreateFile(t, filepath.Join(inputDir, file), "body { color: black; }")
		testutils.CreateFile(t, filepath.Join(outputDir, file+".templ"), "/* [tempo] BEGIN */\n/* [tempo] END */")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := worker.WorkerPoolOptions{
		Context:      ctx,
		InputDir:     inputDir,
		OutputDir:    outputDir,
		MarkerName:   "tempo",
		NumWorkers:   2,
		IsProduction: false,
	}

	// Run worker pool with JSON file output
	err := runWorkerPool(cmdCtx, opts, &worker.SummaryOptions{Format: "json", ReportFile: summaryFile})
	if err != nil {
		t.Fatalf("Worker pool execution failed: %v", err)
	}

	// Validate JSON file exists
	if _, err := os.Stat(summaryFile); os.IsNotExist(err) {
		t.Fatalf("Expected summary file %s to exist", summaryFile)
	}

	// Validate JSON content
	jsonData, err := os.ReadFile(summaryFile)
	if err != nil {
		t.Fatalf("Failed to read summary file: %v", err)
	}

	var summaryData map[string]any
	if err := json.Unmarshal(jsonData, &summaryData); err != nil {
		t.Fatalf("Invalid JSON output in file: %v", err)
	}

	// Check for required keys
	expectedKeys := []string{"metrics", "errors", "skipped_files"}
	for _, key := range expectedKeys {
		if _, exists := summaryData[key]; !exists {
			t.Errorf("Missing key in JSON summary file: %s", key)
		}
	}

	t.Log("[DEBUG] JSON summary file validated successfully")
}

func TestWorkerErrorHandling(t *testing.T) {
	// Step 1: Create a channel to simulate errors
	errorsChan := make(chan worker.ProcessingError, 3)
	errorsChan <- worker.ProcessingError{FilePath: "/path/to/file1", Message: "Error 1"}
	errorsChan <- worker.ProcessingError{FilePath: "/path/to/file2", Message: "Error 2"}
	errorsChan <- worker.ProcessingError{FilePath: "/path/to/file3", Message: "Error 3"}
	close(errorsChan)

	// Capture output using testutils
	output, err := testhelpers.CaptureStdout(func() {
		collectedErrors := worker.CollectErrors(errorsChan)
		worker.PrintErrors(collectedErrors)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Expected output messages
	expectedMessages := []string{
		"Errors encountered:",
		"- File: /path/to/file1\n  Error: Error 1",
		"- File: /path/to/file2\n  Error: Error 2",
		"- File: /path/to/file3\n  Error: Error 3",
	}

	// Validate output
	testhelpers.ValidateCLIOutput(t, output, expectedMessages)
}

func TestQueueFilesForProcessing(t *testing.T) {
	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "input")
	outputDir := filepath.Join(tempDir, "output")
	cacheFile := filepath.Join(tempDir, ".tempo-lastrun")

	wpOpts := worker.WorkerPoolOptions{
		InputDir:     inputDir,
		OutputDir:    outputDir,
		IsProduction: false,
		IsForce:      false,
	}

	// Step 1: Setup test environment
	setupTestFiles(t, inputDir)
	lastRunTimestamp := recordLastRun(t, cacheFile)

	// Step 2: Test non-production mode (should skip unchanged files)
	t.Log("[DEBUG] Testing non-production mode")
	expectedSkippedFiles := []string{"file1.css", "file2.js", "file3.txt"}
	processedJobs, skippedFiles := verifyFileProcessing(t, wpOpts, lastRunTimestamp, 0, len(expectedSkippedFiles))

	// Step 3: Test production mode (should process all files)
	t.Log("[DEBUG] Testing production mode")
	wpOpts.IsProduction = true
	verifyFileProcessing(t, wpOpts, lastRunTimestamp, len(processedJobs)+len(skippedFiles), 0)
}

func TestValidateSyncPrerequisites(t *testing.T) {
	tempDir := t.TempDir() // Temporary directory for test isolation

	// Create a mock configuration
	cfg := config.DefaultConfig()
	cfg.Paths.TemplatesDir = filepath.Join(tempDir, "templates")
	cfg.App.AssetsDir = filepath.Join(tempDir, "assets")
	cfg.App.GoPackage = filepath.Join(tempDir, "package")

	cmdCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	t.Run("All Required Folders Exist", func(t *testing.T) {
		inputDir := cmdCtx.Config.App.AssetsDir
		outputDir := cmdCtx.Config.App.GoPackage
		// Create all required folders
		requiredDirs := []string{
			filepath.Join(cfg.Paths.TemplatesDir, "component"),
			filepath.Join(cfg.Paths.TemplatesDir, "component-variant"),
			inputDir,
			outputDir,
		}
		for _, dir := range requiredDirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("failed to create directory %s: %v", dir, err)
			}
		}

		err := validateSyncPrerequisites(inputDir, outputDir)
		if err != nil {
			t.Errorf("expected no error, but got: %v", err)
		}
	})

	t.Run("Some Required Folders Are Missing", func(t *testing.T) {
		// Remove one required folder
		missingDir := cmdCtx.Config.App.AssetsDir
		outputDir := cmdCtx.Config.App.GoPackage

		if err := os.RemoveAll(missingDir); err != nil {
			t.Fatalf("failed to remove directory %s: %v", missingDir, err)
		}

		err := validateSyncPrerequisites(missingDir, outputDir)
		if err == nil {
			t.Errorf("expected an error due to missing folders, but got nil")
		} else if !utils.ContainsSubstring(err.Error(), "Missing folders:") {
			t.Errorf("expected error message to mention missing folders, but got: %v", err)
		}
	})
}

func TestResolveSyncFlags(t *testing.T) {
	tempDir := t.TempDir()

	cfg := config.DefaultConfig()
	cfg.App.AssetsDir = filepath.Join(tempDir, "assets")
	cfg.App.GoPackage = filepath.Join(tempDir, "package")
	cfg.Templates.GuardMarker = "GUARD_MARKER"

	cmdCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    tempDir,
	}

	tests := []struct {
		name            string
		flags           map[string]any
		expectedOpts    worker.WorkerPoolOptions
		expectedSummary worker.SummaryOptions
		expectedForce   bool
		expectError     bool
	}{
		{
			name: "Valid Flags",
			flags: map[string]any{
				"prod":        true,
				"force":       true,
				"workers":     "8",
				"summary":     "json",
				"report-file": "report.json",
			},
			expectedOpts: worker.WorkerPoolOptions{
				InputDir:     filepath.Join(tempDir, "assets"),
				OutputDir:    filepath.Join(tempDir, "package"),
				MarkerName:   "GUARD_MARKER",
				NumWorkers:   8,
				IsProduction: true,
			},
			expectedSummary: worker.SummaryOptions{
				Format:     worker.SummaryFormat("json"),
				ReportFile: "report.json",
			},
			expectedForce: true,
			expectError:   false,
		},
		{
			name:  "Missing Flags, Defaults Applied",
			flags: map[string]any{},
			expectedOpts: worker.WorkerPoolOptions{
				InputDir:     filepath.Join(tempDir, "assets"),
				OutputDir:    filepath.Join(tempDir, "package"),
				MarkerName:   "GUARD_MARKER",
				NumWorkers:   runtime.NumCPU() * 2, // Default from config
				IsProduction: false,
			},
			expectedSummary: worker.SummaryOptions{
				Format:     worker.SummaryFormat("compact"), // Default from config
				ReportFile: "",
			},
			expectedForce: false,
			expectError:   false,
		},
		{
			name: "Invalid Workers Value",
			flags: map[string]any{
				"workers": "invalid",
			},
			expectedOpts:    worker.WorkerPoolOptions{},
			expectedSummary: worker.SummaryOptions{},
			expectedForce:   false,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a CLI app to parse flags properly
			app := &cli.Command{
				Flags: flagSetFromMap(tt.flags),
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cctx := context.Background()
					opts, summaryOpts, err := resolveSyncFlags(cctx, cmd, cmdCtx)

					if tt.expectError {
						if err == nil {
							t.Errorf("expected error but got nil")
						}
					} else {
						if err != nil {
							t.Errorf("unexpected error: %v", err)
						}

						// Compare options
						if opts.InputDir != tt.expectedOpts.InputDir ||
							opts.OutputDir != tt.expectedOpts.OutputDir ||
							opts.MarkerName != tt.expectedOpts.MarkerName ||
							opts.NumWorkers != tt.expectedOpts.NumWorkers ||
							opts.IsProduction != tt.expectedOpts.IsProduction {
							t.Errorf("expected opts %+v, got %+v", tt.expectedOpts, opts)
						}

						// Compare summary options
						if summaryOpts.Format != tt.expectedSummary.Format || summaryOpts.ReportFile != tt.expectedSummary.ReportFile {
							t.Errorf("expected summary %+v, got %+v", tt.expectedSummary, summaryOpts)
						}

						// Compare force flag
						if opts.IsForce != tt.expectedForce {
							t.Errorf("expected isForce %v, got %v", tt.expectedForce, opts.IsForce)
						}
					}
					return nil
				},
			}

			// Execute the CLI app to properly parse flags
			args := []string{"cmd"}
			for k, v := range tt.flags {
				args = append(args, "--"+k, formatFlagValue(v))
			}
			ctx := context.Background()
			err := app.Run(ctx, args)
			if err != nil {
				t.Fatalf("failed to run CLI app: %v", err)
			}
		})
	}
}

func TestQueueFilesForProcessing_NonDirectory(t *testing.T) {
	tempDir := t.TempDir()
	// Create a file instead of a directory to use as InputDir.
	filePath := filepath.Join(tempDir, "not_a_dir")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	opts := worker.WorkerPoolOptions{
		InputDir:     filePath, // This is a file, not a directory.
		OutputDir:    t.TempDir(),
		IsProduction: false,
	}
	manager := worker.NewWorkerPoolManager(opts)
	err := queueFilesForProcessing(opts, manager, 0)
	// Expect no error.
	if err != nil {
		t.Errorf("expected nil error when inputDir is not a directory, got: %v", err)
	}

	// Close the job channel to drain it.
	close(manager.JobChan)
	var jobs []worker.Job
	for job := range manager.JobChan {
		jobs = append(jobs, job)
	}
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs enqueued when inputDir is a file, got %d", len(jobs))
	}
}

func TestQueueFilesForProcessing_ExcludedFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create excluded files in the input directory
	excludedFiles := []string{".DS_Store", "Thumbs.db"}
	for _, fileName := range excludedFiles {
		filePath := filepath.Join(tempDir, fileName)
		if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create excluded file %q: %v", fileName, err)
		}
	}

	opts := worker.WorkerPoolOptions{
		InputDir:     tempDir,
		OutputDir:    t.TempDir(),
		IsProduction: false,
		NumWorkers:   2, // Ensure a reasonable buffer size
	}
	manager := worker.NewWorkerPoolManager(opts)

	// Ensure SkippedChan is initialized before testing
	if manager.SkippedChan == nil {
		t.Fatal("SkippedChan is nil, check worker initialization")
	}

	err := queueFilesForProcessing(opts, manager, 0)
	if err != nil {
		t.Errorf("expected nil error when processing inputDir, got: %v", err)
	}

	// Close SkippedChan and drain it to ensure all values are read
	close(manager.SkippedChan)

	// Collect skipped files
	skippedFiles := make(map[string]bool)
	for skip := range manager.SkippedChan {
		skippedFiles[filepath.Base(skip.FilePath)] = true
	}

	// Verify all excluded files were skipped
	for _, fileName := range excludedFiles {
		if !skippedFiles[fileName] {
			t.Errorf("expected %q to be skipped, but it was not", fileName)
		}
	}

	// Verify that the total skipped count matches expected
	if len(skippedFiles) != len(excludedFiles) {
		t.Errorf("expected %d skipped files, got %d", len(excludedFiles), len(skippedFiles))
	}
}

func TestShouldProcessFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	excludedFile := filepath.Join(tempDir, ".DS_Store")
	oldFile := filepath.Join(tempDir, "old.js")
	newFile := filepath.Join(tempDir, "new.js")

	// Write dummy content to files
	if err := os.WriteFile(excludedFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create excluded file: %v", err)
	}
	if err := os.WriteFile(oldFile, []byte("old"), 0644); err != nil {
		t.Fatalf("failed to create old file: %v", err)
	}
	if err := os.WriteFile(newFile, []byte("new"), 0644); err != nil {
		t.Fatalf("failed to create new file: %v", err)
	}

	// Set old file's mod time to an old timestamp
	oldTimestamp := time.Now().Add(-24 * time.Hour).Unix()
	if err := os.Chtimes(oldFile, time.Unix(oldTimestamp, 0), time.Unix(oldTimestamp, 0)); err != nil {
		t.Fatalf("failed to modify old file timestamp: %v", err)
	}

	// Set new file's mod time to a recent timestamp
	newTimestamp := time.Now().Unix()
	if err := os.Chtimes(newFile, time.Unix(newTimestamp, 0), time.Unix(newTimestamp, 0)); err != nil {
		t.Fatalf("failed to modify new file timestamp: %v", err)
	}

	// Define test cases
	tests := []struct {
		name           string
		filePath       string
		opts           worker.WorkerPoolOptions
		lastRun        int64
		expectedResult bool
		expectedSkip   bool
	}{
		{
			name:           "Excluded file",
			filePath:       excludedFile,
			opts:           worker.WorkerPoolOptions{IsProduction: false, IsForce: false, NumWorkers: 1},
			lastRun:        newTimestamp,
			expectedResult: false,
			expectedSkip:   true,
		},
		{
			name:           "Old file not in force mode",
			filePath:       oldFile,
			opts:           worker.WorkerPoolOptions{IsProduction: false, IsForce: false, NumWorkers: 1},
			lastRun:        newTimestamp,
			expectedResult: false,
			expectedSkip:   true,
		},
		{
			name:           "Old file in force mode",
			filePath:       oldFile,
			opts:           worker.WorkerPoolOptions{IsProduction: false, IsForce: true, NumWorkers: 1},
			lastRun:        newTimestamp,
			expectedResult: true,
			expectedSkip:   false,
		},
		{
			name:           "New file should be processed",
			filePath:       newFile,
			opts:           worker.WorkerPoolOptions{IsProduction: false, IsForce: false, NumWorkers: 1},
			lastRun:        oldTimestamp,
			expectedResult: true,
			expectedSkip:   false,
		},
		{
			name:           "Production mode ignores timestamp",
			filePath:       oldFile,
			opts:           worker.WorkerPoolOptions{IsProduction: true, IsForce: false, NumWorkers: 1},
			lastRun:        newTimestamp,
			expectedResult: true,
			expectedSkip:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := worker.NewWorkerPoolManager(tt.opts)

			// Run the function
			result := shouldProcessFile(tt.filePath, tt.opts, tt.lastRun, manager)

			// Validate the result
			if result != tt.expectedResult {
				t.Errorf("shouldProcessFile(%q) = %v; want %v", tt.filePath, result, tt.expectedResult)
			}

			// Close SkippedChan and drain it to validate skipped files
			close(manager.SkippedChan)
			skippedFiles := make(map[string]bool)
			for skip := range manager.SkippedChan {
				skippedFiles[filepath.Base(skip.FilePath)] = true
			}

			if tt.expectedSkip {
				if !skippedFiles[filepath.Base(tt.filePath)] {
					t.Errorf("expected %q to be skipped, but it was not", tt.filePath)
				}
			} else {
				if skippedFiles[filepath.Base(tt.filePath)] {
					t.Errorf("did not expect %q to be skipped, but it was", tt.filePath)
				}
			}
		})
	}
}

func TestEnqueueJob(t *testing.T) {
	manager := &worker.WorkerPoolManager{
		JobChan: make(chan worker.Job, 1),
	}
	// Should succeed initially.
	ok := enqueueJob(manager, "input", "output")
	if !ok {
		t.Errorf("expected enqueueJob to succeed, but it failed")
	}
	// Now the channel is full, so enqueueJob should return false.
	ok = enqueueJob(manager, "input", "output")
	if ok {
		t.Errorf("expected enqueueJob to fail when channel is full, but it succeeded")
	}
}

func TestShouldExcludeDir(t *testing.T) {
	excludeDir := "/tmp/exclude"
	absPath := "/tmp/exclude/subdir"
	if !shouldExcludeDir(excludeDir, absPath) {
		t.Errorf("expected path %q to be excluded", absPath)
	}
	absPath = "/tmp/include/subdir"
	if shouldExcludeDir(excludeDir, absPath) {
		t.Errorf("expected path %q not to be excluded", absPath)
	}
}

func TestIsExcludedFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{"Exclude .DS_Store", "/some/path/.DS_Store", true},
		{"Exclude Thumbs.db", "/some/path/Thumbs.db", true},
		{"Normal file", "/some/path/style.css", false},
		{"Hidden file", "/some/path/.hiddenfile", false},
		{"Nested .DS_Store", "/some/path/assets/.DS_Store", true},
		{"Nested Thumbs.db", "/some/path/assets/Thumbs.db", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isExcludedFile(tt.filePath)
			if result != tt.expected {
				t.Errorf("isExcludedFile(%q) = %v; want %v", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	manager := worker.NewWorkerPoolManager(worker.WorkerPoolOptions{
		InputDir:     "dummy",
		OutputDir:    "dummy",
		IsProduction: false,
	})
	manager.ErrorsChan = make(chan worker.ProcessingError, 1)
	fakeErr := errors.New("test error")
	handleError(manager, "some/path", fakeErr)
	close(manager.ErrorsChan)

	count := 0
	for pe := range manager.ErrorsChan {
		count++
		if !utils.ContainsSubstring(pe.Message, "test error") {
			t.Errorf("expected error message to contain 'test error', got: %s", pe.Message)
		}
	}
	if count == 0 {
		t.Errorf("expected at least one error from handleError, got 0")
	}
}

/* ------------------------------------------------------------------------- */
/* Testing Helpers                                                           */
/* ------------------------------------------------------------------------- */

// setupTestFiles creates mock input files
func setupTestFiles(t *testing.T, inputDir string) {
	t.Helper()
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatalf("Failed to create input folder: %v", err)
	}

	files := []string{"file1.css", "file2.js", "file3.txt"}
	for _, file := range files {
		filePath := filepath.Join(inputDir, file)
		if err := os.WriteFile(filePath, []byte("mock content"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	// Ensure timestamps are different
	time.Sleep(2 * time.Second)
}

// recordLastRun saves the last run timestamp
func recordLastRun(t *testing.T, cacheFile string) int64 {
	t.Helper()
	lastRunTimestamp := time.Now().Unix()
	if err := saveLastRunTimestamp(cacheFile); err != nil {
		t.Fatalf("Failed to save last run timestamp: %v", err)
	}
	return lastRunTimestamp
}

// verifyFileProcessing runs queueFilesForProcessing and validates expected jobs & skipped files
func verifyFileProcessing(
	t *testing.T,
	wpOpts worker.WorkerPoolOptions,
	lastRunTimestamp int64,
	expectedJobs int,
	expectedSkipped int,
) ([]worker.Job, []worker.ProcessingError) {
	t.Helper()

	// Setup channels
	jobChan := make(chan worker.Job, 10)
	errorsChan := make(chan worker.ProcessingError, 10)
	skippedChan := make(chan worker.ProcessingError, 10)

	manager := &worker.WorkerPoolManager{
		JobChan:     jobChan,
		ErrorsChan:  errorsChan,
		SkippedChan: skippedChan,
	}

	// Run function under test
	err := queueFilesForProcessing(wpOpts, manager, lastRunTimestamp)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Collect results
	close(jobChan)
	close(skippedChan)

	var processedJobs []worker.Job
	var skippedFiles []worker.ProcessingError

	for job := range jobChan {
		processedJobs = append(processedJobs, job)
	}
	for skip := range skippedChan {
		skippedFiles = append(skippedFiles, skip)
	}

	// Validate counts
	if len(processedJobs) != expectedJobs {
		t.Errorf("Expected %d jobs, but got %d", expectedJobs, len(processedJobs))
	}
	if len(skippedFiles) != expectedSkipped {
		t.Errorf("Expected %d skipped files, but got %d", expectedSkipped, len(skippedFiles))
	}

	return processedJobs, skippedFiles
}

// Helper function to convert a map into cli.Flags
func flagSetFromMap(flags map[string]any) []cli.Flag {
	var flagSet []cli.Flag
	for k, v := range flags {
		switch v := v.(type) {
		case bool:
			flagSet = append(flagSet, &cli.BoolFlag{Name: k, Value: v})
		case int:
			flagSet = append(flagSet, &cli.IntFlag{Name: k, Value: int64(v)})
		case string:
			flagSet = append(flagSet, &cli.StringFlag{Name: k, Value: v})
		}
	}
	return flagSet
}

// Helper function to convert a flag value to a string
func formatFlagValue(value any) string {
	switch v := value.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int:
		return fmt.Sprintf("%d", v)
	case string:
		return v
	default:
		return ""
	}
}
