package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/indaco/tempo/internal/utils"
)

// TestFindProviderFile verifies that findProviderFile correctly finds provider.go and extracts the package name.
func TestFindProviderFile(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(tempDir string) (expectedPath, expectedPackage string)
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name: "ProviderInRoot",
			setup: func(tempDir string) (string, string) {
				providerPath := filepath.Join(tempDir, "provider.go")
				packageName := "mockprovider"
				writeTestFile(t, providerPath, "package "+packageName)
				return providerPath, packageName
			},
		},
		{
			name: "ProviderInSubdirectory",
			setup: func(tempDir string) (string, string) {
				subDir := filepath.Join(tempDir, "provider")
				_ = os.MkdirAll(subDir, 0755)
				providerPath := filepath.Join(subDir, "provider.go")
				packageName := "provider"
				writeTestFile(t, providerPath, "package "+packageName)
				return providerPath, packageName
			},
		},
		{
			name: "MultipleProviders_ReturnFirstMatch",
			setup: func(tempDir string) (string, string) {
				provider1 := filepath.Join(tempDir, "provider.go")
				provider2Dir := filepath.Join(tempDir, "nested")
				_ = os.MkdirAll(provider2Dir, 0755)
				provider2 := filepath.Join(provider2Dir, "provider.go")
				writeTestFile(t, provider1, "package mockprovider")
				writeTestFile(t, provider2, "package nestedprovider")
				return provider1, "mockprovider"
			},
		},
		{
			name:        "NoProviderFile",
			setup:       func(tempDir string) (string, string) { return "", "" },
			expectError: true,
		},
		{
			name:        "EmptyDirectory",
			setup:       func(tempDir string) (string, string) { return "", "" },
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			expectedPath, expectedPackage := tt.setup(tempDir)

			meta, err := findProviderFile(tempDir)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error, but got nil")
				}
				if tt.expectedErrorMsg != "" && !utils.ContainsSubstring(err.Error(), tt.expectedErrorMsg) {
					t.Fatalf("Expected error containing '%s', got '%s'", tt.expectedErrorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if meta.FilePath != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, meta.FilePath)
			}
			if meta.Package != expectedPackage {
				t.Errorf("Expected package '%s', got '%s'", expectedPackage, meta.Package)
			}
			if meta.ModuleDir != tempDir {
				t.Errorf("Expected module directory '%s', got '%s'", tempDir, meta.ModuleDir)
			}
		})
	}
}

// writeTestFile is a helper to create test files.
func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
}

// TestExtractPackageName verifies that extractPackageName correctly identifies package names.
func TestExtractPackageName(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedPackage string
		expectError     bool
	}{
		{"ValidPackageName", "package mockprovider", "mockprovider", false},
		{"PackageNameWithLeadingSpaces", "   package examplepkg", "examplepkg", false},
		{"PackageNameWithComments", "// Comment\n// Another\npackage providerpkg", "providerpkg", false},
		{"InvalidPackageDeclaration", "package", "", true},
		{"MissingPackageDeclaration", "// Just a comment", "", true},
		{"EmptyFile", "", "", true},
		{"NonExistentFile", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var providerFile string
			tempDir := t.TempDir()

			if tt.name != "NonExistentFile" {
				providerFile = filepath.Join(tempDir, "provider.go")
				writeTestFile(t, providerFile, tt.content)
			} else {
				providerFile = filepath.Join(tempDir, "nonexistent.go")
			}

			packageName, err := extractPackageName(providerFile)

			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error, but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if packageName != tt.expectedPackage {
				t.Errorf("Expected package name '%s', got '%s'", tt.expectedPackage, packageName)
			}
		})
	}
}
