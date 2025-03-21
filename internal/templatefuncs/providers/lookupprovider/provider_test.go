package lookupprovider

import "testing"

func TestLookupProvider(t *testing.T) {
	provider := Provider
	funcs := provider.GetFunctions()

	if _, exists := funcs["lookup"]; !exists {
		t.Errorf("Expected function 'lookup' to be registered, but it was not found.")
	}
}
