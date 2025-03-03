package gonameprovider

import "testing"

func TestToGoPackageName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"MyPackage", "my_package"},
		{"my-package", "my_package"},
		{"my package name", "my_package_name"},
		{"123InvalidName", "_123_invalid_name"},
		{"snake_case_example", "snake_case_example"},
		{"PascalCase", "pascal_case"},
		{"", ""},
		{"Special$Character&Name", "special_character_name"},
		{" leading-and-trailing ", "leading_and_trailing"},
		{"pkg/project/internal", "pkg/project/internal"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToGoPackageName(tt.input)
			if result != tt.expected {
				t.Errorf("ToGoPackageName(%q) = %q; expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToGoExportedName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123invalid-name", "Invalid123invalidName"},
		{"PascalCase", "PascalCase"},
		{"special$name", "SpecialName"},
		{"my-variable", "MyVariable"},
		{"123", "Invalid123"},
		{"snake_case_example", "SnakeCaseExample"},
		{"", ""},
		{"multiple---words", "MultipleWords"},
		{"trailing-special!!!", "TrailingSpecial"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToGoExportedName(tt.input)
			if result != tt.expected {
				t.Errorf("ToExportedName(%q) = %q; expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToGoUnexportedName(t *testing.T) {
	tests := map[string]string{
		"PascalCase":             "pascalCase",
		"snake_case_example":     "snakeCaseExample",
		"kebab-case-example":     "kebabCaseExample",
		"123invalid-name":        "invalid123invalidName",
		"special$name":           "specialName",
		"UPPER_case":             "upperCase",
		"multi--hyphen":          "multiHyphen",
		"___leading_trailing___": "leadingTrailing",
		"_":                      "",
		"123":                    "invalid123",
		"___":                    "",
		"-hyphen-":               "hyphen",
	}

	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			result := ToGoUnexportedName(input)
			if result != expected {
				t.Errorf("toUnexportedName(%q) = %q; expected %q", input, result, expected)
			}
		})
	}
}
