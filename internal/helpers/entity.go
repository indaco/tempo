package helpers

import (
	"fmt"
	"path/filepath"

	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/templatefuncs/providers/textprovider"
)

// CheckEntityForNew logs a warning or info message when creating a new entity that already exists.
// It handles both component and variant entity types with appropriate path formatting.
func CheckEntityForNew(entityType, entityName, outputPath string, force bool, logr logger.LoggerInterface) {
	// Select logging function and action message based on `force` flag
	logFunc, action := logr.Warning, "Use '--force' to overwrite it. Any changes will be lost."
	if force {
		logFunc, action = logr.Info, "Overwriting due to '--force' flag."
	}

	// Determine the appropriate path format based on entity type
	paths := map[string]string{
		"component": filepath.Join(outputPath, entityName),
		"variant":   outputPath,
	}
	path, exists := paths[entityType]
	if !exists {
		path = outputPath
	}

	msg := fmt.Sprintf("%s '%s' already exists.\n  %s\n", textprovider.TitleCase(entityType), entityName, action)
	logFunc(msg).WithAttrs("path", path)
}

// CheckEntityForDefine logs a warning or info message when defining templates that already exist.
func CheckEntityForDefine(entityType, outputPath string, force bool, logr logger.LoggerInterface) {
	// Select logging function and action message based on `force` flag
	logFunc, action := logr.Warning, "Use '--force' to overwrite them. Any changes will be lost."
	if force {
		logFunc, action = logr.Info, "Overwriting due to '--force' flag."
	}

	msg := fmt.Sprintf("Templates for '%s' already exist.\n  %s", entityType, action)
	logFunc(msg).WithAttrs("path", outputPath)
}
