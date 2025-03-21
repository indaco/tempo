// GoPackage utils provides utility functions and helpers for common operations
// such as string manipulation and template rendering. It integrates custom
// template functions with third-party libraries to enhance templating capabilities.
package utils

import (
	"bytes"
	"sync"
	"text/template"

	"github.com/indaco/tempo/internal/errors"
	"github.com/indaco/tempo/internal/templatefuncs/providers/gonameprovider"
	"github.com/indaco/tempo/internal/templatefuncs/providers/lookupprovider"
	"github.com/indaco/tempo/internal/templatefuncs/providers/textprovider"
	"github.com/indaco/tempo/internal/templatefuncs/registry"
)

var registerOnce sync.Once

// RenderTemplate renders a template string with the provided data.
//
// This function combines the `text/template` package with additional
// user registered functions to extend templating capabilities.
func RenderTemplate(templateContent string, data any) (string, error) {
	// Ensure all registered functions (default + user-defined) are available
	registerOnce.Do(func() {
		registry.RegisterFuncProvider(textprovider.Provider)
		registry.RegisterFuncProvider(gonameprovider.Provider)
		registry.RegisterFuncProvider(lookupprovider.Provider)
	})

	// Retrieve all registered functions, including user-defined ones
	funcMap := registry.GetRegisteredFunctions()

	tmpl, err := template.New("template").
		Funcs(funcMap).
		Option("missingkey=error").
		Parse(templateContent)
	if err != nil {
		return "", errors.Wrap("failed to run RenderTemplate", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
