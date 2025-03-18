package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"slices"

	"github.com/indaco/tempo/internal/errors"
	"golang.org/x/mod/modfile"
)

/* ------------------------------------------------------------------------- */
/* FILE AND DIRECTORY UTILITIES                                              */
/* ------------------------------------------------------------------------- */

// GetCWD returns the current working directory.
// It logs an error and exits the program if the directory cannot be retrieved.
func GetCWD() string {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current working directory: %v\n", err)
		os.Exit(1) // Exit the program in case of an error
	}
	return cwd
}

// FileOrDirExistsFunc is a function variable to allow testing overrides.
var FileOrDirExistsFunc = FileOrDirExists

// FileOrDirExists checks whether a file or directory exists at the specified path.
func FileOrDirExists(path string) (exists bool, isDir bool, err error) {
	if path == "" {
		return false, false, errors.Wrap("path is empty")
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, false, nil
	}
	if err != nil {
		return false, false, errors.Wrap("error checking path", err)
	}
	return true, info.IsDir(), nil
}

// FileExistsFunc is a function variable to allow testing overrides.
var FileExistsFunc = FileExists

// FileExists checks if a file exists at the specified path.
// It wraps FileOrDirExists and ensures the path points to a file, not a directory.
func FileExists(path string) (bool, error) {
	exists, isDir, err := FileOrDirExists(path)
	if err != nil {
		return false, err
	}
	if isDir {
		return false, errors.Wrap("path exists but is a directory: '%s'", nil, path)
	}
	return exists, nil
}

// DirExists checks if a directory exists at the specified path.
// It wraps FileOrDirExists and ensures the path points to a directory, not a file.
func DirExists(path string) (bool, error) {
	exists, isDir, err := FileOrDirExists(path)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}
	if !isDir {
		return false, errors.Wrap("path exists but is not a directory", path)
	}
	return true, nil
}

// EnsureDirExists ensures that a directory exists or creates it if it does not.
func EnsureDirExists(dir string) error {
	stat, err := os.Stat(dir)
	if err == nil {
		if !stat.IsDir() {
			return errors.Wrap("path '%s' exists but is not a directory", dir)
		}
		return nil // Directory already exists
	}
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrap("failed to create directory '%s'", err, dir)
		}
		return nil
	}
	return errors.Wrap("error checking directory '%s'", err, dir)
}

// ValidateFoldersExistence checks if the specified folders exist and returns an error if any of them are missing.
func ValidateFoldersExistence(folders []string, errorMessage string) error {
	for _, folder := range folders {
		exists, err := DirExists(folder)
		if err != nil {
			return errors.Wrap("Failed to check folder existence", err)
		}
		if !exists {
			return errors.Wrap(errorMessage)
		}
	}
	return nil
}

// RemoveIfExists removes a file or directory if it exists.
func RemoveIfExists(path string) error {
	// Check if the file or directory exists
	exists, _, err := FileOrDirExists(path)
	if err != nil {
		return errors.Wrap("failed to check if path '%s' exists", err, path)
	}

	// If it exists, remove it
	if exists {
		if err := os.RemoveAll(path); err != nil {
			return errors.Wrap("failed to remove path '%s'", err, path)
		}
	}
	return nil
}

var CheckMissingFoldersFunc = CheckMissingFolders

func CheckMissingFolders(folders map[string]string) (map[string]string, error) {
	missingFolders := make(map[string]string)

	for name, path := range folders {
		if exists, err := DirExists(path); err != nil || !exists {
			missingFolders[name] = path
		}
	}

	return missingFolders, nil
}

/* ------------------------------------------------------------------------- */
/* FILE CONTENT READ/WRITE                                                   */
/* ------------------------------------------------------------------------- */

// ReadFileAsString reads a file and returns its content as a string.
func ReadFileAsString(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", errors.Wrap("failed to read file '%s'", err, filePath)
	}
	return string(content), nil
}

// WriteToFile writes byte content to the specified file, ensuring directories exist.
func WriteToFile(path string, content []byte) error {
	if err := EnsureDirExists(filepath.Dir(path)); err != nil {
		return err
	}

	return os.WriteFile(path, content, 0644)
}

// WriteStringToFile writes string content to the specified file.
func WriteStringToFile(path string, content string) error {
	return WriteToFile(path, []byte(content))
}

// WriteJSONToFile writes the given data as JSON to the specified file.
func WriteJSONToFile(filePath string, data any) error {
	// Marshal data into JSON with proper indentation
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errors.Wrap("failed to marshal data to JSON", err)
	}

	// Write JSON to file
	return WriteToFile(filePath, jsonData)
}

/* ------------------------------------------------------------------------- */
/* TEMPLATING UTILITIES                                                      */
/* ------------------------------------------------------------------------- */

// RemoveTemplatingExtension removes specific templating extensions (e.g., ".gotxt", ".gotmpl")
// while retaining the rest of the filename.
func RemoveTemplatingExtension(filename string, extensions []string) string {
	ext := filepath.Ext(filename)
	if slices.Contains(extensions, ext) {
		return strings.TrimSuffix(filename, ext)
	}
	return filename
}

// ToTemplFilename converts a filename to its `.templ` equivalent by removing its extension.
func ToTemplFilename(fileName string) string {
	return fileNameWithoutExt(fileName) + ".templ"
}

// fileNameWithoutExt returns the filename without its extension.
func fileNameWithoutExt(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}

// RebasePathToOutput transforms an input file path by replacing its root directory (`inputDir`)
// with `outputDir` and ensuring the file extension is `.templ`.
//
// The function performs the following steps:
// 1. Cleans and normalizes the input path by removing redundant slashes and "./" prefixes.
// 2. Ensures the input path starts with `inputDir` and removes it.
// 3. Constructs the new path within `outputDir` while preserving the relative structure.
// 4. Converts the file extension to `.templ`.
//
// Parameters:
//   - inputPath: The original file path to be transformed.
//   - inputDir: The root directory to be replaced in `inputPath`.
//   - outputDir: The target directory where the transformed file should be placed.
//
// Returns:
//   - A new file path within `outputDir`, using a `.templ` extension.
func RebasePathToOutput(inputPath, inputDir, outputDir string) string {
	// Normalize paths (removes redundant slashes and "./" prefixes)
	inputPath = path.Clean(inputPath)
	inputDir = path.Clean(inputDir)
	outputDir = path.Clean(outputDir)

	// Ensure inputPath starts with inputDir
	if !strings.HasPrefix(inputPath, inputDir+"/") {
		// If inputPath is not inside inputDir, just place it inside outputDir
		return ToTemplFilename(path.Join(outputDir, path.Base(inputPath)))
	}

	// Remove inputDir prefix and trim any leading slash
	relativePath := strings.TrimPrefix(inputPath, inputDir+"/")

	// Construct the new path inside outputDir
	newPath := path.Join(outputDir, relativePath)

	// Convert to `.templ` extension
	return ToTemplFilename(newPath)
}

// GetModuleName extracts the module name from the go.mod file.
func GetModuleName(goModPath string) (string, error) {
	goModFile := filepath.Join(goModPath, "go.mod")
	content, err := os.ReadFile(goModFile)
	if err != nil {
		return "", errors.Wrap("error reading go.mod file", err)
	}

	parsedModFile, err := modfile.Parse(goModFile, content, nil)
	if err != nil {
		return "", errors.Wrap("error parsing go.mod file", err)
	}

	if parsedModFile.Module == nil || parsedModFile.Module.Mod.Path == "" {
		return "", errors.Wrap("module path not found in go.mod file")
	}

	return parsedModFile.Module.Mod.Path, nil
}
