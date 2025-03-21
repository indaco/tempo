package lookupprovider

import (
	"reflect"
	"testing"
)

func TestLookup_ValidKeys(t *testing.T) {
	testData := map[string]any{
		"config": map[string]any{
			"option1": "value1",
			"option2": "value2",
			"nested": map[string]any{
				"key": "deep_value",
			},
		},
		"author": "John Doe",
		"date":   "2025-03-20",
	}

	tests := []struct {
		name     string
		key      string
		expected any
	}{
		{"Single Level Key", "author", "John Doe"},
		{"Nested Key", "config.option1", "value1"},
		{"Deeply Nested Key", "config.nested.key", "deep_value"},
		{"Another Single Level Key", "date", "2025-03-20"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Lookup(testData, tt.key)
			if result != tt.expected {
				t.Errorf("lookup(%q) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestLookup_MissingKeys(t *testing.T) {
	testData := map[string]any{
		"config": map[string]any{
			"option1": "value1",
		},
	}

	tests := []struct {
		name     string
		key      string
		expected any
	}{
		{"Missing Key", "config.unknown", nil},
		{"Missing Nested Key", "config.nested.unknown", nil},
		{"Completely Missing Key", "unknownKey", nil},
		{"Accessing Beyond a String", "config.option1.value", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Lookup(testData, tt.key)
			if result != tt.expected {
				t.Errorf("lookup(%q) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

func TestLookup_EdgeCases(t *testing.T) {
	emptyData := map[string]any{}

	tests := []struct {
		name     string
		data     map[string]any
		key      string
		expected any
	}{
		{"Empty Map", emptyData, "someKey", nil},
		{"Empty Key", map[string]any{"a": "b"}, "", map[string]any{"a": "b"}}, // Should return full map
		{"Leading Dot", map[string]any{"config": "value"}, ".config", nil},
		{"Trailing Dot", map[string]any{"config": "value"}, "config.", nil},
		{"Key with Only Dots", map[string]any{"config": "value"}, "...", nil},
		//{"Non-String Key", map[string]any{123: "numberKey"}, "123", nil}, // Ensure it doesn't panic
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Lookup(tt.data, tt.key)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("lookup(%q) = %v, want %v", tt.key, result, tt.expected)
			}
		})
	}
}

// func BenchmarkLookup(b *testing.B) {
// 	testData := map[string]any{
// 		"config": map[string]any{
// 			"option1": "value1",
// 			"nested": map[string]any{
// 				"key": "deep_value",
// 			},
// 		},
// 		"author": "John Doe",
// 		"date":   "2025-03-20",
// 	}

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_ = Lookup(testData, "config.nested.key")
// 	}
// }
