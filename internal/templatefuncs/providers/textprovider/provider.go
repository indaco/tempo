package textprovider

import (
	"text/template"

	"github.com/indaco/tempo-api/templatefuncs"
)

// TextProvider implements tempo-api.TemplateFuncProvider
type TextProvider struct{}

// GetFunctions returns the built-in template functions.
// Supported Functions:
//   - `normalizePath`: normalizes a path string.
//   - `isEmpty`: Checks if a string is empty.
func (p *TextProvider) GetFunctions() template.FuncMap {
	return template.FuncMap{
		"normalizePath": NormalizePath,
		"isEmpty":       IsEmptyString,
		"titleCase":     TitleCase,
	}
}

// Expose DefaultProvider as a global instance
var Provider templatefuncs.TemplateFuncProvider = &TextProvider{}
