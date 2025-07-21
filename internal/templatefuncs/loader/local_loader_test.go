package loader

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/templatefuncs/registry"
	"github.com/indaco/tempo/internal/utils"
)

// TestRegisterFunctionsFromPath ensures functions from a local package are loaded correctly.
func TestRegisterFunctionsFromPath(t *testing.T) {

	// Create a temporary directory to simulate a standalone Go module.
	tempDir := t.TempDir()

	// Simulated provider implementation inside `mockprovider`
	mockCode := `package mockprovider

import (
	"text/template"
	"github.com/indaco/tempo-api/templatefuncs"
)

// MockProvider implements TemplateFuncProvider.
type MockProvider struct{}

// GetFunctions returns the function map.
func (p *MockProvider) GetFunctions() template.FuncMap {
	return template.FuncMap{
		"localFunc": func() string { return "Hello from Local!" },
	}
}

// Expose Provider as a global variable.
var Provider templatefuncs.TemplateFuncProvider = &MockProvider{}
`

	// Step 1: Create `mockprovider` directory and `provider.go`
	mockPackagePath := filepath.Join(tempDir, "mockprovider")
	_ = os.MkdirAll(mockPackagePath, 0755)
	providerFile := filepath.Join(mockPackagePath, "provider.go")
	if err := os.WriteFile(providerFile, []byte(mockCode), 0644); err != nil {
		t.Fatalf("Failed to create mock provider file: %v", err)
	}

	// Step 2: Initialize `go.mod` inside `mockprovider`
	cmd := exec.Command("go", "mod", "init", "mockprovider")
	cmd.Dir = mockPackagePath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize go module inside mockprovider: %v", err)
	}

	// Step 4: Run `go mod tidy` to ensure dependencies are resolved
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = mockPackagePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run `go mod tidy` inside mockprovider: %v", err)
	}

	// Step 5: Ensure the package is valid using `go build`
	cmd = exec.Command("go", "build", "./...")
	cmd.Dir = mockPackagePath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Go package is invalid: %v", err)
	}

	// Step 6: Ensure `mockprovider` is detected as a valid package
	cmd = exec.Command("go", "list", "./...")
	cmd.Dir = mockPackagePath
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to inspect Go package: %v", err)
	}

	// Step 7: Attempt to register functions
	err := RegisterFunctionsFromPath(mockPackagePath, logger.NewDefaultLogger())
	if err != nil {
		t.Fatalf("Failed to register functions from local path: %s", err)
	}

	// Step 8: Verify function registration
	funcMap := registry.GetRegisteredFunctions()
	if _, exists := funcMap["localFunc"]; !exists {
		t.Errorf("Expected function 'localFunc' to be registered, but it was not found.")
	} else {
		fmt.Println("Successfully registered function: localFunc")
	}
}

func TestLoadFunctionProvider_InvalidProvider(t *testing.T) {
	logger := logger.NewDefaultLogger()

	t.Run("MissingProviderFile", func(t *testing.T) {
		expectedError := "invalid function provider: missing required provider.go file"
		runInvalidProviderTest(t, logger, "", expectedError)
	})

	t.Run("MissingProviderVariable", func(t *testing.T) {
		providerCode := `package mockprovider

import "fmt"

func NotAProvider() {
	fmt.Println("I am not a valid provider")
}`
		expectedError := "invalid function provider: Missing required exported Provider variable"
		runInvalidProviderTest(t, logger, providerCode, expectedError)
	})

	t.Run("UnexportedProviderVariable", func(t *testing.T) {
		providerCode := `package mockprovider

import (
	"text/template"
	"github.com/indaco/tempo-api/templatefuncs"
)

type MockProvider struct{}

func (p *MockProvider) GetFunctions() template.FuncMap {
	return template.FuncMap{"testFunc": func() string { return "Hello!" }}
}

var provider templatefuncs.TemplateFuncProvider = &MockProvider{} // 🚫 Unexported
`
		expectedError := "invalid function provider: Missing required exported Provider variable"
		runInvalidProviderTest(t, logger, providerCode, expectedError)
	})

	t.Run("ProviderIsNil", func(t *testing.T) {
		providerCode := `package mockprovider

import "github.com/indaco/tempo-api/templatefuncs"

var Provider templatefuncs.TemplateFuncProvider = nil
`
		expectedError := "invalid function provider: Provider is nil"
		runInvalidProviderTest(t, logger, providerCode, expectedError)
	})

	t.Run("ProviderDoesNotImplementTemplateFuncProvider", func(t *testing.T) {
		providerCode := `package mockprovider

import "github.com/indaco/tempo-api/templatefuncs"

type MockProvider struct{} // 🚫 Missing GetFunctions method

var Provider templatefuncs.TemplateFuncProvider = &MockProvider{}
`
		expectedError := "invalid function provider: Provider does not implement TemplateFuncProvider (missing method GetFunctions)"
		runInvalidProviderTest(t, logger, providerCode, expectedError)
	})

	t.Run("NoFunctionsInProvider", func(t *testing.T) {
		providerCode := `package mockprovider

import (
	"text/template"
	"github.com/indaco/tempo-api/templatefuncs"
)

type MockProvider struct{}

func (p *MockProvider) GetFunctions() template.FuncMap {
	return template.FuncMap{}
}

var Provider templatefuncs.TemplateFuncProvider = &MockProvider{}
`
		expectedError := "failed to import package: invalid function provider: No functions found in Provider"
		runInvalidProviderTest(t, logger, providerCode, expectedError)
	})
}

func runInvalidProviderTest(t *testing.T, logger logger.LoggerInterface, providerCode, expectedError string) {
	tempDir, err := os.MkdirTemp("", "invalid_provider_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			log.Printf("Failed to remove test directory %s: %v", tempDir, err)
		}
	}()

	// Initialize Go module
	cmd := exec.Command("go", "mod", "init", "mockprovider")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize go module inside mockprovider: %v", err)
	}

	// If providerCode is provided, create provider.go
	if providerCode != "" {
		providerFile := filepath.Join(tempDir, "provider.go")
		if err := os.WriteFile(providerFile, []byte(providerCode), 0644); err != nil {
			t.Fatalf("Failed to create provider file: %v", err)
		}

		// Run `go mod tidy`
		cmd = exec.Command("go", "mod", "tidy")
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to run `go mod tidy` inside invalid provider: %v", err)
		}
	}

	// Run the test
	err = RegisterFunctionsFromPath(tempDir, logger)
	if err == nil {
		t.Fatalf("Expected error: %q, but got nil", expectedError)
	}

	// Check if the error contains the expected message
	if !utils.ContainsSubstring(err.Error(), expectedError) {
		t.Fatalf("Expected error: %q, but got: %q", expectedError, err.Error())
	}
}

func TestLoadFunctionProvider_WithSeparateFunctionFile(t *testing.T) {

	// Create a temporary directory to simulate a standalone Go module.
	tempDir := t.TempDir()

	// Define `provider.go` (declares `Provider`, but no function implementation)
	providerCode := `package mockprovider

import (
	"text/template"
	"github.com/indaco/tempo-api/templatefuncs"
)

// MockProvider implements TemplateFuncProvider.
type MockProvider struct{}

// GetFunctions returns the function map.
func (p *MockProvider) GetFunctions() template.FuncMap {
	return template.FuncMap{
		"fromFuncsFile": FromFuncsFile, // Function is implemented in funcs.go
	}
}

// Expose Provider as a global variable.
var Provider templatefuncs.TemplateFuncProvider = &MockProvider{}
`

	// Define `funcs.go` (actual function implementation)
	funcsCode := `package mockprovider

// FromFuncsFile is implemented in a separate file.
func FromFuncsFile() string {
	return "Hello from funcs.go!"
}
`

	// Step 1: Create `mockprovider` directory and files
	mockPackagePath := filepath.Join(tempDir, "mockprovider")
	_ = os.MkdirAll(mockPackagePath, 0755)

	providerFile := filepath.Join(mockPackagePath, "provider.go")
	if err := os.WriteFile(providerFile, []byte(providerCode), 0644); err != nil {
		t.Fatalf("Failed to create provider.go: %v", err)
	}

	funcsFile := filepath.Join(mockPackagePath, "funcs.go")
	if err := os.WriteFile(funcsFile, []byte(funcsCode), 0644); err != nil {
		t.Fatalf("Failed to create funcs.go: %v", err)
	}

	// Step 2: Initialize `go.mod` inside `mockprovider`
	cmd := exec.Command("go", "mod", "init", "mockprovider")
	cmd.Dir = mockPackagePath
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to initialize go module inside mockprovider: %v", err)
	}

	// Step 4: Run `go mod tidy` to ensure dependencies are resolved
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = mockPackagePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run `go mod tidy` inside mockprovider: %v", err)
	}

	// Step 5: Attempt to register functions
	err := RegisterFunctionsFromPath(mockPackagePath, logger.NewDefaultLogger())
	if err != nil {
		t.Fatalf("Failed to register functions from local path: %s", err)
	}

	// Step 6: Verify function registration
	funcMap := registry.GetRegisteredFunctions()
	if _, exists := funcMap["fromFuncsFile"]; !exists {
		t.Errorf("Expected function 'fromFuncsFile' to be registered, but it was not found.")
	} else {
		fmt.Println("✅ Successfully registered function: fromFuncsFile (defined in funcs.go)")
	}
}
