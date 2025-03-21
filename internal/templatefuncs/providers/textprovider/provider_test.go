package textprovider

import "testing"

func TestTextProvider(t *testing.T) {
	provider := Provider
	funcs := provider.GetFunctions()

	if _, exists := funcs["normalizePath"]; !exists {
		t.Errorf("Expected function 'normalizePath' to be registered, but it was not found.")
	}

	if _, exists := funcs["isEmpty"]; !exists {
		t.Errorf("Expected function 'isEmpty' to be registered, but it was not found.")
	}

	if _, exists := funcs["titleCase"]; !exists {
		t.Errorf("Expected function 'titleCase' to be registered, but it was not found.")
	}
}
