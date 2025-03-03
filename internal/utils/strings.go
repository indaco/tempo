package utils

import (
	"path/filepath"
	"strings"
)

// ContainsSubstring checks if a substring exists within a string
func ContainsSubstring(str, substr string) bool {
	return strings.Contains(str, substr)
}

// ExtractNameFromURL extracts the repository name from a URL, removing `.git` if present.
func ExtractNameFromURL(url string) string {
	parts := strings.Split(strings.TrimSuffix(url, ".git"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

// ExtractNameFromPath extracts the last folder name from a given file path.
func ExtractNameFromPath(path string) string {
	if path == "" {
		return ""
	}
	return filepath.Base(filepath.Clean(path))
}
