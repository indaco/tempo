package processor

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/indaco/tempo/internal/processor/transformers"
	"github.com/indaco/tempo/internal/utils"
)

// processWithTransformation applies a transformation function to the input content
// and inserts the transformed content between configurable guard markers in the output file.
func processWithTransformation(cfg transformers.TransformationConfig, outputFilePath string) error {

	// Step 1: Read the output file content
	outputContent, err := os.ReadFile(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to read output file: %w", err)
	}

	// Step 2: Validate Guard Markers
	startMarker := fmt.Sprintf("/* [%s] BEGIN - Do not edit! This section is auto-generated. */", cfg.MarkerName)
	endMarker := fmt.Sprintf("/* [%s] END */", cfg.MarkerName)

	startIndex := bytes.Index(outputContent, []byte(startMarker))
	endIndex := bytes.Index(outputContent, []byte(endMarker))

	if err := validateGuardMarkers(startIndex, endIndex, outputFilePath); err != nil {
		return err
	}
	if startIndex == -1 && endIndex == -1 {
		return nil // No processing required if markers are absent
	}

	// Step 3: Apply transformation
	transformedContent, err := cfg.Transform(cfg.RawData)
	if err != nil {
		return fmt.Errorf("failed to transform content: %w", err)
	}

	// Step 4: Construct new content (removing any old content between markers)
	beforeMarker := strings.TrimRight(string(outputContent[:startIndex+len(startMarker)]), " \n") + "\n"
	afterMarker := strings.TrimLeft(string(outputContent[endIndex:]), " \n")

	var updatedContent strings.Builder
	updatedContent.WriteString(beforeMarker)
	updatedContent.WriteString(transformedContent + "\n")
	updatedContent.WriteString(afterMarker)

	// Step 5: Write the updated content back to the output file
	if err := utils.WriteStringToFile(outputFilePath, updatedContent.String()); err != nil {
		return fmt.Errorf("failed to write updated content to output file: %w", err)
	}

	return nil
}

// validateGuardMarkers ensures the markers exist and are properly ordered
func validateGuardMarkers(startIndex, endIndex int, outputFilePath string) error {
	switch {
	case startIndex == -1 && endIndex == -1:
		return nil // No markers, no error
	case startIndex == -1 || endIndex == -1 || startIndex > endIndex:
		return fmt.Errorf("invalid or missing guard markers in %s", outputFilePath)
	}
	return nil
}
