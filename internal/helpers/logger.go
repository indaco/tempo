package helpers

import (
	"path/filepath"

	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/logger"
)

// EnableLoggerIndentation ensures consistent indentation.
func EnableLoggerIndentation(log logger.LoggerInterface) {
	log.WithIndent(true)
}

func ResetLogger(log logger.LoggerInterface) {
	log.Reset()
}

// LogSuccessMessages logs the success messages for generated template files and assets.
func LogSuccessMessages(entityType string, cfg *config.Config, logger logger.LoggerInterface) {
	var message string

	basePath := filepath.Join(cfg.Paths.TemplatesDir, entityType, "templ")
	assetsPath := filepath.Join(cfg.Paths.TemplatesDir, entityType, "assets")

	switch entityType {
	case "component":
		message = "Templates for the component and assets (CSS and JS) have been created"
	case "component-variant":
		message = "Templates for the component variant and assets (CSS) have been created"
	default:
		message = "Templates and assets have been created"
	}

	logger.Success(message).
		WithAttrs(
			"templates_path", basePath,
			"assets_path", assetsPath,
		)
}
