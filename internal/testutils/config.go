package testutils

import (
	"path/filepath"

	"github.com/indaco/tempo/internal/config"
)

func SetupConfig(tempDir string, overrides func(cfg *config.Config)) *config.Config {
	cfg := config.DefaultConfig()
	cfg.TempoRoot = filepath.Join(tempDir, ".tempo-files")
	cfg.App.GoPackage = filepath.Join(tempDir, "custom-package")
	cfg.App.AssetsDir = filepath.Join(tempDir, "custom-assets")
	cfg.Paths.TemplatesDir = filepath.Join(cfg.TempoRoot, "templates")
	cfg.Paths.ActionsDir = filepath.Join(cfg.TempoRoot, "actions")

	if overrides != nil {
		overrides(cfg)
	}
	return cfg
}

func PrepareTestConfig(cfg *config.Config, tempDir string) {
	cfg.TempoRoot = filepath.Join(tempDir, ".tempo-files")
	cfg.Paths.TemplatesDir = filepath.Join(cfg.TempoRoot, "templates")
	cfg.Paths.ActionsDir = filepath.Join(cfg.TempoRoot, "actions")
}
