package processor

import (
	"testing"

	"github.com/evanw/esbuild/pkg/api"
)

// TestGetLoader ensures the correct loader is returned for each file extension.
func TestGetLoader(t *testing.T) {
	tests := []struct {
		ext      string
		expected api.Loader
	}{
		{".js", api.LoaderJS},
		{".css", api.LoaderCSS},
		{".txt", api.LoaderNone},
		{".html", api.LoaderNone},
		{"", api.LoaderNone},
	}

	for _, test := range tests {
		result := GetLoader(test.ext)
		if result != test.expected {
			t.Errorf("GetLoader(%q) = %v; want %v", test.ext, result, test.expected)
		}
	}
}
