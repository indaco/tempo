package processor

import "github.com/evanw/esbuild/pkg/api"

// GetLoader determines the appropriate esbuild loader based on the file extension.
func GetLoader(ext string) api.Loader {
	switch ext {
	case ".js":
		return api.LoaderJS
	case ".css":
		return api.LoaderCSS
	default:
		return api.LoaderNone
	}
}
