package utils

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/indaco/tempo/internal/errors"
	"github.com/indaco/tempo/internal/templates"
)

var embeddedFiles = templates.EmbeddedFiles

// IsEmbeddedFunc is a function variable to allow testing overrides.
var IsEmbeddedFunc = IsEmbedded

// IsEmbedded checks if a file or directory is embedded in the embedded filesystem.
func IsEmbedded(path string) bool {
	_, err := embeddedFiles.Open(path)
	return err == nil
}

// ReadEmbeddedFile reads a file from embedded resources and returns its content as a byte slice.
func ReadEmbeddedFile(path string) ([]byte, error) {
	file, err := embeddedFiles.Open(path)
	if err != nil {
		return nil, errors.Wrap("failed to open embedded file '%s'", err, path)
	}
	defer file.Close()

	return io.ReadAll(file)
}

// ReadEmbeddedDir reads the entries of a directory from embedded resources.
func ReadEmbeddedDir(path string) ([]os.DirEntry, error) {
	entries, err := fs.ReadDir(embeddedFiles, path)
	if err != nil {
		return nil, errors.Wrap("failed to read embedded directory '%s'", err, path)
	}
	return entries, nil
}

// ensureDirExists ensures that the specified directory exists, creating it if necessary.
func ensureDirExists(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.Wrap("failed to create directory '%s'", err, dir)
	}
	return nil
}

// CopyDirFromEmbedFunc is a function variable to allow testing overrides.
var CopyDirFromEmbedFunc = CopyDirFromEmbed

// CopyDirFromEmbed copies all files and subdirectories from an embedded source to a destination directory.
func CopyDirFromEmbed(source, destination string) error {
	// Ensure the destination directory exists
	if err := ensureDirExists(destination); err != nil {
		return err
	}

	// Read the source directory entries
	entries, err := ReadEmbeddedDir(source)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(source, entry.Name())
		destPath := filepath.Join(destination, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectories
			if err := CopyDirFromEmbed(srcPath, destPath); err != nil {
				return err
			}
		} else {
			// Copy individual files
			if err := CopyFileFromEmbed(srcPath, destPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// CopyFileFromEmbedFunc is a function variable to allow testing overrides.
var CopyFileFromEmbedFunc = CopyFileFromEmbed

// CopyFileFromEmbed copies a single file from an embedded source to a destination file path.
func CopyFileFromEmbed(source, destination string) error {
	// Ensure the destination directory exists
	if err := ensureDirExists(filepath.Dir(destination)); err != nil {
		return err
	}

	// Read the embedded file content
	content, err := ReadEmbeddedFile(source)
	if err != nil {
		return err
	}

	// Write the content to the destination file
	if err := os.WriteFile(destination, content, 0644); err != nil {
		return errors.Wrap("failed to write to file '%s': %w", err, destination)
	}

	return nil
}
