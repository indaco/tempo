package testutils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper to check if content exists between guard markers
func ContainsBetweenMarkers(content, expected string) bool {
	startMarker := "/* [tempo] BEGIN - Do not edit! This section is auto-generated. */"
	endMarker := "/* [tempo] END */"

	startIndex := strings.Index(content, startMarker)
	endIndex := strings.Index(content, endMarker)

	if startIndex == -1 || endIndex == -1 || startIndex > endIndex {
		return false
	}

	actualBetweenMarkers := content[startIndex+len(startMarker) : endIndex]
	return strings.Contains(actualBetweenMarkers, expected)
}

// Helper function to check if a string contains another string
func Contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

// Helper to create files
func CreateFile(t *testing.T, path, content string) {
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	err = os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
}

// Helper to generate templ files with guard markers
func GenerateTemplContent(packageName string) string {
	return "package " + packageName + "\n\nfunc Component() {\n/* [tempo] BEGIN - Do not edit! This section is auto-generated. */\n/* [tempo] END */\n}"
}

// Helper to verify if content was inserted correctly
func VerifyFileContent(t *testing.T, filePath, expectedContent string) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if !ContainsBetweenMarkers(string(data), expectedContent) {
		t.Errorf("Expected content missing in %s\nExpected: %q\nGot:\n%s", filePath, expectedContent, string(data))
	}
}
