package registry

import (
	"text/template"

	"maps"

	"github.com/indaco/tempo-api/templatefuncs"
)

// Global function registry for template functions.
var funcRegistry = template.FuncMap{}

// RegisterFuncProvider allows external function providers to register their functions.
func RegisterFuncProvider(provider templatefuncs.TemplateFuncProvider) {
	maps.Copy(funcRegistry, provider.GetFunctions())
}

// RegisterFunction registers an individual function.
func RegisterFunction(name string, fn any) {
	funcRegistry[name] = fn
}

// GetRegisteredFunctions returns all registered template functions.
func GetRegisteredFunctions() template.FuncMap {
	return funcRegistry
}
