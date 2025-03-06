package transformers

import (
	"fmt"

	"github.com/evanw/esbuild/pkg/api"
)

type EsbuildTransformer struct {
	Loader api.Loader
}

func (e *EsbuildTransformer) Transform(input string) (string, error) {
	result := api.Transform(input, api.TransformOptions{
		Loader:            e.Loader,
		MinifySyntax:      true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: false,
	})

	if len(result.Errors) > 0 {
		return "", fmt.Errorf("esbuild minification error: %v", result.Errors)
	}

	return string(result.Code), nil
}
