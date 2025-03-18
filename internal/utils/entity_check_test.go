package utils

import (
	"strings"
	"testing"

	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/testhelpers"
)

func TestComponentCommand_NewSubCmd_Func_componentNewHandleEntityExistence(t *testing.T) {
	tests := []struct {
		name         string
		entityType   string
		entityName   string
		outputPath   string
		force        bool
		shouldWarn   bool
		expectedPath string
	}{
		{"Component Exists Without Force", "component", "button", "/mock/path/button", false, true, "/mock/path/button/button"},
		{"Component Exists With Force", "component", "button", "/mock/path/button", true, false, "/mock/path/button/button"},
		{"Variant Exists Without Force", "variant", "outline", "/mock/path/button/css/variants/outline.templ", false, true, "/mock/path/button/css/variants/outline.templ"},
		{"Variant Exists With Force", "variant", "outline", "/mock/path/button/css/variants/outline.templ", true, false, "/mock/path/button/css/variants/outline.templ"},
		{"Unknown Entity Type", "unknown", "mystery", "/mock/path/unknown", false, true, "/mock/path/unknown"}, // NEW CASE
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, err := testhelpers.CaptureStdout(func() {
				logger := logger.NewDefaultLogger()
				CheckEntityForNew("component", tc.entityName, tc.outputPath, tc.force, logger)
			})

			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			// Verify the warning or overwrite messages
			if tc.shouldWarn {
				if !strings.Contains(output, "Use '--force' to overwrite it.") {
					t.Errorf("Expected warning message, got: %s", output)
				}
			} else {
				if !strings.Contains(output, "Overwriting due to '--force' flag.") {
					t.Errorf("Expected overwrite message, got: %s", output)
				}
			}

			// Check if the correct path was used in the log output
			if !strings.Contains(output, tc.expectedPath) {
				t.Errorf("Expected path %q in output, but got: %s", tc.expectedPath, output)
			}
		})
	}
}

func TestComponentCommand_DefineSubCmd_Func_HandleEntityExistence(t *testing.T) {
	tests := []struct {
		name        string
		outputPath  string
		force       bool
		expectedMsg []string
	}{
		{
			name:       "Entity exists without force flag",
			outputPath: "/mock/path/to/component",
			force:      false,
			expectedMsg: []string{
				"⚠ Templates for 'component' already exist.",
				"  Use '--force' to overwrite them. Any changes will be lost.",
				"  - path: /mock/path/to/component",
			},
		},
		{
			name:       "Entity exists with force flag",
			outputPath: "/mock/path/to/component",
			force:      true,
			expectedMsg: []string{
				"ℹ Templates for 'component' already exist.",
				"  Overwriting due to '--force' flag.",
				"  - path: /mock/path/to/component",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture logger output
			output, err := testhelpers.CaptureStdout(func() {
				log := logger.NewDefaultLogger()
				CheckEntityForDefine("component", tt.outputPath, tt.force, log)
			})

			if err != nil {
				t.Fatalf("Failed to capture logger output: %v", err)
			}

			// Validate output
			testhelpers.ValidateCLIOutput(t, output, tt.expectedMsg)
		})
	}
}

func TestVariantCommand_DefineSubCmd_Func_HandleEntityExistence(t *testing.T) {
	tests := []struct {
		name        string
		outputPath  string
		force       bool
		expectedMsg []string
	}{
		{
			name:       "Entity exists without force flag",
			outputPath: "/mock/path/to/variant",
			force:      false,
			expectedMsg: []string{
				"⚠ Templates for 'variant' already exist.",
				"  Use '--force' to overwrite them. Any changes will be lost.",
				"  - path: /mock/path/to/variant",
			},
		},
		{
			name:       "Entity exists with force flag",
			outputPath: "/mock/path/to/variant",
			force:      true,
			expectedMsg: []string{
				"ℹ Templates for 'variant' already exist.",
				"  Overwriting due to '--force' flag.",
				"  - path: /mock/path/to/variant",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture logger output
			output, err := testhelpers.CaptureStdout(func() {
				log := logger.NewDefaultLogger()
				CheckEntityForDefine("variant", tt.outputPath, tt.force, log)
			})

			if err != nil {
				t.Fatalf("Failed to capture logger output: %v", err)
			}

			// Validate output
			testhelpers.ValidateCLIOutput(t, output, tt.expectedMsg)
		})
	}
}

// func TestHandleEntityExistence(t *testing.T) {
// 	tests := []struct {
// 		name         string
// 		entityType   string
// 		entityName   string
// 		outputPath   string
// 		force        bool
// 		shouldWarn   bool
// 		expectedPath string
// 	}{
// 		{"Component Exists Without Force", "component", "button", "/mock/path/button", false, true, "/mock/path/button/button"},
// 		{"Component Exists With Force", "component", "button", "/mock/path/button", true, false, "/mock/path/button/button"},
// 		{"Variant Exists Without Force", "variant", "outline", "/mock/path/button/css/variants/outline.templ", false, true, "/mock/path/button/css/variants/outline.templ"},
// 		{"Variant Exists With Force", "variant", "outline", "/mock/path/button/css/variants/outline.templ", true, false, "/mock/path/button/css/variants/outline.templ"},
// 		{"Unknown Entity Type", "unknown", "mystery", "/mock/path/unknown", false, true, "/mock/path/unknown"}, // NEW CASE
// 	}

// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			output, err := testhelpers.CaptureStdout(func() {
// 				logger := logger.NewDefaultLogger()
// 				handleEntityExistence(tc.entityType, tc.entityName, tc.outputPath, tc.force, logger)
// 			})

// 			if err != nil {
// 				t.Fatalf("Failed to capture stdout: %v", err)
// 			}

// 			// Verify the warning or overwrite messages
// 			if tc.shouldWarn {
// 				if !strings.Contains(output, "Use '--force' to overwrite it.") {
// 					t.Errorf("Expected warning message, got: %s", output)
// 				}
// 			} else {
// 				if !strings.Contains(output, "Overwriting due to '--force' flag.") {
// 					t.Errorf("Expected overwrite message, got: %s", output)
// 				}
// 			}

// 			// Check if the correct path was used in the log output
// 			if !strings.Contains(output, tc.expectedPath) {
// 				t.Errorf("Expected path %q in output, but got: %s", tc.expectedPath, output)
// 			}
// 		})
// 	}
// }
