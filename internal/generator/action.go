package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/errors"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/utils"
)

/* ------------------------------------------------------------------------- */
/* INTERFACES                                                                */
/* ------------------------------------------------------------------------- */

// ActionHandler defines the interface for processing actions.
type ActionHandler interface {
	Execute(action Action, data *TemplateData) error
}

/* ------------------------------------------------------------------------- */
/* TYPES & CONSTANTS                                                         */
/* ------------------------------------------------------------------------- */

// Action represents a templating action with configurable properties.
// It is used internally to process actions based on type (copy/render).
type Action struct {
	Type         string `json:"type,omitempty"`         // "copy" or "render"
	Item         string `json:"item,omitempty"`         // "file" or "folder"
	Path         string `json:"path,omitempty"`         // Output path (for "file")
	TemplateFile string `json:"templateFile,omitempty"` // Template file path (for "file")
	Source       string `json:"source,omitempty"`       // Base directory (for "folder")
	Destination  string `json:"destination,omitempty"`  // Destination directory (for "folder")
	OnlyIfJs     bool   `json:"onlyIfJs,omitempty"`     // Include only if --js is true
	SkipIfExists bool   `json:"skipIfExists,omitempty"` // Skips a file if it already exists
	Force        bool   `json:"force,omitempty"`        // Overwrites files if they exist
}

// ActionList represents a collection of Action objects.
type ActionList []Action

// JSONAction represents a templating action without the `Type` field.
type JSONAction struct {
	Item         string `json:"item"`
	TemplateFile string `json:"templateFile,omitempty"`
	Path         string `json:"path,omitempty"`
	Source       string `json:"source,omitempty"`
	Destination  string `json:"destination,omitempty"`
	OnlyIfJs     bool   `json:"onlyIfJs,omitempty"`     // Include only if --js is true
	SkipIfExists bool   `json:"skipIfExists,omitempty"` // Skips a file if it already exists
	Force        bool   `json:"force,omitempty"`        // Overwrites files if they exist
}

// JSONActionList represents a collection of JSONAction objects.
type JSONActionList []JSONAction

const (
	RenderActionId = "render"
	CopyActionId   = "copy"
)

/* ------------------------------------------------------------------------- */
/* CONVERSION METHODS                                                        */
/* ------------------------------------------------------------------------- */

// ToJSONAction converts an Action to JSONAction.
func (a *Action) ToJSONAction() JSONAction {
	return JSONAction{
		Item:         a.Item,
		TemplateFile: a.TemplateFile,
		Path:         a.Path,
		Source:       a.Source,
		Destination:  a.Destination,
		OnlyIfJs:     a.OnlyIfJs,
	}
}

// ToJSONAction converts an ActionList to a slice of JSONAction.
func (al ActionList) ToJSONAction() []JSONAction {
	jsonActions := make([]JSONAction, len(al))
	for i, action := range al {
		jsonActions[i] = action.ToJSONAction()
	}
	return jsonActions
}

// ToAction converts a JSONAction to an Action with a specified type.
func (jsa *JSONAction) ToAction(actionType string) Action {
	return Action{
		Type:         actionType,
		Item:         jsa.Item,
		TemplateFile: jsa.TemplateFile,
		Path:         jsa.Path,
		Source:       jsa.Source,
		Destination:  jsa.Destination,
		OnlyIfJs:     jsa.OnlyIfJs,
	}
}

// ToActions converts a JSONActionList to a slice of Action with a specified type.
func (jsaList JSONActionList) ToActions(actionType string) []Action {
	actions := make([]Action, len(jsaList))
	for i, jsa := range jsaList {
		actions[i] = jsa.ToAction(actionType)
	}
	return actions
}

/* ------------------------------------------------------------------------- */
/* ACTION HANDLERS                                                           */
/* ------------------------------------------------------------------------- */

// CopyAction handles copying files and folders.
type CopyAction struct{}

func (a *CopyAction) Execute(action Action, data *TemplateData) error {
	switch action.Item {
	case "file":
		destination := filepath.Join(data.TemplatesDir, action.TemplateFile)
		return utils.CopyFileFromEmbedFunc(action.TemplateFile, destination)
	case "folder":
		destinationPath := filepath.Join(data.TemplatesDir, action.Source)
		return utils.CopyDirFromEmbedFunc(action.Source, destinationPath)
	default:
		return errors.Wrap("unknown item type: %s", action.Item)
	}
}

// RenderAction handles rendering templates into files or folders.
type RenderAction struct{}

func (a *RenderAction) Execute(action Action, data *TemplateData) error {
	switch action.Item {
	case "file":
		return renderActionFile(action, data)
	case "folder":
		return renderActionFolder(action, data)
	default:
		return errors.Wrap("unknown item type: %s", action.Item)
	}
}

/* ------------------------------------------------------------------------- */
/* ACTION HELPERS                                                            */
/* ------------------------------------------------------------------------- */

func renderActionFile(action Action, data *TemplateData) error {
	// Skip if this action is JS-specific but the --js flag is not set
	if action.OnlyIfJs && !data.WithJs {
		return nil
	}
	// Step 1: Read and render the template file content
	filePath := filepath.Join(data.TemplatesDir, action.TemplateFile)
	renderedContent, err := readAndRenderTemplate(filePath, data)
	if err != nil {
		return errors.Wrap("failed to process template file", err, action.TemplateFile)
	}

	// Step 2: Render the output path
	outputPath, err := utils.RenderTemplate(action.Path, data)
	if err != nil {
		return errors.Wrap("failed to render output path", err, action.Path)
	}

	// Step 3: Handle output file existence and writing
	return handleOutputFile(outputPath, renderedContent, action, utils.FileOrDirExists, utils.WriteStringToFile)
}

func renderActionFolder(action Action, data *TemplateData) error {
	// Step 1: Render base and destination directories
	base, destination, err := renderBaseAndDestination(action, data)
	if err != nil {
		return err
	}

	// Step 2: Ensure the destination directory exists
	if err := os.MkdirAll(destination, 0755); err != nil {
		return errors.Wrap("failed to create destination directory", err, destination)
	}

	// Step 3: Read files from the base directory
	files, err := readDir(base)
	if err != nil {
		return errors.Wrap("failed to read base directory", err, base)
	}

	// Step 4: Process each file in the base directory
	for _, entry := range files {
		fileInfo, err := entry.Info() // Convert to os.FileInfo
		if err != nil {
			return errors.Wrap("failed to get file info", err, entry.Name())
		}
		if err := processFileInActionFolder(fileInfo, base, destination, action, data); err != nil {
			return err
		}
	}

	return nil
}

/* ------------------------------------------------------------------------- */
/* UTILITY FUNCTIONS                                                         */
/* ------------------------------------------------------------------------- */

// LoadUserActionsFunc is a function variable to allow testing overrides.
var LoadUserActionsFunc = LoadUserActions

// LoadUserActions reads an actions.json file and parses it into a slice of JSONAction structs.
func LoadUserActions(filePath string) ([]JSONAction, error) {
	// Check if the file exists and is not a directory
	exists, isDir, err := utils.FileOrDirExists(filePath)
	if err != nil {
		return nil, errors.Wrap("error checking actions file '%s': %w", err, filePath)
	}
	if !exists || isDir {
		return nil, errors.Wrap("actions file '%s' does not exist or is a directory", filePath)
	}

	// Read the file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap("failed to read actions file '%s': %w", err, filePath)
	}

	// Unmarshal the JSON content into a slice of JSONAction
	var actions []JSONAction
	if err := json.Unmarshal(data, &actions); err != nil {
		return nil, errors.Wrap("failed to parse actions file '%s': %w", err, filePath)
	}

	return actions, nil
}

// GenerateActionJSONFile marshals a slice of Actions into a JSON file.
func GenerateActionJSONFile(filePath string, actions ActionList) error {
	return utils.WriteJSONToFile(filePath, actions.ToJSONAction())
}

/* ------------------------------------------------------------------------- */
/* HELPER FUNCTIONS                                                          */
/* ------------------------------------------------------------------------- */

// GenerateActionFile generates an action file for the given entity type.
func GenerateActionFile(entityType string, data *TemplateData, actions []Action, logger logger.LoggerInterface) error {
	actionFileName := fmt.Sprintf("%s.json", entityType)
	actionsPath := filepath.Join(data.ActionsDir, actionFileName)

	if err := GenerateActionJSONFile(actionsPath, actions); err != nil {
		return errors.Wrap("Failed to generate action file for %s", entityType, err)
	}

	logger.Success(fmt.Sprintf("Tempo action file for '%s' has been created", entityType)).
		WithAttrs("action_file_path", actionsPath)
	return nil
}

// RetrieveActionsFile retrieves actions from a JSON file.
func RetrieveActionsFile(logger logger.LoggerInterface, actionFilePath string, cfg *config.Config) (JSONActionList, error) {
	// Step 1: Resolve action file path
	resolvedPath, err := resolveActionFilePath(cfg.Paths.ActionsDir, actionFilePath)
	if err != nil {
		return nil, errors.Wrap("failed to resolve action file path", err)
	}

	// Step 2: Load user-defined actions
	userActions, err := LoadUserActions(resolvedPath)
	if err != nil {
		return nil, errors.Wrap("failed to load user-defined actions from:", err, resolvedPath)
	}

	logger.Info("Actions loaded").
		WithAttrs(
			"action_file", resolvedPath,
			"num_actions", len(userActions),
		)

	return userActions, nil
}

// ProcessEntityActions retrieves and processes actions from a JSON file.
func ProcessEntityActions(logger logger.LoggerInterface, pathToActionsFile string, data *TemplateData, cfg *config.Config) error {
	// Retrieve user actions
	userActions, err := RetrieveActionsFile(logger, pathToActionsFile, cfg)
	if err != nil {
		return errors.Wrap("failed to get component actions file", err)
	}

	// Convert to built-in actions
	builtInActions := userActions.ToActions(RenderActionId)

	if data.Force {
		for i := range builtInActions {
			builtInActions[i].Force = true
		}
	}

	// Process actions
	if err := ProcessActionsFunc(logger, builtInActions, data); err != nil {
		return errors.Wrap("failed to process actions", err)
	}

	return nil
}

// resolveActionFilePath resolves the path to an action file.
func resolveActionFilePath(ActionsDir, actionFileFlag string) (string, error) {
	// Step 1: Resolve the action file path relative to the actions folder, if provided
	if ActionsDir != "" {
		resolvedPath := filepath.Join(ActionsDir, actionFileFlag)
		exists, err := utils.FileExistsFunc(resolvedPath)
		if err != nil {
			return "", err
		} else if exists {
			return resolvedPath, nil
		}
	}

	// Step 2: Check if the provided actionFileFlag is a valid full path
	// Check the actionFileFlag as an absolute path
	exists, err := utils.FileExistsFunc(actionFileFlag)
	if err != nil {
		return "", errors.Wrap("error checking action file path", err, actionFileFlag)
	}
	if !exists {
		return "", errors.Wrap("action file does not exist", actionFileFlag)
	}

	return actionFileFlag, nil
}

// Helper to read files from embedded or disk.
func readFile(path string) ([]byte, error) {
	// Check if the file exists on disk
	if _, err := os.Stat(path); err == nil {
		return os.ReadFile(path)
	}

	return nil, errors.Wrap("file not found in both embedded and disk: %s", path)
}

// Helper to read directories from embedded or disk.
func readDir(path string) ([]os.DirEntry, error) {
	if utils.IsEmbeddedFunc(path) {
		return utils.ReadEmbeddedDir(path)
	}
	return os.ReadDir(path) // Fallback to normal directory reading
}

// readAndRenderTemplate reads a file and renders its content.
func readAndRenderTemplate(filePath string, data *TemplateData) (string, error) {
	content, err := readFile(filePath)
	if err != nil {
		return "", errors.Wrap("failed to read file", err, filePath)
	}

	renderedContent, err := utils.RenderTemplate(string(content), data)
	if err != nil {
		return "", errors.Wrap("failed to render template", err, filePath)
	}

	return renderedContent, nil
}

// renderBaseAndDestination resolves and renders the base and destination directories.
func renderBaseAndDestination(action Action, data *TemplateData) (string, string, error) {
	base, err := utils.RenderTemplate(filepath.Join(data.TemplatesDir, action.Source), data)
	if err != nil {
		return "", "", errors.Wrap("failed to render base directory", err, action.Source)
	}

	destination, err := utils.RenderTemplate(action.Destination, data)
	if err != nil {
		return "", "", errors.Wrap("failed to render destination directory", err, action.Destination)
	}

	return base, destination, nil
}

// processFileInActionFolder processes a single file inside the action folder.
func processFileInActionFolder(file os.FileInfo, base, destination string, action Action, data *TemplateData) error {
	// Skip directories and specific system files
	if file.IsDir() || file.Name() == ".DS_Store" {
		return nil
	}

	originalFilename := file.Name()
	transformedFilename := utils.RemoveTemplatingExtension(originalFilename, config.DefaultTemplateExtensions)
	outputPath := filepath.Join(destination, transformedFilename)

	// Step 1: Read and render file content
	renderedContent, err := readAndRenderTemplate(filepath.Join(base, originalFilename), data)
	if err != nil {
		return err
	}

	// Step 2: Handle file existence and writing
	return handleOutputFile(outputPath, renderedContent, action, utils.FileOrDirExists, utils.WriteStringToFile)
}

func handleOutputFile(
	outputPath, renderedContent string,
	action Action,
	fileExistsFunc func(string) (bool, bool, error),
	writeFileFunc func(string, string) error,
) error {
	exists, isDir, err := fileExistsFunc(outputPath)
	if err != nil {
		return errors.Wrap("error checking output file", err, outputPath)
	}

	if exists {
		if isDir {
			return errors.Wrap("output path exists but is a directory", outputPath)
		}
		if action.SkipIfExists {
			return nil // Skip processing if the file exists and skip is enabled
		}
		if !action.Force {
			return errors.Wrap("file already exists and force is not set", outputPath)
		}
	}

	// Step 4: Write the rendered content to the output file
	if err := writeFileFunc(outputPath, renderedContent); err != nil {
		return errors.Wrap("failed to write rendered content", err, outputPath)
	}

	return nil
}

func isJsAction(action Action) bool {
	containsJs := func(s string) bool {
		s = filepath.ToSlash(strings.ToLower(s))
		return utils.ContainsSubstring(s, "/js") || strings.HasSuffix(s, "js")
	}

	if containsJs(action.TemplateFile) || containsJs(action.Path) {
		return true
	}

	if action.Item == "folder" {
		return containsJs(action.Source) || containsJs(action.Destination)
	}

	return false
}
