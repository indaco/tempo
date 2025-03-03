package textprovider

import "testing"

func TestIsEmptyString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Empty string", "", true},
		{"Non-empty string", "hello", false},
		{"String with spaces", "   ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmptyString(tt.input)
			if result != tt.expected {
				t.Errorf("IsEmptyString(%q) = %v; expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple path", "folder/subfolder/file", "folder/subfolder/file"},
		{"Path with leading slash", "/folder/subfolder/file", "folder/subfolder/file"},
		{"Path with trailing slash", "folder/subfolder/file/", "folder/subfolder/file"},
		{"Path with leading and trailing slashes", "/folder/subfolder/file/", "folder/subfolder/file"},
		{"Path with spaces", "  /folder/subfolder/file  ", "folder/subfolder/file"},
		{"Only dots", "...", ""},
		{"Only slashes", "///", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePath(%q) = %q; expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTitleCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"world", "World"},
		{"Go", "Go"},             // Already capitalized
		{"123abc", "123abc"},     // Starts with a number
		{"@mention", "@mention"}, // Special character
		{"éxample", "Éxample"},   // Unicode
		{"", ""},                 // Empty string
	}

	for _, test := range tests {
		result := TitleCase(test.input)
		if result != test.expected {
			t.Errorf("CapitalizeWord(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}
