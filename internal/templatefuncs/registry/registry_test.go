package registry

import (
	"testing"
	"text/template"
)

// Mock provider for testing
type MockProvider struct{}

func (p *MockProvider) GetFunctions() template.FuncMap {
	return template.FuncMap{
		"mockFunction": func() string { return "mocked!" },
	}
}

// TestRegisterFuncProvider ensures providers are correctly registered
func TestRegisterFuncProvider(t *testing.T) {
	// Clear registry before testing
	funcRegistry = template.FuncMap{}

	mockProvider := &MockProvider{}
	RegisterFuncProvider(mockProvider)

	// Check if mock function exists in registry
	if _, exists := funcRegistry["mockFunction"]; !exists {
		t.Errorf("Expected function 'mockFunction' to be registered, but it wasn't")
	}
}

// TestRegisterFunction ensures individual function registration works
func TestRegisterFunction(t *testing.T) {
	// Clear registry before testing
	funcRegistry = template.FuncMap{}

	mockFunc := func() string { return "Hello, Test!" }
	RegisterFunction("testFunc", mockFunc)

	// Check if function was registered
	if _, exists := funcRegistry["testFunc"]; !exists {
		t.Errorf("Expected function 'testFunc' to be registered, but it wasn't")
	}
}

// TestGetRegisteredFunctions ensures retrieving all functions works
func TestGetRegisteredFunctions(t *testing.T) {
	// Clear registry before testing
	funcRegistry = template.FuncMap{}

	// Register a test function
	mockFunc := func() string { return "Hello, Test!" }
	RegisterFunction("testFunc", mockFunc)

	funcs := GetRegisteredFunctions()
	if _, exists := funcs["testFunc"]; !exists {
		t.Errorf("Expected function 'testFunc' in registered functions, but it wasn't found")
	}
}
