package lookupprovider

import (
	"text/template"

	"github.com/indaco/tempo-api/templatefuncs"
)

// LookupProvider implements TemplateFuncProvider
type LookupProvider struct{}

// GetFunctions returns the built-in template functions.
// Supported Functions:
//   - `lookup`: Retrieves a value from a nested map using a dot-separated key.
func (p *LookupProvider) GetFunctions() template.FuncMap {
	return template.FuncMap{
		"lookup": Lookup,
	}
}

// Expose LookupProvider as a global instance
var Provider templatefuncs.TemplateFuncProvider = &LookupProvider{}
