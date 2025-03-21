package gonameprovider

import (
	"text/template"

	"github.com/indaco/tempo-api/templatefuncs"
)

// GoNameProvider implements TemplateFuncProvider
type GoNameProvider struct{}

// GetFunctions returns the built-in template functions.
// Supported Functions:
//   - `goPackageName`: Convert string to a valid Go package name.
//   - `goExportedName`:  Convert string to a valid **exported** Go function name.
//   - `goUnexportedName`:  Convert string to valid **unexported** Go function name.
func (p *GoNameProvider) GetFunctions() template.FuncMap {
	return template.FuncMap{
		"goPackageName":    ToGoPackageName,
		"goExportedName":   ToGoExportedName,
		"goUnexportedName": ToGoUnexportedName,
	}
}

// Expose GoNameProvider as a global instance
var Provider templatefuncs.TemplateFuncProvider = &GoNameProvider{}
