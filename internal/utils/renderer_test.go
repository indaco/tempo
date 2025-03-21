package utils

import (
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name            string
		templateContent string
		data            any
		expected        string
		expectError     bool
	}{
		{
			name:            "Basic template rendering",
			templateContent: "Hello, {{ .Name }}!",
			data:            map[string]string{"Name": "World"},
			expected:        "Hello, World!",
			expectError:     false,
		},
		{
			name:            "Custom function - isEmpty",
			templateContent: "IsEmpty: {{ isEmpty .Value }}",
			data:            map[string]string{"Value": ""},
			expected:        "IsEmpty: true",
			expectError:     false,
		},
		{
			name:            "Custom function - Title Case",
			templateContent: "Hello, {{ titleCase .Name }}!",
			data:            map[string]string{"Name": "wonderful world"},
			expected:        "Hello, Wonderful world!",
			expectError:     false,
		},
		{
			name:            "Custom function - snakeToTitle",
			templateContent: "Hello, {{ snakeToTitle .Name }}!",
			data:            map[string]string{"Name": "wonderful_world"},
			expected:        "Hello, Wonderful World!",
			expectError:     false,
		},
		{
			name:            "Custom function - normalizePath",
			templateContent: "Normalized: {{ normalizePath .Path }}",
			data:            map[string]string{"Path": " /folder/./subfolder/../file "},
			expected:        "Normalized: folder/file",
			expectError:     false,
		},
		{
			name:            "Custom function - goPackageName",
			templateContent: "GoPackageName: {{ goPackageName .Input }}",
			data:            map[string]string{"Input": "my-package"},
			expected:        "GoPackageName: my_package",
			expectError:     false,
		},
		{
			name:            "Custom function - goExportedName",
			templateContent: "ExportedName: {{ goExportedName .Input }}",
			data:            map[string]string{"Input": "my-exported_name"},
			expected:        "ExportedName: MyExportedName",
			expectError:     false,
		},
		{
			name:            "Custom function - goUnexportedName",
			templateContent: "UnexportedName: {{ goUnexportedName .Input }}",
			data:            map[string]string{"Input": "my-unexported_name"},
			expected:        "UnexportedName: myUnexportedName",
			expectError:     false,
		},
		{
			name:            "Error in template parsing",
			templateContent: "Hello, {{ .Name ",
			data:            map[string]string{"Name": "World"},
			expected:        "",
			expectError:     true,
		},
		{
			name:            "Error in template execution",
			templateContent: "Invalid {{ .MissingField }}",
			data:            nil, // MissingField is not provided
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := RenderTemplate(tt.templateContent, tt.data)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if output != tt.expected {
				t.Errorf("Unexpected output:\nGot: %q\nWant: %q", output, tt.expected)
			}
		})
	}
}
