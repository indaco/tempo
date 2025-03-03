package generator

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/utils"
)

// TestGenerateActionJSONFile tests the GenerateActionJSONFile function.
func TestGenerateActionJSONFile(t *testing.T) {
	t.Run("ValidActions", func(t *testing.T) {
		actions := ActionList{
			{Type: "render", Item: "file", TemplateFile: "template1", Path: "path1"},
			{Type: "copy", Item: "folder", Source: "source1", Destination: "dest1"},
		}

		tempFile := "test_output.json"
		defer os.Remove(tempFile)

		err := GenerateActionJSONFile(tempFile, actions)
		if err != nil {
			t.Fatalf("GenerateActionJSONFile() returned error: %v", err)
		}

		// Read the generated file
		data, err := os.ReadFile(tempFile)
		if err != nil {
			t.Fatalf("Failed to read generated file: %v", err)
		}

		// Define the expected content
		expectedContent := `[{"item":"file","templateFile":"template1","path":"path1"},{"item":"folder","source":"source1","destination":"dest1"}]`

		// Unmarshal both the generated and expected JSON content
		var generatedActions []JSONAction
		var expectedActions []JSONAction

		if err := json.Unmarshal(data, &generatedActions); err != nil {
			t.Fatalf("Failed to unmarshal generated JSON: %v", err)
		}
		if err := json.Unmarshal([]byte(expectedContent), &expectedActions); err != nil {
			t.Fatalf("Failed to unmarshal expected JSON: %v", err)
		}

		// Compare the unmarshaled content
		if !reflect.DeepEqual(generatedActions, expectedActions) {
			t.Errorf("Generated actions = %v; want %v", generatedActions, expectedActions)
		}
	})
}

// TestActionConversion tests the ToJSONAction and ToActions conversion methods.
func TestActionConversion(t *testing.T) {
	t.Run("ToJSONAction", func(t *testing.T) {
		actions := ActionList{
			{Type: "render", Item: "file", TemplateFile: "template1", Path: "path1"},
		}

		expectedJSONActions := []JSONAction{
			{Item: "file", TemplateFile: "template1", Path: "path1"},
		}

		if !reflect.DeepEqual(actions.ToJSONAction(), expectedJSONActions) {
			t.Errorf("ActionList.ToJSONAction() = %v; want %v", actions.ToJSONAction(), expectedJSONActions)
		}
	})

	t.Run("ToActions", func(t *testing.T) {
		jsonActions := JSONActionList{
			{Item: "file", TemplateFile: "template1", Path: "path1"},
		}

		expectedActions := []Action{
			{Type: "render", Item: "file", TemplateFile: "template1", Path: "path1"},
		}

		if !reflect.DeepEqual(jsonActions.ToActions("render"), expectedActions) {
			t.Errorf("JSONActionList.ToActions() = %v; want %v", jsonActions.ToActions("render"), expectedActions)
		}
	})
}

func TestLoadUserActions(t *testing.T) {
	testDataDir := path.Join("..", "..", "testdata")
	validFile := filepath.Join(testDataDir, "valid_actions.json")
	invalidFile := filepath.Join(testDataDir, "invalid_actions.json")
	nonExistentFile := filepath.Join(testDataDir, "nonexistent.json")

	t.Run("ValidActionsFile", func(t *testing.T) {
		actions, err := LoadUserActions(validFile)
		if err != nil {
			t.Fatalf("LoadUserActions(%q) returned error: %v; want nil", validFile, err)
		}

		if len(actions) != 2 {
			t.Fatalf("expected 2 actions, got %d", len(actions))
		}

		expectedItems := []string{"file", "folder"}
		for i, action := range actions {
			if action.Item != expectedItems[i] {
				t.Errorf("expected action item %q, got %q", expectedItems[i], action.Item)
			}
		}
	})

	t.Run("InvalidActionsFile", func(t *testing.T) {
		_, err := LoadUserActions(invalidFile)
		if err == nil {
			t.Fatalf("LoadUserActions(%q) returned nil; want error", invalidFile)
		}
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		_, err := LoadUserActions(nonExistentFile)
		if err == nil {
			t.Fatalf("LoadUserActions(%q) returned nil; want error", nonExistentFile)
		}

		expectedError := "actions file '" + nonExistentFile + "' does not exist or is a directory"
		if err.Error() != expectedError {
			t.Errorf("unexpected error: got %q, want %q", err.Error(), expectedError)
		}
	})
}

func TestExecute_CopyAction(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{TempoRoot: tempDir}

	action := Action{
		Item:         "file",
		TemplateFile: "test.txt",
	}

	copyAction := CopyAction{}

	err := copyAction.Execute(action, &TemplateData{TemplatesDir: cfg.TempoRoot})
	if err == nil {
		t.Errorf("Expected error due to missing template file, but got none")
	}
}

func TestExecute_RenderAction(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{TempoRoot: tempDir}

	action := Action{
		Item:         "file",
		TemplateFile: "template.templ",
		Path:         filepath.Join(tempDir, "output.txt"),
	}

	renderAction := RenderAction{}

	err := renderAction.Execute(action, &TemplateData{TemplatesDir: cfg.TempoRoot})
	if err == nil {
		t.Errorf("Expected error due to missing template file, but got none")
	}
}

func TestRenderActionFile(t *testing.T) {
	tempDir := t.TempDir()
	templateFile := filepath.Join(tempDir, "test.templ")
	outputFile := filepath.Join(tempDir, "output.txt")

	// Create a valid template file
	content := "Hello {{.ComponentName}}"
	if err := os.WriteFile(templateFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Verify template file was created correctly

	if !fileExists(templateFile) {
		t.Fatalf("Template file does not exist at path: %s", templateFile)
	}

	fileContent, err := getFileContent(templateFile)
	if err != nil {
		t.Fatalf("Failed to read template file before rendering: %v", err)
	}
	t.Logf("Template File Content:\n%s", string(fileContent))

	// Define action and template data
	action := Action{
		TemplateFile: templateFile,
		Path:         outputFile,
	}
	data := &TemplateData{
		ComponentName: "World", // 🔹 Ensure this matches the template placeholder
	}

	// Render the file
	err = renderActionFile(action, data)
	if err != nil {
		t.Fatalf("Unexpected error rendering action file: %v\nTemplate Path: %s\nFile Content: %s", err, templateFile, content)
	}

	// Verify the rendered file was created
	renderedData, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read rendered file: %v", err)
	}

	expectedOutput := "Hello World"
	if string(renderedData) != expectedOutput {
		t.Errorf("Unexpected rendered content: got %q, expected %q", string(renderedData), expectedOutput)
	}
}

func TestRenderActionFolder(t *testing.T) {
	tempDir := t.TempDir()
	baseDir := filepath.Join(tempDir, "templates")
	destDir := filepath.Join(tempDir, "output")

	// Create a mock template folder with files
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		t.Fatalf("Failed to create base template folder: %v", err)
	}

	// Create a mock template file (test.templ.gotxt)
	templateFile := "test.templ.gotxt"
	templateFilePath := filepath.Join(baseDir, templateFile)
	if err := os.WriteFile(templateFilePath, []byte("Hello {{.ComponentName}}"), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	files, err := os.ReadDir(baseDir)
	if err != nil {
		t.Fatalf("Failed to read base template folder: %v", err)
	}

	// Check if the template directory contains files
	if len(files) == 0 {
		t.Fatalf("No files found in base template folder: %s", baseDir)
	}

	// Define action and template data
	action := Action{
		Item:        "folder",
		Source:      baseDir,
		Destination: destDir,
	}
	data := &TemplateData{
		ComponentName: "World",
	}

	// Run renderActionFolder
	err = renderActionFolder(action, data)
	if err != nil {
		t.Fatalf("Unexpected error rendering folder: %v", err)
	}

	// Verify files exist in the destination
	expectedFilename := utils.RemoveTemplatingExtension(templateFile, config.DefaultTemplateExtensions)
	outputFile := filepath.Join(destDir, expectedFilename)
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Rendered file not found in destination: %s", outputFile)
	}

	// Read and validate rendered file content
	renderedData, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read rendered file: %v", err)
	}

	expectedOutput := "Hello World"
	if string(renderedData) != expectedOutput {
		t.Errorf("Unexpected rendered content: got %q, expected %q", string(renderedData), expectedOutput)
	}
}

// TestHandleOutputFile tests error handling in handleOutputFile
func TestHandleOutputFile(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output.txt")

	tests := []struct {
		name           string
		setup          func()
		fileExistsFunc func(string) (bool, bool, error)
		writeFileFunc  func(string, string) error
		action         Action
		expectedErrMsg string
	}{
		{
			name: "Error checking output file",
			fileExistsFunc: func(_ string) (bool, bool, error) {
				return false, false, errors.New("mock error checking file")
			},
			writeFileFunc:  utils.WriteStringToFile, // Use real function
			expectedErrMsg: "error checking output file",
		},
		{
			name: "Output path exists but is a directory",
			fileExistsFunc: func(path string) (bool, bool, error) {
				return true, true, nil // Simulate existing directory
			},
			writeFileFunc:  utils.WriteStringToFile,
			expectedErrMsg: "output path exists but is a directory",
		},
		{
			name: "Output file exists and SkipIfExists is enabled (should succeed)",
			setup: func() {
				if err := os.WriteFile(outputFile, []byte("existing content"), 0644); err != nil {
					t.Fatalf("Failed to create test output file: %v", err)
				}
			},
			fileExistsFunc: func(path string) (bool, bool, error) {
				if path == outputFile {
					return true, false, nil // Simulate existing file
				}
				return utils.FileOrDirExists(path)
			},
			writeFileFunc: utils.WriteStringToFile,
			action:        Action{SkipIfExists: true},
		},
		{
			name: "Output file exists but Force is not enabled",
			setup: func() {
				if err := os.WriteFile(outputFile, []byte("existing content"), 0644); err != nil {
					t.Fatalf("Failed to create test output file: %v", err)
				}
			},
			fileExistsFunc: func(path string) (bool, bool, error) {
				if path == outputFile {
					return true, false, nil // Simulate existing file
				}
				return utils.FileOrDirExists(path)
			},
			writeFileFunc:  utils.WriteStringToFile,
			action:         Action{Force: false},
			expectedErrMsg: "file already exists and force is not set",
		},
		{
			name: "Error writing rendered content",
			fileExistsFunc: func(_ string) (bool, bool, error) {
				return false, false, nil // File does not exist
			},
			writeFileFunc: func(_ string, _ string) error {
				return errors.New("mock write error")
			},
			expectedErrMsg: "failed to write rendered content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare test setup
			if tt.setup != nil {
				tt.setup()
			}

			// Run handleOutputFile with mock dependencies
			err := handleOutputFile(outputFile, "test content", tt.action, tt.fileExistsFunc, tt.writeFileFunc)

			// Validate error messages if expected
			if tt.expectedErrMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("Expected error containing %q, got %v", tt.expectedErrMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// UTILITY FUNCTIONS

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getFileContent(path string) ([]byte, error) {
	return os.ReadFile(path)
}
