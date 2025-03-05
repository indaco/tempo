package app

import (
	"path/filepath"

	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/errors"
	"github.com/indaco/tempo/internal/utils"
)

// IsTempoProject checks if any of the prioritized Tempo config files exist.
func IsTempoProject(pathToFile string) error {
	for _, file := range config.TempoConfigFiles {
		exists, isDir, err := utils.FileOrDirExistsFunc(filepath.Join(pathToFile, file))
		if err != nil {
			return errors.Wrap("error checking config file '%s'", err, file)
		}
		if exists && !isDir {
			return nil // File exists and is not a directory
		}
	}

	return errors.Wrap("no config file found; checked: %v. Run 'tempo init' first", config.TempoConfigFiles)
}
