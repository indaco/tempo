package worker

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/indaco/tempo/internal/utils"
)

// SkipType defines the type of skipped reason.
type SkipType string

const (
	SkipUnsupportedFile  SkipType = "unsupported_file"  // Not CSS/JS
	SkipMismatchedPath   SkipType = "mismatched_output" // Structure mismatch
	SkipMissingTemplFile SkipType = "missing_templ"     // Missing templ file matches
	SkipUnchangedFile    SkipType = "unchanged_file"    // File not changed
	SkipQueueFull        SkipType = "queue_full"        // job queue is full
	SkipExcluded         SkipType = "user_skipped"      // Excluded by user
)

// SkippedFile holds metadata about a skipped file.
type SkippedFile struct {
	Source    string   // Path to the source file
	Dest      string   // Expected output file path (if applicable)
	InputDir  string   // Root input directory
	OutputDir string   // Root output directory
	Reason    string   // Why the file was skipped
	SkipType  SkipType // Type of skip reason
}

// ProcessingError stores detailed error info.
type ProcessingError struct {
	Source   string   `json:"source"`         // Source file path
	Dest     string   `json:"dest,omitempty"` // Expected output file path (if applicable)
	Message  string   `json:"message,omitempty"`
	Reason   string   `json:"reason,omitempty"`    // Why it was skipped
	SkipType SkipType `json:"skip_type,omitempty"` // Type of skip reason
}

// CollectErrors collects and aggregates errors or skipped files.
func CollectErrors(errorsChan <-chan ProcessingError) []ProcessingError {
	var errors []ProcessingError

	// If the channel is nil, return an empty slice immediately
	if errorsChan == nil {
		return errors
	}

	// If the channel is closed, range will exit cleanly
	for err := range errorsChan {
		// Ensure the Source field is not empty before appending
		if err.Source != "" {
			errors = append(errors, err)
		}
	}

	return errors
}

// FormatError formats errors for logging.
func FormatError(filePath string, err error) ProcessingError {
	return ProcessingError{
		Source:  filePath,
		Message: err.Error(),
	}
}

// FormatSkipReason creates a structured skip log entry.
func FormatSkipReason(skipped SkippedFile) ProcessingError {
	relSource := skipped.Source
	relDest := skipped.Dest

	// Make source path relative to InputDir
	if rel, err := filepath.Rel(skipped.InputDir, skipped.Source); err == nil {
		relSource = filepath.Join(skipped.InputDir, rel) // Ensure base dir is included
	}

	// Handle the case when OutputDir is the current working directory (cwd)
	if skipped.OutputDir == utils.GetCWD() {
		// Make relDest relative to OutputDir (which is cwd in this case)
		if rel, err := filepath.Rel(skipped.OutputDir, skipped.Dest); err == nil {
			relDest = rel
		}
	} else {
		// Normal case: make it relative to OutputDir
		if rel, err := filepath.Rel(skipped.OutputDir, skipped.Dest); err == nil {
			relDest = filepath.Join(skipped.OutputDir, rel)
		}
	}

	return ProcessingError{
		Reason:   skipped.Reason,
		SkipType: skipped.SkipType,
		Source:   relSource,
		Dest:     relDest,
	}
}

// PrintErrors logs errors in a structured way.
func PrintErrors(errors []ProcessingError) {
	if len(errors) == 0 {
		return
	}

	var sb strings.Builder
	sb.WriteString("\nErrors encountered:\n")

	for _, err := range errors {
		if err.Reason != "" {
			sb.WriteString(fmt.Sprintf("- Skipped File: %s\n  Reason: %s (Type: %s)\n", err.Source, err.Reason, err.SkipType))
		} else {
			sb.WriteString(fmt.Sprintf("- File: %s\n  Error: %s\n", err.Source, err.Message))
		}
	}

	fmt.Print(sb.String())
}
