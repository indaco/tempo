// GoPackage generator provides functionality for processing templating actions,
// including adding single files or multiple files to a destination directory.
package generator

import (
	"github.com/indaco/tempo/internal/errors"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/utils"
)

// Map of action types to their respective handlers.
var actionHandlers = map[string]ActionHandler{
	CopyActionId:   &CopyAction{},
	RenderActionId: &RenderAction{},
}

// ProcessActions processes a list of actions using the appropriate handlers.
func ProcessActions(logger logger.LoggerInterface, actions []Action, data *TemplateData) error {
	for _, action := range actions {
		if data.DryRun {
			handleDryRun(logger, action, data)
			continue
		}

		handler, exists := actionHandlers[action.Type]
		if !exists {
			return errors.Wrap("unknown action type", action.Type)
		}

		if err := handler.Execute(action, data); err != nil {
			return errors.Wrap("error executing action", err, action.Type)
		}
	}
	return nil
}

// handleDryRun handles the dry-run mode by resolving templates and logging the actions.
func handleDryRun(logger logger.LoggerInterface, action Action, data *TemplateData) {
	switch action.Item {
	case "file":
		// Handle single file addition
		resolvedPath, _ := utils.RenderTemplate(action.Path, data)
		resolvedTemplate, _ := utils.RenderTemplate(action.TemplateFile, data)
		logger.Info("Dry Run: Would execute action:", action.Item, " with template: ", resolvedTemplate, " to path ", resolvedPath)
	case "folder":
		// Handle multiple file additions
		resolvedBase, _ := utils.RenderTemplate(action.Source, data)
		resolvedDestination, _ := utils.RenderTemplate(action.Destination, data)
		logger.Info("Dry Run: Would execute action", action.Item, "from base", resolvedBase, "to destination", resolvedDestination)
	default:
		logger.Warning("Dry Run: Unknown action type", action.Item)
	}
}
