package loader

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/indaco/tempo/internal/utils"
)

// setupMockProvider creates a mock Go module and writes provider.go.
func setupMockProvider(t *testing.T, tempDir, providerCode string) string {
	t.Helper()

	mockModulePath := filepath.Join(tempDir, "mockprovider")
	if err := os.MkdirAll(mockModulePath, 0755); err != nil {
		t.Fatalf("failed to create mock module directory: %v", err)
	}

	providerFile := filepath.Join(mockModulePath, "provider.go")
	if err := os.WriteFile(providerFile, []byte(providerCode), 0644); err != nil {
		t.Fatalf("failed to create provider.go: %v", err)
	}

	initializeGoModule(t, mockModulePath)
	return mockModulePath
}

// initializeGoModule runs 'go mod init' and 'go mod tidy'.
func initializeGoModule(t *testing.T, modulePath string) {
	t.Helper()

	cmd := exec.Command("go", "mod", "init", "mockprovider")
	cmd.Dir = modulePath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to initialize Go module: %v", err)
	}

	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = modulePath
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to run go mod tidy: %v", err)
	}
}

// TestValidateProviderPresence ensures that `validateProviderPresence` correctly detects the `Provider` variable.
func TestValidateProviderPresence(t *testing.T) {
	t.Run("ValidProvider", func(t *testing.T) {
		tempDir := t.TempDir()
		providerPath := filepath.Join(tempDir, "provider.go")

		// Valid provider.go with Provider variable
		providerContent := `package mockprovider

import "text/template"

// MockProvider struct
type MockProvider struct{}

// GetFunctions implementation
func (p *MockProvider) GetFunctions() template.FuncMap {
	return template.FuncMap{}
}

// Expose Provider as a global variable
var Provider = &MockProvider{}
`
		if err := os.WriteFile(providerPath, []byte(providerContent), 0644); err != nil {
			t.Fatalf("failed to create provider.go: %v", err)
		}

		// Should pass without errors
		if err := validateProviderPresence(providerPath); err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}
	})

	t.Run("MissingProviderVariable", func(t *testing.T) {
		tempDir := t.TempDir()
		providerPath := filepath.Join(tempDir, "provider.go")

		// No "Provider" variable
		invalidContent := `package mockprovider

import "text/template"

type MockProvider struct{}

func (p *MockProvider) GetFunctions() template.FuncMap {
	return template.FuncMap{}
}
`
		if err := os.WriteFile(providerPath, []byte(invalidContent), 0644); err != nil {
			t.Fatalf("failed to create provider.go: %v", err)
		}

		// Should return an error
		err := validateProviderPresence(providerPath)
		if err == nil || !utils.ContainsSubstring(err.Error(), "invalid function provider: Missing required exported Provider variable") {
			t.Errorf("Expected missing Provider error, got: %v", err)
		}
	})

	t.Run("MalformedProviderFile", func(t *testing.T) {
		tempDir := t.TempDir()
		providerPath := filepath.Join(tempDir, "provider.go")

		// Malformed Go file
		malformedContent := `package mockprovider

var Provider = `
		if err := os.WriteFile(providerPath, []byte(malformedContent), 0644); err != nil {
			t.Fatalf("failed to create provider.go: %v", err)
		}

		// Should return an error due to invalid Go syntax
		err := validateProviderPresence(providerPath)
		if err == nil {
			t.Errorf("Expected syntax error, but got nil")
		}
	})
}

// TestBuildProviderExecutable ensures that `buildProviderExecutable` correctly compiles a provider package.
func TestBuildProviderExecutable(t *testing.T) {
	tests := []struct {
		name           string
		providerCode   string
		expectedErrMsg string
	}{
		{
			name: "ValidProviderBuild",
			providerCode: `package mockprovider
import (
	"text/template"
	"github.com/indaco/tempo-api/templatefuncs"
)
type MockProvider struct{}
func (p *MockProvider) GetFunctions() template.FuncMap {
	return template.FuncMap{"hello": func() string { return "Hello" }}
}
var Provider templatefuncs.TemplateFuncProvider = &MockProvider{}
`,
		},
		{
			name: "MissingGetFunctionsMethod",
			providerCode: `package mockprovider
import "github.com/indaco/tempo-api/templatefuncs"
type MockProvider struct{}
var Provider templatefuncs.TemplateFuncProvider = &MockProvider{}
`,
			expectedErrMsg: "invalid function provider: Provider does not implement TemplateFuncProvider (missing method GetFunctions)",
		},
		{
			name: "InvalidGoSyntax",
			providerCode: `package mockprovider

var Provider = `,
			expectedErrMsg: "failed to compile provider package", // Matches actual error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			mockModulePath := setupMockProvider(t, tempDir, tt.providerCode)

			// Attempt to build the provider
			binPath, err := buildProviderExecutable(mockModulePath, "mockprovider")

			if tt.expectedErrMsg != "" {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got nil", tt.expectedErrMsg)
				} else if !utils.ContainsSubstring(err.Error(), tt.expectedErrMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.expectedErrMsg, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected successful build, but got error: %v", err)
				}

				// Ensure binary exists
				if _, statErr := os.Stat(binPath); os.IsNotExist(statErr) {
					t.Fatalf("Expected binary %s to exist, but it does not", binPath)
				}
			}
		})
	}
}

// TestExtractModuleName tests the extractModuleName function.
func TestExtractModuleName(t *testing.T) {
	t.Run("ValidGoMod", func(t *testing.T) {
		// Create a temporary directory
		tempDir := t.TempDir()

		// Create a valid go.mod file
		goModContent := "module github.com/example/project\n\ngo 1.23\n"
		goModPath := filepath.Join(tempDir, "go.mod")
		if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
			t.Fatalf("failed to create go.mod: %v", err)
		}

		// Run test
		moduleName, err := extractModuleName(tempDir)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		expected := "github.com/example/project"
		if moduleName != expected {
			t.Errorf("Expected %q, got %q", expected, moduleName)
		}
	})

	t.Run("GoModMissingModuleDeclaration", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create an empty go.mod file
		goModContent := "\ngo 1.18\n"
		goModPath := filepath.Join(tempDir, "go.mod")
		if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
			t.Fatalf("failed to create go.mod: %v", err)
		}

		// Run test
		_, err := extractModuleName(tempDir)
		if err == nil {
			t.Fatal("Expected error due to missing module declaration, but got nil")
		}
	})

	t.Run("GoModDoesNotExist", func(t *testing.T) {
		tempDir := t.TempDir()

		// Run test without creating go.mod
		_, err := extractModuleName(tempDir)
		if err == nil {
			t.Fatal("Expected error due to missing go.mod file, but got nil")
		}
	})

	t.Run("GoModWithInvalidFormat", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create a malformed go.mod file
		goModContent := "modu github.com/example/project\n"
		goModPath := filepath.Join(tempDir, "go.mod")
		if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
			t.Fatalf("failed to create go.mod: %v", err)
		}

		// Run test
		_, err := extractModuleName(tempDir)
		if err == nil {
			t.Fatal("Expected error due to incorrect module declaration format, but got nil")
		}
	})
}

// TestRunProviderExecutable ensures that `runProviderExecutable` correctly extracts registered functions.
func TestRunProviderExecutable(t *testing.T) {
	tests := []struct {
		name           string
		providerCode   string
		expectedErrMsg string
		expectedFunc   string
	}{
		{
			name: "ValidProviderExecution",
			providerCode: `package mockprovider
import (
	"text/template"
	"github.com/indaco/tempo-api/templatefuncs"
)
type MockProvider struct{}
func (p *MockProvider) GetFunctions() template.FuncMap {
	return template.FuncMap{"hello": func() string { return "Hello" }}
}
var Provider templatefuncs.TemplateFuncProvider = &MockProvider{}
`,
			expectedFunc: "hello",
		},
		{
			name: "ProviderIsNil",
			providerCode: `package mockprovider
import "github.com/indaco/tempo-api/templatefuncs"
var Provider templatefuncs.TemplateFuncProvider = nil
`,
			expectedErrMsg: "invalid function provider: Provider is nil",
		},
		{
			name: "NoFunctionsInProvider",
			providerCode: `package mockprovider
import (
	"text/template"
	"github.com/indaco/tempo-api/templatefuncs"
)
type MockProvider struct{}
func (p *MockProvider) GetFunctions() template.FuncMap {
	return template.FuncMap{}
}
var Provider templatefuncs.TemplateFuncProvider = &MockProvider{}
`,
			expectedErrMsg: "invalid function provider: No functions found in Provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			mockModulePath := setupMockProvider(t, tempDir, tt.providerCode)

			// Build provider binary
			binPath, err := buildProviderExecutable(mockModulePath, "mockprovider")
			if err != nil && tt.expectedErrMsg == "" {
				t.Fatalf("failed to build provider executable: %v", err)
			}

			// Run provider executable
			funcMap, err := runProviderExecutable(binPath, mockModulePath)

			if tt.expectedErrMsg != "" {
				if err == nil || err.Error() != tt.expectedErrMsg {
					t.Errorf("Expected error '%s', got: %v", tt.expectedErrMsg, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if _, exists := funcMap[tt.expectedFunc]; !exists {
					t.Errorf("Expected function '%s' to be registered, but it was not found.", tt.expectedFunc)
				}
			}
		})
	}
}
