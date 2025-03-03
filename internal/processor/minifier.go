package processor

import (
	"fmt"
	"os"

	"github.com/indaco/tempo/internal/processor/transformers"
)

type MinifierProcessor struct {
	Transform func(string) (string, error) // Transformation function
}

func (p *MinifierProcessor) Process(inputFilePath, outputFilePath, markerName string) error {
	inputContent, err := os.ReadFile(inputFilePath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	transformerConfig := transformers.TransformationConfig{
		RawData:    string(inputContent),
		Transform:  p.Transform,
		MarkerName: markerName,
	}

	return processWithTransformation(transformerConfig, outputFilePath)
}
