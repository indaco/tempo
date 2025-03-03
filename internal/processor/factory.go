package processor

import (
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/indaco/tempo/internal/processor/transformers"
)

// ProcessorFactory provides the correct FileProcessor based on file extension.
type ProcessorFactory struct {
	Production bool // Whether to use minification
}

// GetProcessor returns the appropriate FileProcessor.
func (f *ProcessorFactory) GetProcessor(filePath string) FileProcessor {
	ext := filepath.Ext(filePath)
	loader := GetLoader(ext)
	if f.Production && (ext == ".js" || ext == ".css") {
		if loader == api.LoaderNone {
			return &PassthroughProcessor{} // Fallback if loader is unknown
		}
		return &MinifierProcessor{Transform: newEsbuildTransformer(loader).Transform}
	}

	return &PassthroughProcessor{}
}

// newEsbuildTransformer initializes an EsbuildTransformer with the correct loader.
func newEsbuildTransformer(loader api.Loader) *transformers.EsbuildTransformer {
	return &transformers.EsbuildTransformer{Loader: loader}
}
