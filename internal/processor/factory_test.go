package processor

import (
	"testing"

	"github.com/evanw/esbuild/pkg/api"
)

func TestProcessorFactory_GetProcessor(t *testing.T) {
	tests := []struct {
		filePath   string
		production bool
		expected   any // Expected processor type
	}{
		// Production  mode ON, should return MinifierProcessor
		{"script.js", true, &MinifierProcessor{}},
		{"styles.css", true, &MinifierProcessor{}},

		// Production mode OFF, should return PassthroughProcessor
		{"script.js", false, &PassthroughProcessor{}},
		{"styles.css", false, &PassthroughProcessor{}},

		// ❌ Unsupported file types should always return PassthroughProcessor
		{"index.html", true, &PassthroughProcessor{}},
		{"data.json", false, &PassthroughProcessor{}},
		{"readme.md", true, &PassthroughProcessor{}},
	}

	for _, test := range tests {
		factory := ProcessorFactory{Production: test.production}
		processor := factory.GetProcessor(test.filePath)

		switch test.expected.(type) {
		case *MinifierProcessor:
			if _, ok := processor.(*MinifierProcessor); !ok {
				t.Errorf("Expected MinifierProcessor for %s (minify: %t), got %T", test.filePath, test.production, processor)
			}
		case *PassthroughProcessor:
			if _, ok := processor.(*PassthroughProcessor); !ok {
				t.Errorf("Expected PassthroughProcessor for %s (minify: %t), got %T", test.filePath, test.production, processor)
			}
		default:
			t.Errorf("Unknown expected processor type for %s", test.filePath)
		}
	}
}

func TestNewEsbuildTransformer(t *testing.T) {
	tests := []struct {
		loader   api.Loader
		expected api.Loader
	}{
		{api.LoaderJS, api.LoaderJS},
		{api.LoaderCSS, api.LoaderCSS},
		{api.LoaderNone, api.LoaderNone}, // ❌ Invalid loader case
	}

	for _, test := range tests {
		transformer := newEsbuildTransformer(test.loader)
		if transformer.Loader != test.expected {
			t.Errorf("Expected loader %v, got %v", test.expected, transformer.Loader)
		}
	}
}
