package processor

import (
	"fmt"
	"os"

	"github.com/indaco/tempo/internal/processor/transformers"
)

// StandardProcessor processes files without modifications.
type PassthroughProcessor struct{}

// Process simply inserts the raw content from the input file into the output file.
func (p *PassthroughProcessor) Process(inputFilePath, outputFilePath, markerName string) error {
	// Read input file content
	inputContent, err := os.ReadFile(inputFilePath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	transformerConfig := transformers.TransformationConfig{
		RawData:    string(inputContent),
		Transform:  func(input string) (string, error) { return input, nil },
		MarkerName: markerName,
	}

	return processWithTransformation(transformerConfig, outputFilePath)
}
