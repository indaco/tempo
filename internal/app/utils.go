package app

import (
	"os"
	"path/filepath"

	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/errors"
	"github.com/indaco/tempo/internal/utils"
)

// IsTempoProject checks if any of the prioritized Tempo config files exist.
func IsTempoProject(workingDir string) error {
	if err := isGolangProject(workingDir); err != nil {
		return err
	}

	for _, file := range config.TempoConfigFiles {
		exists, isDir, err := utils.FileOrDirExistsFunc(filepath.Join(workingDir, file))
		if err != nil {
			return errors.Wrap("error checking config file '%s'", err, file)
		}
		if exists && !isDir {
			return nil // File exists and is not a directory
		}
	}

	return errors.Wrap("no config file found; checked: %v. Run 'tempo init' first", config.TempoConfigFiles)
}

// isGolangProject checks if the working dir is a valid Golang project .
func isGolangProject(workingDir string) error {
	goModPath := filepath.Join(workingDir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return errors.Wrap("missing go.mod file. Run 'go mod init' to create one")
	}
	return nil
}
