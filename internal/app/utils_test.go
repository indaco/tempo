package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/indaco/tempo/internal/utils"
)

func TestIsTempoProject(t *testing.T) {
	t.Run("Valid tempo.yaml file exists", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "tempo.yaml")
		if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test config file: %v", err)
		}

		err := IsTempoProject(tempDir)
		if err != nil {
			t.Errorf("expected nil, got error: %v", err)
		}
	})

	t.Run("Valid tempo.yml file exists", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "tempo.yml")
		if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create test config file: %v", err)
		}

		err := IsTempoProject(tempDir)
		if err != nil {
			t.Errorf("expected nil, got error: %v", err)
		}
	})

	t.Run("No config file exists", func(t *testing.T) {
		tempDir := t.TempDir()
		err := IsTempoProject(tempDir)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("Error when checking for file existence", func(t *testing.T) {
		oldFileOrDirExists := utils.FileOrDirExists
		defer func() { utils.FileOrDirExistsFunc = oldFileOrDirExists }()

		utils.FileOrDirExistsFunc = func(path string) (bool, bool, error) {
			return false, false, errors.New("mocked error")
		}

		tempDir := t.TempDir()
		err := IsTempoProject(tempDir)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}
