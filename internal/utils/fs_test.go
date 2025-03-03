package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetCWD(t *testing.T) {
	// Get the current working directory using os.Getwd for comparison
	expectedCWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	// Call the GetCWD function
	actualCWD := GetCWD()

	// Compare the output
	if actualCWD != expectedCWD {
		t.Errorf("Unexpected working directory:\nGot: %s\nWant: %s", actualCWD, expectedCWD)
	}
}

func TestFileOrDirExists(t *testing.T) {
	// Temporary paths setup
	tempFile := "testfile.tmp"
	tempDir := "testdir"
	invalidPath := ""

	// Cleanup function to ensure temporary files and directories are removed
	cleanup := func() {
		os.Remove(tempFile)
		os.Remove(tempDir)
	}
	defer cleanup()

	// Test case: Path is a file
	t.Run("PathIsFile", func(t *testing.T) {
		// Create a temporary file
		file, err := os.Create(tempFile)
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		file.Close()

		// Check if the file exists and is not a directory
		exists, isDir, err := FileOrDirExists(tempFile)
		if err != nil {
			t.Errorf("FileOrDirExists(%q) returned error: %v; want nil", tempFile, err)
		}
		if !exists || isDir {
			t.Errorf("FileOrDirExists(%q) = (%t, %t); want (true, false)", tempFile, exists, isDir)
		}
	})

	// Test case: Path is a directory
	t.Run("PathIsDirectory", func(t *testing.T) {
		// Create a temporary directory
		err := os.Mkdir(tempDir, 0755)
		if err != nil {
			t.Fatalf("failed to create temp directory: %v", err)
		}

		// Check if the directory exists
		exists, isDir, err := FileOrDirExists(tempDir)
		if err != nil {
			t.Errorf("FileOrDirExists(%q) returned error: %v; want nil", tempDir, err)
		}
		if !exists || !isDir {
			t.Errorf("FileOrDirExists(%q) = (%t, %t); want (true, true)", tempDir, exists, isDir)
		}
	})

	// Test case: Nonexistent path
	t.Run("PathDoesNotExist", func(t *testing.T) {
		nonExistentPath := "nonexistent.tmp"
		exists, isDir, err := FileOrDirExists(nonExistentPath)
		if err != nil {
			t.Errorf("FileOrDirExists(%q) returned error: %v; want nil", nonExistentPath, err)
		}
		if exists || isDir {
			t.Errorf("FileOrDirExists(%q) = (%t, %t); want (false, false)", nonExistentPath, exists, isDir)
		}
	})

	// Test case: Invalid path
	t.Run("InvalidPath", func(t *testing.T) {
		exists, isDir, err := FileOrDirExists(invalidPath)
		if err == nil {
			t.Errorf("FileOrDirExists(%q) returned no error; want error", invalidPath)
		}
		if exists || isDir {
			t.Errorf("FileOrDirExists(%q) = (%t, %t); want (false, false)", invalidPath, exists, isDir)
		}
	})
}

func TestFileExists(t *testing.T) {
	t.Run("File exists", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test_file.txt")

		// Create a temporary file
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Check if the file exists
		exists, err := FileExists(testFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !exists {
			t.Errorf("Expected file to exist, but it does not")
		}
	})

	t.Run("File does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		nonExistentFile := filepath.Join(tempDir, "non_existent.txt")

		// Check for a non-existent file
		exists, err := FileExists(nonExistentFile)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if exists {
			t.Errorf("Expected file to not exist, but it does")
		}
	})

	t.Run("Path is a directory", func(t *testing.T) {
		tempDir := t.TempDir()

		// Check if the directory is detected as a file
		exists, err := FileExists(tempDir)
		if err == nil {
			t.Errorf("Expected error for directory, but got none")
		}
		if exists {
			t.Errorf("Expected directory to not be treated as a file, but it was")
		}

		expectedErr := "path exists but is a directory"
		if err != nil && !ContainsSubstring(err.Error(), expectedErr) {
			t.Errorf("Unexpected error message. Got: %v, Want substring: %s", err, expectedErr)
		}
	})

	t.Run("Empty path", func(t *testing.T) {
		// Check for an empty path
		exists, err := FileExists("")
		if err == nil {
			t.Errorf("Expected error for empty path, but got none")
		}
		if exists {
			t.Errorf("Expected empty path to not exist, but it does")
		}

		expectedErr := "path is empty"
		if err != nil && !ContainsSubstring(err.Error(), expectedErr) {
			t.Errorf("Unexpected error message. Got: %v, Want substring: %s", err, expectedErr)
		}
	})
}

func TestDirExists(t *testing.T) {
	t.Run("Directory exists", func(t *testing.T) {
		tempDir := t.TempDir()

		// Check if the directory exists
		exists, err := DirExists(tempDir)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !exists {
			t.Errorf("Expected directory to exist, but it does not")
		}
	})

	t.Run("Directory does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		nonExistentDir := filepath.Join(tempDir, "non_existent_dir")

		// Check for a non-existent directory
		exists, err := DirExists(nonExistentDir)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if exists {
			t.Errorf("Expected directory to not exist, but it does")
		}
	})

	t.Run("Path is a file", func(t *testing.T) {
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test_file.txt")

		// Create a temporary file
		err := os.WriteFile(testFile, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Check if the file is incorrectly treated as a directory
		exists, err := DirExists(testFile)
		if err == nil {
			t.Errorf("Expected error for file, but got none")
		}
		if exists {
			t.Errorf("Expected file to not be treated as a directory, but it was")
		}

		expectedErr := "path exists but is not a directory"
		if err != nil && !ContainsSubstring(err.Error(), expectedErr) {
			t.Errorf("Unexpected error message. Got: %v, Want substring: %s", err, expectedErr)
		}
	})

	t.Run("Empty path", func(t *testing.T) {
		// Check for an empty path
		exists, err := DirExists("")
		if err == nil {
			t.Errorf("Expected error for empty path, but got none")
		}
		if exists {
			t.Errorf("Expected empty path to not exist, but it does")
		}

		expectedErr := "path is empty"
		if err != nil && !ContainsSubstring(err.Error(), expectedErr) {
			t.Errorf("Unexpected error message. Got: %v, Want substring: %s", err, expectedErr)
		}
	})
}

func TestEnsureDirExists(t *testing.T) {
	// Test paths
	tempDir := filepath.Join(os.TempDir(), "ensureDirTest")
	invalidPath := ""

	// Cleanup function
	defer os.RemoveAll(tempDir)

	t.Run("CreateNewDirectory", func(t *testing.T) {
		err := EnsureDirExists(tempDir)
		if err != nil {
			t.Errorf("EnsureDirExists(%q) returned error: %v; want nil", tempDir, err)
		}

		// Check if the directory exists
		info, err := os.Stat(tempDir)
		if err != nil || !info.IsDir() {
			t.Errorf("EnsureDirExists(%q) did not create a valid directory", tempDir)
		}
	})

	t.Run("ExistingDirectory", func(t *testing.T) {
		err := EnsureDirExists(os.TempDir())
		if err != nil {
			t.Errorf("EnsureDirExists(%q) returned error: %v; want nil", os.TempDir(), err)
		}
	})

	t.Run("InvalidPath", func(t *testing.T) {
		err := EnsureDirExists(invalidPath)
		if err == nil {
			t.Errorf("EnsureDirExists(%q) returned no error; want error", invalidPath)
		}
	})
}

func TestValidateFoldersExistence(t *testing.T) {
	tempDir := t.TempDir() // Create a temporary base directory

	// Create some valid directories
	existingDirs := []string{
		filepath.Join(tempDir, "dir1"),
		filepath.Join(tempDir, "dir2"),
	}

	for _, dir := range existingDirs {
		if err := os.Mkdir(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}

	tests := []struct {
		name         string
		folders      []string
		errorMessage string
		expectError  bool
	}{
		{
			name:         "All directories exist",
			folders:      existingDirs,
			errorMessage: "Missing folder",
			expectError:  false,
		},
		{
			name:         "One missing directory",
			folders:      append(existingDirs, filepath.Join(tempDir, "missingDir")),
			errorMessage: "Missing folder",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFoldersExistence(tt.folders, tt.errorMessage)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got: %v", err)
				}
			}
		})
	}
}

func TestRemoveIfExists(t *testing.T) {
	// Temporary paths
	tempFile := filepath.Join(os.TempDir(), "testRemoveFile.tmp")
	tempDir := filepath.Join(os.TempDir(), "testRemoveDir")

	// Setup: Create test file and directory
	file, err := os.Create(tempFile)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	file.Close()

	err = os.Mkdir(tempDir, 0755)
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// Test case: Remove file
	t.Run("RemoveFile", func(t *testing.T) {
		err := RemoveIfExists(tempFile)
		if err != nil {
			t.Errorf("RemoveIfExists(%q) returned error: %v; want nil", tempFile, err)
		}

		if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
			t.Errorf("RemoveIfExists(%q) did not remove the file", tempFile)
		}
	})

	// Test case: Remove directory
	t.Run("RemoveDirectory", func(t *testing.T) {
		err := RemoveIfExists(tempDir)
		if err != nil {
			t.Errorf("RemoveIfExists(%q) returned error: %v; want nil", tempDir, err)
		}

		if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
			t.Errorf("RemoveIfExists(%q) did not remove the directory", tempDir)
		}
	})

	// Test case: Path does not exist
	t.Run("PathDoesNotExist", func(t *testing.T) {
		nonExistentPath := filepath.Join(os.TempDir(), "nonexistentPath")
		err := RemoveIfExists(nonExistentPath)
		if err != nil {
			t.Errorf("RemoveIfExists(%q) returned error: %v; want nil", nonExistentPath, err)
		}
	})
}

func TestCheckMissingFolders(t *testing.T) {
	// Setup: Create temporary directories
	tempDir := t.TempDir()
	existingDir := filepath.Join(tempDir, "existing")
	if err := os.Mkdir(existingDir, 0755); err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	tests := []struct {
		name            string
		folders         map[string]string
		expectedMissing []string
	}{
		{
			name: "All Folders Exist",
			folders: map[string]string{
				"Existing Folder": existingDir,
			},
			expectedMissing: nil,
		},
		{
			name: "One Missing Folder",
			folders: map[string]string{
				"Existing Folder": existingDir,
				"Missing Folder":  filepath.Join(tempDir, "missing"),
			},
			expectedMissing: []string{"  - Missing Folder: " + filepath.Join(tempDir, "missing")},
		},
		{
			name: "All Folders Missing",
			folders: map[string]string{
				"Missing Folder 1": filepath.Join(tempDir, "missing1"),
				"Missing Folder 2": filepath.Join(tempDir, "missing2"),
			},
			expectedMissing: []string{
				"  - Missing Folder 1: " + filepath.Join(tempDir, "missing1"),
				"  - Missing Folder 2: " + filepath.Join(tempDir, "missing2"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			missingFolders, err := CheckMissingFolders(tt.folders)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(missingFolders) != len(tt.expectedMissing) {
				t.Errorf("Expected %d missing folders, got %d", len(tt.expectedMissing), len(missingFolders))
			}

			for i, missing := range missingFolders {
				if missing != tt.expectedMissing[i] {
					t.Errorf("Expected missing folder: %s, got: %s", tt.expectedMissing[i], missing)
				}
			}
		})
	}
}

func TestReadFileAsString(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")

	t.Run("ValidFile", func(t *testing.T) {
		content := "Hello, World!"
		if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		got, err := ReadFileAsString(tempFile)
		if err != nil || got != content {
			t.Errorf("ReadFileAsString(%q) = %q, %v; want %q, nil", tempFile, got, err, content)
		}
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		_, err := ReadFileAsString(filepath.Join(tempDir, "nonexistent.txt"))
		if err == nil {
			t.Errorf("Expected error for non-existent file, but got nil")
		}
	})
}

func TestWriteToFile(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		filePath    string
		content     []byte
		expectError bool
		setup       func() error
	}{
		{
			name:     "Write to a new file in a new directory",
			filePath: filepath.Join(tempDir, "newdir", "test.txt"),
			content:  []byte("Hello, World!"),
		},
		{
			name:     "Write to an existing file",
			filePath: filepath.Join(tempDir, "existing.txt"),
			content:  []byte("Existing file content"),
			setup: func() error {
				return os.WriteFile(filepath.Join(tempDir, "existing.txt"), []byte("Initial content"), 0644)
			},
		},
		{
			name:        "Fail when parent path is a file",
			filePath:    filepath.Join(tempDir, "afile", "test.txt"),
			content:     []byte("Invalid operation"),
			expectError: true,
			setup: func() error {
				return os.WriteFile(filepath.Join(tempDir, "afile"), []byte("I am a file"), 0644)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			err := WriteToFile(tt.filePath, tt.content)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				// Verify file content
				content, err := os.ReadFile(tt.filePath)
				if err != nil {
					t.Fatalf("Failed to read file: %v", err)
				}
				if string(content) != string(tt.content) {
					t.Errorf("File content mismatch:\nGot: %s\nWant: %s", content, tt.content)
				}
			}
		})
	}
}

func TestWriteStringToFile(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		filePath    string
		content     string
		expectError bool
		setup       func() error
	}{
		{
			name:     "Write string to a new file",
			filePath: filepath.Join(tempDir, "test.txt"),
			content:  "String content",
		},
		{
			name:     "Write string to an existing file",
			filePath: filepath.Join(tempDir, "existing.txt"),
			content:  "Updated string content",
			setup: func() error {
				return os.WriteFile(filepath.Join(tempDir, "existing.txt"), []byte("Old content"), 0644)
			},
		},
		{
			name:        "Fail when parent path is a file",
			filePath:    filepath.Join(tempDir, "afile", "test.txt"),
			content:     "Should fail",
			expectError: true,
			setup: func() error {
				return os.WriteFile(filepath.Join(tempDir, "afile"), []byte("I am a file"), 0644)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			err := WriteStringToFile(tt.filePath, tt.content)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				// Verify file content
				content, err := os.ReadFile(tt.filePath)
				if err != nil {
					t.Fatalf("Failed to read file: %v", err)
				}
				if string(content) != tt.content {
					t.Errorf("File content mismatch:\nGot: %s\nWant: %s", content, tt.content)
				}
			}
		})
	}
}

func TestWriteJSONToFile(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.json")

	data := map[string]string{"key": "value"}

	t.Run("WriteJSON", func(t *testing.T) {
		if err := WriteJSONToFile(tempFile, data); err != nil {
			t.Fatalf("Failed to write JSON to file: %v", err)
		}
		content, err := os.ReadFile(tempFile)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		expected := `{
  "key": "value"
}`
		if string(content) != expected {
			t.Errorf("File content mismatch: got %q, want %q", content, expected)
		}
	})
}

func TestRemoveTemplatingExtension(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "Single extension (non-templating)",
			filename: "example.txt",
			expected: "example.txt",
		},
		{
			name:     "Single templating extension",
			filename: "example.txt.gotxt",
			expected: "example.txt",
		},
		{
			name:     "Another templating extension",
			filename: "example.css.gotmpl",
			expected: "example.css",
		},
		{
			name:     "Multiple extensions (last one templating)",
			filename: "archive.tar.gotmpl",
			expected: "archive.tar",
		},
		{
			name:     "Multiple extensions (last one non-templating)",
			filename: "archive.tar.gz",
			expected: "archive.tar.gz",
		},
		{
			name:     "No extension",
			filename: "example",
			expected: "example",
		},
		{
			name:     "Hidden file with templating extension",
			filename: ".config.gotxt",
			expected: ".config",
		},
		{
			name:     "Hidden file without extension",
			filename: ".config",
			expected: ".config",
		},
		{
			name:     "Filename with dot but no extension",
			filename: "filename.",
			expected: "filename.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveTemplatingExtension(tt.filename, []string{".gotxt", ".gotmpl", ".tpl"})
			if result != tt.expected {
				t.Errorf("removeLastExtension(%q) = %q, want %q", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestToTemplFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"style.css", "style.templ"},
		{"app.js", "app.templ"},
		{"index.html", "index.templ"},
		{"config.json", "config.templ"},
		{"my-script.min.js", "my-script.min.templ"},       // Preserve `.min` in filename
		{"nested.file.name.ts", "nested.file.name.templ"}, // Multi-dot filenames
		{"no-extension", "no-extension.templ"},            // No extension case
		{".hiddenfile", ".templ"},                         // Edge case: dotfiles with no extension
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToTemplFilename(tt.input)
			if result != tt.expected {
				t.Errorf("ToTemplFilename(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFileNameWithoutExt(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"style.css", "style"},
		{"app.js", "app"},
		{"index.html", "index"},
		{"config.json", "config"},
		{"my-script.min.js", "my-script.min"},       // Preserve `.min`
		{"nested.file.name.ts", "nested.file.name"}, // Multi-dot filenames
		{"no-extension", "no-extension"},            // No extension case
		{".hiddenfile", ""},                         // Edge case: dotfiles with no extension
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := fileNameWithoutExt(tt.input)
			if result != tt.expected {
				t.Errorf("fileNameWithoutExt(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRebasePathToOutput(t *testing.T) {
	tests := []struct {
		inputPath string
		inputDir  string
		outputDir string
		expected  string
	}{
		{"assets/button/css/base.css", "assets", "components", "components/button/css/base.templ"},
		{"./assets/button/css/base.css", "assets", "components", "components/button/css/base.templ"},
		{"src/component/styles.css", "src", "components", "components/component/styles.templ"},
		{"somefolder/file.css", "somefolder", "components", "components/file.templ"},
		{"./somefolder/file.css", "somefolder", "components", "components/file.templ"},
		{"file.css", "somefolder", "components", "components/file.templ"}, // Edge case: inputPath with no directory
		{"./file.css", "somefolder", "components", "components/file.templ"},
		{"assets/file", "assets", "components", "components/file.templ"}, // Case with no extension
		{"./assets/file", "assets", "components", "components/file.templ"},
		{"deeply/nested/path/style.css", "deeply", "output", "output/nested/path/style.templ"}, // Deeply nested case
	}

	for _, tt := range tests {
		t.Run(tt.inputPath, func(t *testing.T) {
			result := RebasePathToOutput(tt.inputPath, tt.inputDir, tt.outputDir)
			if result != tt.expected {
				t.Errorf("rebasePathToOutput(%q, %q) = %q; want %q", tt.inputPath, tt.outputDir, result, tt.expected)
			}
		})
	}
}
