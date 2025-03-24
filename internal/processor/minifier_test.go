package processor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/evanw/esbuild/pkg/api"
	"github.com/indaco/tempo/internal/processor/transformers"
	"github.com/indaco/tempo/internal/testutils"
	"github.com/indaco/tempo/internal/utils"
)

func TestMinifierProcessor_Process(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		name          string
		inputFile     string
		outputFile    string
		loader        api.Loader
		inputContent  string
		expectedMin   string
		expectedError bool
	}{
		{
			name:          "JavaScript minification",
			inputFile:     "input.js",
			outputFile:    "output.templ",
			loader:        api.LoaderJS,
			inputContent:  "function test() { console.log('Hello'); }",
			expectedMin:   "function test(){console.log(\"Hello\")}",
			expectedError: false,
		},
		{
			name:          "CSS minification",
			inputFile:     "input.css",
			outputFile:    "output.templ",
			loader:        api.LoaderCSS,
			inputContent:  "body { color: black; font-size: 16px; }",
			expectedMin:   "body{color:#000;font-size:16px}",
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputFilePath := filepath.Join(tempDir, tc.inputFile)
			outputFilePath := filepath.Join(tempDir, tc.outputFile)

			// Pre-create output file with guard markers
			err := os.WriteFile(outputFilePath, []byte("/* [tempo] BEGIN - Do not edit! This section is auto-generated. */\n/* [tempo] END */"), 0644)
			if err != nil {
				t.Fatalf("Failed to pre-create output file: %v", err)
			}

			// Write test input file
			err = os.WriteFile(inputFilePath, []byte(tc.inputContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create input file: %v", err)
			}

			// Initialize the transformer with correct Loader
			transformer := &transformers.EsbuildTransformer{Loader: tc.loader}
			processor := MinifierProcessor{Transform: transformer.Transform}

			// Run the process function
			err = processor.Process(inputFilePath, outputFilePath, "tempo")

			// Validate errors
			if tc.expectedError {
				if err == nil {
					t.Fatalf("Expected an error but got none")
				}
				return // No further validation needed
			}

			// Ensure no error occurred
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Read the output file
			outputContent, err := os.ReadFile(outputFilePath)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			// Ensure the output contains the expected minified content
			if !testutils.ContainsBetweenMarkers(string(outputContent), tc.expectedMin) {
				t.Errorf("Minified output mismatch.\nExpected:\n%s\nGot:\n%s", tc.expectedMin, string(outputContent))
			}
		})
	}
}

func TestMinifierProcessor_PreservesGuardMarkers(t *testing.T) {
	tempDir := t.TempDir()
	inputFilePath := filepath.Join(tempDir, "input.js")
	outputFilePath := filepath.Join(tempDir, "output.templ")

	inputContent := `function test() {
    console.log("Test");
}`
	expectedContent := `function test(){console.log("Test")}`

	// Pre-create output file with guard markers
	outputTemplate := `/* [tempo] BEGIN - Do not edit! This section is auto-generated. */
/* [tempo] END */`
	if err := os.WriteFile(outputFilePath, []byte(outputTemplate), 0644); err != nil {
		t.Fatalf("Failed to pre-create output file: %v", err)
	}

	// Write test input file
	if err := os.WriteFile(inputFilePath, []byte(inputContent), 0644); err != nil {
		t.Fatalf("Failed to create input file: %v", err)
	}

	// Create a MinifierProcessor with EsbuildTransformer
	transformer := &transformers.EsbuildTransformer{Loader: GetLoader(filepath.Ext(inputFilePath))}
	processor := MinifierProcessor{Transform: transformer.Transform}

	// Run minification process
	err := processor.Process(inputFilePath, outputFilePath, "tempo")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Read the output file
	resultContent, err := os.ReadFile(outputFilePath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Validate that guard markers are preserved and content is minified
	if !testutils.ContainsBetweenMarkers(string(resultContent), expectedContent) {
		t.Errorf("Guard markers not preserved. Expected content between markers:\n%s\nGot:\n%s", expectedContent, string(resultContent))
	}
}

func TestMinifierProcessor_FailsOnMissingInputFile(t *testing.T) {
	tempDir := t.TempDir()
	inputFilePath := filepath.Join(tempDir, "non_existent.js") // Non-existent input file
	outputFilePath := filepath.Join(tempDir, "output.templ")

	// Pre-create output file (since it's a prerequisite)
	outputTemplate := ""
	if err := os.WriteFile(outputFilePath, []byte(outputTemplate), 0644); err != nil {
		t.Fatalf("Failed to pre-create output file: %v", err)
	}

	// Create a MinifierProcessor with EsbuildTransformer
	transformer := &transformers.EsbuildTransformer{Loader: GetLoader(filepath.Ext(inputFilePath))}
	processor := MinifierProcessor{Transform: transformer.Transform}

	// Run minification process with a missing input file
	err := processor.Process(inputFilePath, outputFilePath, "tempo")

	// Expect an error due to missing input file
	if err == nil {
		t.Fatalf("Expected error for missing input file, but got none")
	}

	// Validate error message
	expectedError := "failed to read input file"
	if !utils.ContainsSubstring(err.Error(), expectedError) {
		t.Errorf("Expected error to contain %q, but got: %q", expectedError, err.Error())
	}
}
