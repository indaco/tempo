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
		{"String with spaces", "   ", true},
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

func TestIsValidValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		allowed  []string
		expected bool
	}{
		{
			name:     "Valid value - long",
			value:    "long",
			allowed:  []string{"long", "compact", "json", "none"},
			expected: true,
		},
		{
			name:     "Valid value - compact",
			value:    "compact",
			allowed:  []string{"long", "compact", "json", "none"},
			expected: true,
		},
		{
			name:     "Valid value - json",
			value:    "json",
			allowed:  []string{"long", "compact", "json", "none"},
			expected: true,
		},
		{
			name:     "Valid value - none",
			value:    "none",
			allowed:  []string{"long", "compact", "json", "none"},
			expected: true,
		},
		{
			name:     "Invalid value - empty string",
			value:    "",
			allowed:  []string{"long", "compact", "json", "none"},
			expected: false,
		},
		{
			name:     "Invalid value - typo",
			value:    "lonng",
			allowed:  []string{"long", "compact", "json", "none"},
			expected: false,
		},
		{
			name:     "Invalid value - completely wrong",
			value:    "invalid",
			allowed:  []string{"long", "compact", "json", "none"},
			expected: false,
		},
		{
			name:     "Valid in different set",
			value:    "high",
			allowed:  []string{"low", "medium", "high"},
			expected: true,
		},
		{
			name:     "Invalid in different set",
			value:    "extreme",
			allowed:  []string{"low", "medium", "high"},
			expected: false,
		},
		{
			name:     "Valid value - single option",
			value:    "single",
			allowed:  []string{"single"},
			expected: true,
		},
		{
			name:     "Invalid value - single option mismatch",
			value:    "wrong",
			allowed:  []string{"single"},
			expected: false,
		},
		{
			name:     "No allowed values",
			value:    "anything",
			allowed:  []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidValue(tt.value, tt.allowed)
			if result != tt.expected {
				t.Errorf("IsValidValue(%q, %v) = %v; want %v", tt.value, tt.allowed, result, tt.expected)
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
