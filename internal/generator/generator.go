// GoPackage generator provides functionality for processing templating actions,
// including adding single files or multiple files to a destination directory.
package generator

import (
	"context"

	"github.com/indaco/tempo/internal/apperrors"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/utils"
)

// ActionProcessor defines the interface for processing actions.
// This allows for dependency injection and easier testing.
type ActionProcessor interface {
	ProcessActions(ctx context.Context, logger logger.Logger, actions []Action, data *TemplateData) error
}

// DefaultActionProcessor is the default implementation of ActionProcessor.
type DefaultActionProcessor struct{}

// NewActionProcessor creates a new ActionProcessor instance.
func NewActionProcessor() ActionProcessor {
	return &DefaultActionProcessor{}
}

// ProcessActions implements ActionProcessor.ProcessActions.
func (p *DefaultActionProcessor) ProcessActions(ctx context.Context, logger logger.Logger, actions []Action, data *TemplateData) error {
	return ProcessActions(ctx, logger, actions, data)
}

// Ensure DefaultActionProcessor implements ActionProcessor
var _ ActionProcessor = (*DefaultActionProcessor)(nil)

// Map of action types to their respective handlers.
var actionHandlers = map[string]ActionHandler{
	CopyActionID:   &CopyAction{},
	RenderActionID: &RenderAction{},
}

// Default function implementation (kept for backward compatibility)
//
// Deprecated: Use ActionProcessor interface instead for new code.
var ProcessActionsFunc = ProcessActions

// ProcessActions processes a list of actions using the appropriate handlers.
func ProcessActions(ctx context.Context, logger logger.Logger, actions []Action, data *TemplateData) error {
	// Validate context - use Background as fallback for non-critical operations
	if ctx == nil {
		ctx = context.Background()
	}

	for _, action := range actions {
		if data.DryRun {
			handleDryRun(logger, action, data)
			continue
		}

		// Skip JS-related actions unless OnlyIfJs/WithJs is true
		if isJsAction(action) && !data.WithJs {
			continue
		}

		handler, exists := actionHandlers[action.Type]
		if !exists {
			return apperrors.Wrap("unknown action type", action.Type)
		}

		if err := handler.Execute(ctx, action, data); err != nil {
			return apperrors.Wrap("error executing action", err, action.Type)
		}
	}
	return nil
}

// handleDryRun handles the dry-run mode by resolving templates and logging the actions.
func handleDryRun(logger logger.Logger, action Action, data *TemplateData) {
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
