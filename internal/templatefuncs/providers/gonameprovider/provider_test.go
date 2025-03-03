package gonameprovider

import "testing"

func TestDefaultProvider(t *testing.T) {
	provider := Provider
	funcs := provider.GetFunctions()

	if _, exists := funcs["goPackageName"]; !exists {
		t.Errorf("Expected function 'goPackageName' to be registered, but it was not found.")
	}

	if _, exists := funcs["goExportedName"]; !exists {
		t.Errorf("Expected function 'goExportedName' to be registered, but it was not found.")
	}

	if _, exists := funcs["goUnexportedName"]; !exists {
		t.Errorf("Expected function 'goUnexportedName' to be registered, but it was not found.")
	}
}
