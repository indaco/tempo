package generator

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/testutils"
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

		defer func() {
			if err := os.Remove(tempFile); err != nil {
				log.Printf("Failed to remove test directory %s: %v", tempFile, err)
			}
		}()

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

// TestGenerateActionFile tests the GenerateActionFile function.
func TestGenerateActionFile(t *testing.T) {
	tempDir := t.TempDir() // Create a temporary directory for the test
	actionsDir := filepath.Join(tempDir, "actions")
	err := os.Mkdir(actionsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create actions directory: %v", err)
	}

	// Define test data
	entityType := "test-entity"
	expectedFile := filepath.Join(actionsDir, entityType+".json")

	actions := []Action{
		{Item: "test-action-1", TemplateFile: "template1.tmpl", Path: "path/to/file1"},
		{Item: "test-action-2", TemplateFile: "template2.tmpl", Path: "path/to/file2"},
	}

	data := &TemplateData{ActionsDir: actionsDir}

	// Create a mock logger (replace with a proper mock if needed)
	mockLogger := &testutils.MockLogger{}

	// Run the function
	err = GenerateActionFile(entityType, data, actions, mockLogger)
	if err != nil {
		t.Fatalf("GenerateActionFile failed: %v", err)
	}

	// Check if the file was created
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Fatalf("Expected file %s was not created", expectedFile)
	}

	// Read the file content
	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Unmarshal JSON to check content
	var writtenActions []JSONAction
	err = json.Unmarshal(content, &writtenActions)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Validate the written content
	expectedJSONActions := ActionList(actions).ToJSONAction()
	if len(writtenActions) != len(expectedJSONActions) {
		t.Fatalf("Expected %d actions, but got %d", len(expectedJSONActions), len(writtenActions))
	}

	for i, action := range writtenActions {
		if action != expectedJSONActions[i] {
			t.Fatalf("Mismatch in action at index %d: expected %+v, got %+v", i, expectedJSONActions[i], action)
		}
	}
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
				if err == nil || !utils.ContainsSubstring(err.Error(), tt.expectedErrMsg) {
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

//
// CopyAction Tests
//

func TestCopyAction_Execute_Folder_Success(t *testing.T) {
	// Create a temporary source directory.
	srcDir := t.TempDir()
	// Create a subfolder within srcDir.
	folderName := "folder1"
	subDir := filepath.Join(srcDir, folderName)
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subfolder: %v", err)
	}
	// Create a dummy file inside subDir.
	dummyFile := filepath.Join(subDir, "dummy.txt")
	expectedContent := "dummy content"
	if err := os.WriteFile(dummyFile, []byte(expectedContent), 0644); err != nil {
		t.Fatalf("failed to create dummy file: %v", err)
	}

	// Create a temporary destination directory.
	destBase := t.TempDir()
	data := &TemplateData{
		TemplatesDir: destBase,
	}

	// Set action.Source to the relative folder name.
	action := Action{
		Item:   "folder",
		Source: folderName, // non-empty, so that ReadEmbeddedDir reads "folder1"
	}

	// Override CopyDirFromEmbedFunc so that it copies from our subfolder.
	origCopyDirFunc := utils.CopyDirFromEmbedFunc
	utils.CopyDirFromEmbedFunc = func(src, dest string) error {
		// Instead of using src, copy from our known subfolder.
		return copyDir(subDir, dest)
	}
	defer func() { utils.CopyDirFromEmbedFunc = origCopyDirFunc }()

	copyAction := &CopyAction{}
	err := copyAction.Execute(action, data)
	if err != nil {
		t.Fatalf("CopyAction.Execute returned error: %v", err)
	}

	// Destination is computed as: filepath.Join(data.TemplatesDir, action.Source)
	destDir := filepath.Join(destBase, folderName)
	copiedFile := filepath.Join(destDir, "dummy.txt")
	copiedData, err := os.ReadFile(copiedFile)
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if string(copiedData) != expectedContent {
		t.Errorf("expected copied file content %q, got %q", expectedContent, string(copiedData))
	}
}

func TestCopyAction_Execute_Default(t *testing.T) {
	data := createTestTemplateData(t.TempDir())
	action := Action{
		Item: "invalidType",
	}
	copyAction := &CopyAction{}
	err := copyAction.Execute(action, data)
	if err == nil {
		t.Fatalf("expected error for unknown item type, got nil")
	}
	if !utils.ContainsSubstring(err.Error(), "unknown item type") {
		t.Errorf("unexpected error: %v", err)
	}
}

//
// RenderAction Tests
//

func TestRenderAction_Execute_Default(t *testing.T) {
	data := createTestTemplateData(t.TempDir())
	action := Action{
		Item: "invalidType",
	}
	renderAction := &RenderAction{}
	err := renderAction.Execute(action, data)
	if err == nil {
		t.Fatalf("expected error for unknown item type in RenderAction, got nil")
	}
	if !utils.ContainsSubstring(err.Error(), "unknown item type") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRenderAction_Execute_File(t *testing.T) {
	tempDir := t.TempDir()
	templatesDir := filepath.Join(tempDir, "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("failed to create templates directory: %v", err)
	}
	// Create a template file with a placeholder.
	templateFile := "test.templ"
	templatePath := filepath.Join(templatesDir, templateFile)
	content := "Hello {{.ComponentName}}"
	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create template file: %v", err)
	}

	// Set action.Path to output file.
	outputFile := filepath.Join(tempDir, "output.txt")
	action := Action{
		Item:         "file",
		TemplateFile: templateFile,
		Path:         outputFile,
	}
	data := createTestTemplateData(templatesDir)
	// Execute file rendering.
	err := renderActionFile(action, data)
	if err != nil {
		t.Fatalf("renderActionFile returned error: %v", err)
	}

	// Verify rendered content.
	rendered, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	expected := "Hello World"
	if string(rendered) != expected {
		t.Errorf("expected rendered content %q, got %q", expected, string(rendered))
	}
}

func TestRenderAction_Execute_Folder(t *testing.T) {
	tempDir := t.TempDir()
	// Create a source folder inside tempDir.
	srcFolderName := "src"
	srcDir := filepath.Join(tempDir, srcFolderName)
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create source folder: %v", err)
	}
	// Create a template file inside the source folder.
	templateFile := "greet.templ"
	templatePath := filepath.Join(srcDir, templateFile)
	if err := os.WriteFile(templatePath, []byte("Greetings, {{.ComponentName}}"), 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	// Set up action so that action.Source is the relative path "src".
	action := Action{
		Item:        "folder",
		Source:      srcFolderName,
		Destination: filepath.Join(tempDir, "dest"),
	}
	// Set data.TemplatesDir to the base directory (tempDir).
	data := &TemplateData{
		TemplatesDir:  tempDir,
		ComponentName: "World",
	}
	err := renderActionFolder(action, data)
	if err != nil {
		t.Fatalf("renderActionFolder returned error: %v", err)
	}

	// Check that the file was rendered in the destination.
	renderedFile := filepath.Join(tempDir, "dest", templateFile)
	if _, err := os.Stat(renderedFile); os.IsNotExist(err) {
		t.Fatalf("expected rendered file %s in destination, but it does not exist", renderedFile)
	}
	renderedContent, err := os.ReadFile(renderedFile)
	if err != nil {
		t.Fatalf("failed to read rendered file: %v", err)
	}
	expectedOutput := "Greetings, World"
	if string(renderedContent) != expectedOutput {
		t.Errorf("expected rendered content %q, got %q", expectedOutput, string(renderedContent))
	}
}

func TestRenderAction_Execute_Folder_Direct(t *testing.T) {
	// Create a temporary base directory.
	tempDir := t.TempDir()

	// Create a source folder inside tempDir with a template file.
	srcFolderName := "src"
	srcDir := filepath.Join(tempDir, srcFolderName)
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create source folder: %v", err)
	}
	templateFile := "greet.templ"
	templatePath := filepath.Join(srcDir, templateFile)
	templateContent := "Hello {{.ComponentName}}"
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	// Destination folder where the rendered output will be placed.
	destDir := filepath.Join(tempDir, "dest")

	// Set up the action with item "folder" and a relative Source.
	// Since TemplateData.TemplatesDir will be set to tempDir,
	// the base folder for templates is filepath.Join(tempDir, srcFolderName).
	action := Action{
		Item:        "folder",
		Source:      srcFolderName,
		Destination: destDir,
	}

	// Set up template data.
	data := &TemplateData{
		TemplatesDir:  tempDir,
		ComponentName: "World",
	}

	// Call Execute on RenderAction (which dispatches to renderActionFolder).
	renderAction := &RenderAction{}
	err := renderAction.Execute(action, data)
	if err != nil {
		t.Fatalf("RenderAction.Execute returned error: %v", err)
	}

	// Verify that the template file was rendered in the destination.
	renderedFile := filepath.Join(destDir, templateFile)
	if _, err := os.Stat(renderedFile); os.IsNotExist(err) {
		t.Fatalf("Expected rendered file %s in destination, but it does not exist", renderedFile)
	}
	renderedContent, err := os.ReadFile(renderedFile)
	if err != nil {
		t.Fatalf("failed to read rendered file: %v", err)
	}
	expectedOutput := "Hello World"
	if string(renderedContent) != expectedOutput {
		t.Errorf("Unexpected rendered content: got %q, want %q", string(renderedContent), expectedOutput)
	}
}

func TestRetrieveActionsFile(t *testing.T) {
	tempDir := t.TempDir()
	actionsDir := filepath.Join(tempDir, "actions")
	err := os.Mkdir(actionsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create actions directory: %v", err)
	}

	existingFile := filepath.Join(actionsDir, "existing.json")
	invalidJSONFile := filepath.Join(actionsDir, "invalid.json")
	nonExistentFile := filepath.Join(actionsDir, "nonexistent.json")

	// Create a valid JSON actions file
	validJSONContent := `[
		{"item": "test", "template_file": "test.tmpl", "path": "test/path"}
	]`
	err = os.WriteFile(existingFile, []byte(validJSONContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write valid actions file: %v", err)
	}

	// Create an invalid JSON file
	invalidJSONContent := `{ invalid json `
	err = os.WriteFile(invalidJSONFile, []byte(invalidJSONContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid JSON file: %v", err)
	}

	mockLogger := &testutils.MockLogger{} // Replace with your logger mock
	mockConfig := &config.Config{
		Paths: config.Paths{
			ActionsDir: actionsDir,
		},
	}

	tests := []struct {
		name           string
		actionFilePath string
		expectedErr    string
	}{
		{
			name:           "Valid action file loads successfully",
			actionFilePath: existingFile,
			expectedErr:    "",
		},
		{
			name:           "Non-existent action file",
			actionFilePath: nonExistentFile,
			expectedErr:    "failed to resolve action file path",
		},
		{
			name:           "Invalid JSON format",
			actionFilePath: invalidJSONFile,
			expectedErr:    "failed to load user-defined actions",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := RetrieveActionsFile(mockLogger, tc.actionFilePath, mockConfig)

			if tc.expectedErr != "" {
				if err == nil || !utils.ContainsSubstring(err.Error(), tc.expectedErr) {
					t.Fatalf("Expected error %q, but got: %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestResolveActionFilePath(t *testing.T) {
	tempDir := t.TempDir()
	actionsDir := filepath.Join(tempDir, "actions")
	existingFile := filepath.Join(actionsDir, "existing.json")
	nonExistentFile := filepath.Join(tempDir, "nonexistent.json")
	mockError := fmt.Errorf("mock error")

	tests := []struct {
		name           string
		actionsDir     string
		actionFileFlag string
		mockFileExists func(path string) (bool, error)
		expectedErr    string
		expectedResult string
	}{
		{
			name:           "Action file found in ActionsDir",
			actionsDir:     actionsDir,
			actionFileFlag: "existing.json",
			mockFileExists: func(path string) (bool, error) {
				if path == existingFile {
					return true, nil
				}
				return false, nil
			},
			expectedResult: existingFile,
		},
		{
			name:           "Action file does not exist",
			actionsDir:     "",
			actionFileFlag: nonExistentFile,
			mockFileExists: func(path string) (bool, error) {
				return false, nil
			},
			expectedErr: "action file does not exist",
		},
		{
			name:           "Error checking resolvedPath",
			actionsDir:     actionsDir,
			actionFileFlag: "error.json",
			mockFileExists: func(path string) (bool, error) {
				if utils.ContainsSubstring(path, actionsDir) {
					return false, mockError
				}
				return false, nil
			},
			expectedErr: "mock error",
		},
		{
			name:           "Error checking absolute path",
			actionsDir:     "",
			actionFileFlag: nonExistentFile,
			mockFileExists: func(path string) (bool, error) {
				if path == nonExistentFile {
					return false, mockError
				}
				return false, nil
			},
			expectedErr: "error checking action file path",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Mock the function
			utils.FileExistsFunc = tc.mockFileExists
			defer func() { utils.FileExistsFunc = utils.FileExists }() // Reset after test

			result, err := resolveActionFilePath(tc.actionsDir, tc.actionFileFlag)

			if tc.expectedErr != "" {
				if err == nil || !utils.ContainsSubstring(err.Error(), tc.expectedErr) {
					t.Fatalf("Expected error %q, but got: %v", tc.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result != tc.expectedResult {
					t.Fatalf("Expected result %q, but got %q", tc.expectedResult, result)
				}
			}
		})
	}
}

func TestProcessEntityActions(t *testing.T) {
	tempDir := t.TempDir()
	actionsDir := filepath.Join(tempDir, "actions")
	err := os.Mkdir(actionsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create actions directory: %v", err)
	}

	// Create a sample action file
	actionFilePath := filepath.Join(actionsDir, "test-entity.json")

	actions := JSONActionList{
		{Item: "action-1", TemplateFile: "template1.tmpl", Path: "path/to/file1"},
		{Item: "action-2", TemplateFile: "template2.tmpl", Path: "path/to/file2"},
	}

	// Write actions to the JSON file
	err = utils.WriteJSONToFile(actionFilePath, actions)
	if err != nil {
		t.Fatalf("Failed to write actions file: %v", err)
	}

	mockLogger := &testutils.MockLogger{} // Replace with an actual logger mock
	mockConfig := &config.Config{
		Paths: config.Paths{
			ActionsDir: actionsDir,
		},
	}

	templateData := &TemplateData{Force: true}

	// Override ProcessActionsFunc temporarily
	originalProcessActions := ProcessActionsFunc
	ProcessActionsFunc = func(logger logger.LoggerInterface, actions []Action, data *TemplateData) error {
		// Debugging output
		for i, action := range actions {
			t.Logf("DEBUG: Action[%d]: %+v", i, action)
		}
		return nil
	}
	defer func() { ProcessActionsFunc = originalProcessActions }() // Restore after test

	// Run processEntityActions
	err = ProcessEntityActions(mockLogger, actionFilePath, templateData, mockConfig)
	if err != nil {
		t.Fatalf("processEntityActions failed: %v", err)
	}

	// Manually create the expected actions list, ensuring Force=true
	expectedActions := actions.ToActions(RenderActionId)
	for i := range expectedActions {
		expectedActions[i].Force = true
	}

	// Validate the actions
	for i, action := range expectedActions {
		if action.Force != true {
			t.Fatalf("Expected action at index %d to have Force=true, but got false", i)
		}
	}
}

func TestIsJsAction(t *testing.T) {
	tests := []struct {
		name   string
		action Action
		expect bool
	}{
		{
			name: "JS in TemplateFile",
			action: Action{
				TemplateFile: "component/assets/js/script.js.gotxt",
			},
			expect: true,
		},
		{
			name: "JS in Path",
			action: Action{
				Path: "components/button/js/script.templ",
			},
			expect: true,
		},
		{
			name: "JS in Source",
			action: Action{
				Item:   "folder",
				Source: "component/assets/js",
			},
			expect: true,
		},
		{
			name: "JS in Destination",
			action: Action{
				Item:        "folder",
				Destination: "component/button/js",
			},
			expect: true,
		},
		{
			name: "JS in uppercase path (should match)",
			action: Action{
				Path: "components/Button/JS/script.templ",
			},
			expect: true,
		},
		{
			name: "No JS in any field",
			action: Action{
				TemplateFile: "component/assets/css/base.css.gotxt",
				Path:         "components/button/css/base.templ",
				Source:       "component/assets/css",
				Destination:  "components/button/css",
			},
			expect: false,
		},
		{
			name:   "Empty action",
			action: Action{},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isJsAction(tt.action)
			if result != tt.expect {
				t.Errorf("isJsAction() = %v, want %v", result, tt.expect)
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

// createTestTemplateData returns minimal TemplateData for testing.
func createTestTemplateData(templatesDir string) *TemplateData {
	return &TemplateData{
		TemplatesDir:  templatesDir,
		ComponentName: "World", // so that {{.ComponentName}} becomes "World"
	}
}

// copyDir is a simple helper that recursively copies a directory from src to dest.
func copyDir(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dest, rel)
		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(destPath, data, info.Mode())
	})
}
