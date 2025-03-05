package processor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/tempo/internal/testutils"
)

func TestPassthroughProcessor(t *testing.T) {
	// Step 1: Setup a temporary test directory
	tempDir := t.TempDir()
	inputFilePath := filepath.Join(tempDir, "input.css")
	outputFilePath := filepath.Join(tempDir, "output.templ")

	// Step 2: Create test input and output files
	testCases := []struct {
		name             string
		inputContent     string
		outputContent    string
		expectedContent  string
		expectError      bool
		expectedErrorMsg string
	}{
		{
			name:         "Content inserted between guard markers",
			inputContent: ".button { color: blue; }",
			outputContent: `package button

func Button() {
/* [tempo] BEGIN - Do not edit! This section is auto-generated. */
/* [tempo] END */
}`,
			expectedContent: `package button

func Button() {
/* [tempo] BEGIN - Do not edit! This section is auto-generated. */
.button { color: blue; }
/* [tempo] END */
}`,
			expectError: false,
		},
		{
			name:         "Missing BEGIN marker",
			inputContent: ".button { color: blue; }",
			outputContent: `package button

func Button() {
/* [tempo] END */
}`,
			expectError:      true,
			expectedErrorMsg: "invalid or missing guard markers",
		},
		{
			name:         "Missing END marker",
			inputContent: ".button { color: blue; }",
			outputContent: `package button

func Button() {
/* [tempo] BEGIN - Do not edit! This section is auto-generated. */
}`,
			expectError:      true,
			expectedErrorMsg: "invalid or missing guard markers",
		},
		{
			name:         "BEGIN marker after END marker",
			inputContent: ".button { color: blue; }",
			outputContent: `package button

func Button() {
/* [tempo] END */
/* [tempo] BEGIN - Do not edit! This section is auto-generated. */
}`,
			expectError:      true,
			expectedErrorMsg: "invalid or missing guard markers",
		},
	}

	// Step 3: Iterate over test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Write test input file
			if err := os.WriteFile(inputFilePath, []byte(tc.inputContent), 0644); err != nil {
				t.Fatalf("Failed to create input file: %v", err)
			}

			// Write test output file
			if err := os.WriteFile(outputFilePath, []byte(tc.outputContent), 0644); err != nil {
				t.Fatalf("Failed to create output file: %v", err)
			}

			// Run processor
			processor := PassthroughProcessor{}
			err := processor.Process(inputFilePath, outputFilePath, "tempo")

			// Validate error handling
			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				if tc.expectedErrorMsg != "" && err != nil && !testutils.Contains(err.Error(), tc.expectedErrorMsg) {
					t.Errorf("Expected error message to contain %q, but got %q", tc.expectedErrorMsg, err.Error())
				}
				return
			}

			// Step 4: Validate the output file content
			resultContent, err := os.ReadFile(outputFilePath)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			// Compare expected output with actual result
			if string(resultContent) != tc.expectedContent {
				t.Errorf("Output content mismatch:\nExpected:\n%s\nGot:\n%s", tc.expectedContent, string(resultContent))
			}
		})
	}
}

func TestStandardProcessor_MatchingFolderStructure(t *testing.T) {
	tempDir := t.TempDir()
	inputDir := filepath.Join(tempDir, "input_folder")
	outputDir := filepath.Join(tempDir, "output_folder")

	// Step 1: Create Input Folder Structure
	testutils.CreateFile(t, filepath.Join(inputDir, "button", "button.css"), ".button { color: red; font-size: 16px; }")
	testutils.CreateFile(t, filepath.Join(inputDir, "button", "css", "themes", "light.css"), ".theme-light { background-color: white; color: black; }")
	testutils.CreateFile(t, filepath.Join(inputDir, "button", "css", "themes", "dark.css"), ".theme-dark { background-color: black; color: white; }")

	// Step 2: Create Output Folder Structure (with guard markers)
	testutils.CreateFile(t, filepath.Join(outputDir, "button", "button.templ"), testutils.GenerateTemplContent("button"))
	testutils.CreateFile(t, filepath.Join(outputDir, "button", "css", "themes", "light.templ"), testutils.GenerateTemplContent("themes"))
	testutils.CreateFile(t, filepath.Join(outputDir, "button", "css", "themes", "dark.templ"), testutils.GenerateTemplContent("themes"))

	// Step 3: Run Processor on All Files
	processor := &PassthroughProcessor{}

	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			t.Fatalf("Error reading input directory: %v", err)
		}
		if info.IsDir() {
			return nil
		}

		// Compute expected output path
		relPath, _ := filepath.Rel(inputDir, path)
		outputPath := filepath.Join(outputDir, relPath)
		outputPath = outputPath[:len(outputPath)-len(filepath.Ext(outputPath))] + ".templ"

		err = processor.Process(path, outputPath, "tempo")
		if err != nil {
			t.Errorf("Processor failed for %s: %v", path, err)
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}

	// Step 4: Verify Output Files
	testutils.VerifyFileContent(t, filepath.Join(outputDir, "button", "button.templ"), ".button { color: red; font-size: 16px; }")
	testutils.VerifyFileContent(t, filepath.Join(outputDir, "button", "css", "themes", "light.templ"), ".theme-light { background-color: white; color: black; }")
	testutils.VerifyFileContent(t, filepath.Join(outputDir, "button", "css", "themes", "dark.templ"), ".theme-dark { background-color: black; color: white; }")
}

func TestStandardProcessor_FailsOnMissingInputFile(t *testing.T) {
	tempDir := t.TempDir()
	inputFilePath := filepath.Join(tempDir, "missing_input.css") // Non-existent input file
	outputFilePath := filepath.Join(tempDir, "output.templ")

	// Pre-create output file (since it's a prerequisite)
	outputTemplate := ""
	if err := os.WriteFile(outputFilePath, []byte(outputTemplate), 0644); err != nil {
		t.Fatalf("Failed to pre-create output file: %v", err)
	}

	// Run processor with missing input file
	processor := PassthroughProcessor{}
	err := processor.Process(inputFilePath, outputFilePath, "tempo")

	// Expect an error due to missing input file
	if err == nil {
		t.Fatalf("Expected error for missing input file, but got none")
	}

	// Validate error message
	expectedError := "failed to read input file"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain %q, but got: %q", expectedError, err.Error())
	}
}
