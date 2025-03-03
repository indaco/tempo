package worker

import (
	"fmt"
	"strings"
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

// ProcessingError stores detailed error info.
type ProcessingError struct {
	FilePath string   `json:"file_path"`
	Message  string   `json:"message,omitempty"`
	Reason   string   `json:"reason,omitempty"`
	SkipType SkipType `json:"skip_type,omitempty"`
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
		errors = append(errors, err)
	}

	return errors
}

// FormatError formats errors for logging.
func FormatError(filePath string, err error) ProcessingError {
	return ProcessingError{
		FilePath: filePath,
		Message:  err.Error(),
	}
}

// FormatSkipReason creates a structured skip log entry.
func FormatSkipReason(filePath, reason string, skipType SkipType) ProcessingError {
	return ProcessingError{
		FilePath: filePath,
		Reason:   reason,
		SkipType: skipType,
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
			sb.WriteString(fmt.Sprintf("- Skipped File: %s\n  Reason: %s (Type: %s)\n", err.FilePath, err.Reason, err.SkipType))
		} else {
			sb.WriteString(fmt.Sprintf("- File: %s\n  Error: %s\n", err.FilePath, err.Message))
		}
	}

	fmt.Print(sb.String())
}
