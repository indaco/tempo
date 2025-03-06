package worker

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

// Metrics tracks processing statistics.
type Metrics struct {
	FilesProcessed       int       `json:"files_processed"`
	DirectoriesProcessed int       `json:"directories_processed"`
	ErrorsEncountered    int       `json:"errors_encountered"`
	SkippedFiles         int       `json:"skipped_files"`
	StartTime            time.Time `json:"start_time"`
	ElapsedTime          string    `json:"elapsed_time"`
	mu                   sync.Mutex
}

// metricsExport is a struct for safely exporting metrics without copying the mutex.
type metricsExport struct {
	FilesProcessed       int       `json:"files_processed"`
	DirectoriesProcessed int       `json:"directories_processed"`
	ErrorsEncountered    int       `json:"errors_encountered"`
	SkippedFiles         int       `json:"skipped_files"`
	StartTime            time.Time `json:"start_time"`
	ElapsedTime          string    `json:"elapsed_time"`
}

// SummaryFormat defines available summary output formats.
type SummaryFormat string

// SummaryOptions holds configuration for summary output.
type SummaryOptions struct {
	Format     SummaryFormat // Output format: text, json, none
	ReportFile string        // File path to export JSON summary
	IsVerbose  bool
}

const (
	FormatNone    SummaryFormat = "none"
	FormatLong    SummaryFormat = "long"
	FormatCompact SummaryFormat = "compact"
	FormatJSON    SummaryFormat = "json"
)

// NewMetrics initializes the metrics struct.
func NewMetrics() *Metrics {
	return &Metrics{StartTime: time.Now()}
}

// Reset clears all metrics and resets the start time.
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.FilesProcessed = 0
	m.DirectoriesProcessed = 0
	m.ErrorsEncountered = 0
	m.SkippedFiles = 0
	m.ElapsedTime = ""
	m.StartTime = time.Now()
}

// IncrementFile updates the file counter.
func (m *Metrics) IncrementFile() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.FilesProcessed++
}

// IncrementDirectory updates the directory counter.
func (m *Metrics) IncrementDirectory() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DirectoriesProcessed++
}

// IncrementError updates the error counter.
func (m *Metrics) IncrementError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ErrorsEncountered++
}

// IncrementSkippedFile updates the skipped file counter.
func (m *Metrics) IncrementSkippedFile() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SkippedFiles++
}

// SummaryAsString generates and returns the processing summary in the requested format.
func (m *Metrics) SummaryAsString(errors []ProcessingError, skippedFiles []ProcessingError, summaryOpts *SummaryOptions) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update elapsed time before returning the summary
	m.ElapsedTime = time.Since(m.StartTime).String()

	// Ensure a valid format, defaulting to "long"
	if summaryOpts.Format != FormatJSON && summaryOpts.Format != FormatLong && summaryOpts.Format != FormatCompact {
		summaryOpts.Format = FormatLong
	}

	// Return JSON-formatted summary
	if summaryOpts.Format == FormatJSON {
		return m.summaryAsJSON(errors, skippedFiles)
	}

	// Return compact summary
	if summaryOpts.Format == FormatCompact {
		return m.summaryAsText(skippedFiles, summaryOpts.IsVerbose, true), nil
	}

	// Default to long format
	return m.summaryAsText(skippedFiles, summaryOpts.IsVerbose, false), nil
}

// PrintSummary prints the summary in text format.
func (m *Metrics) PrintSummary(errors []ProcessingError, skippedFiles []ProcessingError, verbose bool) {
	summaryOpts := &SummaryOptions{
		Format:    FormatLong,
		IsVerbose: verbose,
	}
	summary, _ := m.SummaryAsString(errors, skippedFiles, summaryOpts) // Ignore error since text mode can't fail
	fmt.Print(summary)
}

// ToJSONFile writes the summary to a JSON file.
func (m *Metrics) ToJSONFile(errors []ProcessingError, skippedFiles []ProcessingError, outputPath string) error {
	// Use SummaryAsString to get JSON output
	jsonData, err := m.summaryAsJSON(errors, skippedFiles)
	if err != nil {
		return err
	}

	// Write to JSON file
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(jsonData)
	return err
}

// summaryAsText returns the summary in human-readable text format.
func (m *Metrics) summaryAsText(skippedFiles []ProcessingError, verbose bool, compact bool) string {
	faint := color.New(color.Faint).SprintFunc()

	var sb strings.Builder
	sb.WriteString("\n📋 Processing Summary:\n")

	if compact {
		sb.WriteString(fmt.Sprintf("Files: %d | Dirs: %d | Skipped: %d | Errors: %d | Time: %s\n",
			m.FilesProcessed, m.DirectoriesProcessed, m.SkippedFiles, m.ErrorsEncountered, m.ElapsedTime))
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("  - Total files processed: %d\n", m.FilesProcessed))
	sb.WriteString(fmt.Sprintf("  - Total directories processed: %d\n", m.DirectoriesProcessed))
	sb.WriteString(fmt.Sprintf("  - Total errors encountered: %d\n", m.ErrorsEncountered))
	sb.WriteString(fmt.Sprintf("  - Total skipped files: %d\n", m.SkippedFiles))
	sb.WriteString(fmt.Sprintf("  - Elapsed time: %s\n", m.ElapsedTime))

	if !verbose {
		msg := "For more details, use the `--verbose-summary` flag."
		sb.WriteString("\n" + faint(msg) + "\n")
	}

	// Categorized Skipped File Breakdown
	if verbose && len(skippedFiles) > 0 {

		sb.WriteString("\n📌 Skipped Files Breakdown:\n")

		// Group skipped files by type
		categorized := map[SkipType][]string{
			SkipUnsupportedFile:  {},
			SkipMismatchedPath:   {},
			SkipMissingTemplFile: {},
			SkipUnchangedFile:    {},
			SkipQueueFull:        {},
		}

		for _, file := range skippedFiles {
			//reasonMsg := faint(fmt.Sprintf("(Reason: %s)", file.Reason))
			entry := fmt.Sprintf("    - file: %s", faint(file.FilePath))
			categorized[file.SkipType] = append(categorized[file.SkipType], entry)
		}

		// Define colors
		colorMap := map[SkipType]func(a ...any) string{
			SkipUnsupportedFile:  color.New(color.FgBlue, color.Bold).SprintFunc(),
			SkipMismatchedPath:   color.New(color.FgYellow, color.Bold).SprintFunc(),
			SkipMissingTemplFile: color.New(color.FgMagenta, color.Bold).SprintFunc(),
			SkipUnchangedFile:    color.New(color.FgCyan, color.Bold).SprintFunc(),
			SkipQueueFull:        color.New(color.FgRed, color.Bold).SprintFunc(),
		}

		// Output categorized skipped files
		formatSkippedCategory(&sb, "Unsupported File Types", categorized[SkipUnsupportedFile], colorMap[SkipUnsupportedFile],
			"Only CSS and JS files are supported. Ensure your files have valid extensions and are placed correctly.")

		formatSkippedCategory(&sb, "Mismatched Output Structure", categorized[SkipMismatchedPath], colorMap[SkipMismatchedPath],
			"The input folder structure must mirror the output structure. Check your file paths or adjust the configuration.")

		formatSkippedCategory(&sb, "Missing Templ File", categorized[SkipMissingTemplFile], colorMap[SkipMissingTemplFile],
			"Each CSS or JS file must have a corresponding .templ file. Verify that all expected .templ files exist.")

		formatSkippedCategory(&sb, "Unchanged Files", categorized[SkipUnchangedFile], colorMap[SkipUnchangedFile],
			"These files haven't changed since the last run. Use '--force' to process them anyway if needed.")

		formatSkippedCategory(&sb, "Queue Overflow (Increase Workers)", categorized[SkipQueueFull], colorMap[SkipQueueFull], "Consider increasing the number of workers (--workers) to prevent queue overflow.")
	}

	bold := color.New(color.Bold).SprintFunc()
	redIcon := color.New(color.FgRed, color.Bold).Sprint("✘")
	if m.ErrorsEncountered > 0 {
		sb.WriteString("\n" + redIcon + bold(" Some errors occurred. Check logs for details."))
	}

	return sb.String()
}

// summaryAsJSON returns the summary as a JSON string.
func (m *Metrics) summaryAsJSON(errors []ProcessingError, skippedFiles []ProcessingError) (string, error) {
	// Create a safe copy without the mutex
	exportData := metricsExport{
		FilesProcessed:       m.FilesProcessed,
		DirectoriesProcessed: m.DirectoriesProcessed,
		ErrorsEncountered:    m.ErrorsEncountered,
		SkippedFiles:         m.SkippedFiles,
		StartTime:            m.StartTime,
		ElapsedTime:          m.ElapsedTime,
	}

	data := struct {
		Metrics      metricsExport                `json:"metrics"`
		Errors       []ProcessingError            `json:"errors"`
		SkippedFiles map[string][]ProcessingError `json:"skipped_files"`
	}{
		Metrics: exportData,
		Errors:  errors,
		SkippedFiles: map[string][]ProcessingError{
			"unsupported_file":  filterSkippedFiles(skippedFiles, SkipUnsupportedFile),
			"mismatched_output": filterSkippedFiles(skippedFiles, SkipMismatchedPath),
			"missing_templ":     filterSkippedFiles(skippedFiles, SkipMissingTemplFile),
			"unchanged_file":    filterSkippedFiles(skippedFiles, SkipUnchangedFile),
			"queue_full":        filterSkippedFiles(skippedFiles, SkipQueueFull),
		},
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ") // Pretty-print JSON
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// Utility function to filter skipped files by type.
func filterSkippedFiles(files []ProcessingError, skipType SkipType) []ProcessingError {
	var filtered []ProcessingError
	for _, f := range files {
		if f.SkipType == skipType {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

// formatSkippedCategory prints a section for a given category
func formatSkippedCategory(
	sb *strings.Builder,
	title string,
	entries []string,
	colorFunc func(a ...any) string,
	hint string,
) {
	if len(entries) == 0 {
		return
	}

	// Define styling functions
	bold := color.New(color.Bold).SprintFunc()
	faint := color.New(color.Faint).SprintFunc()

	// Construct hint: "(Hint: some hint text)"
	formattedHint := faint(fmt.Sprintf("(%s %s)", bold("Hint:"), hint))

	// Print category title with icon
	sb.WriteString(fmt.Sprintf("\n  %s %s\n", colorFunc("•"), bold("Reason: "+title)))

	// Print hint below title
	sb.WriteString(fmt.Sprintf("    %s\n", formattedHint))

	// Print skipped file entries
	for _, entry := range entries {
		sb.WriteString(fmt.Sprintf("%s\n", entry))
	}
}
