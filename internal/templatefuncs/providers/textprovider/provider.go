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
//   - `titleCase`: Capitalizes the first letter of a word and preserves the rest of the word as-is.
//   - `isEmpty`: Converts a snake_case string to Title Case.
func (p *TextProvider) GetFunctions() template.FuncMap {
	return template.FuncMap{
		"normalizePath": NormalizePath,
		"isEmpty":       IsEmptyString,
		"titleCase":     TitleCase,
		"snakeToTitle":  SnakeToTitle,
	}
}

// Expose TextProvider as a global instance
var Provider templatefuncs.TemplateFuncProvider = &TextProvider{}
