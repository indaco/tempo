package processor

import (
	"os"

	apperrors "github.com/indaco/tempo/internal/apperrors"
	"github.com/indaco/tempo/internal/processor/transformers"
)

type MinifierProcessor struct {
	Transform func(string) (string, error) // Transformation function
}

func (p *MinifierProcessor) Process(inputFilePath, outputFilePath, markerName string) error {
	inputContent, err := os.ReadFile(inputFilePath)
	if err != nil {
		return apperrors.Wrap("failed to read input file", err)
	}

	transformerConfig := transformers.TransformationConfig{
		RawData:    string(inputContent),
		Transform:  p.Transform,
		MarkerName: markerName,
	}

	return processWithTransformation(transformerConfig, outputFilePath)
}
