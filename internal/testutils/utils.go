package testutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/indaco/tempo/internal/config"
	"gopkg.in/yaml.v3"
)

// WriteConfigToFile saves the configuration to the specified file path in YAML format.
func WriteConfigToFile(filePath string, cfg *config.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// ValidateGeneratedFiles ensures the specified files exist.
func ValidateGeneratedFiles(t *testing.T, paths []string) {
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file not found: %s", path)
		}
	}
}

func CreateModFile(tempDir string) error {
	// Create go.mod inside tempDir (the correct working directory)
	goModPath := filepath.Join(tempDir, "go.mod")
	err := os.WriteFile(goModPath, []byte("module example.com/myproject\n\ngo 1.23\n"), 0644)
	if err != nil {
		return err
	}
	return nil
}
