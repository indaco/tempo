package componentcmd

import (
	"context"
	"path/filepath"

	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/errors"
	"github.com/indaco/tempo/internal/generator"
	"github.com/indaco/tempo/internal/helpers"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/resolver"
	"github.com/indaco/tempo/internal/templatefuncs/providers/gonameprovider"
	"github.com/indaco/tempo/internal/utils"
	"github.com/urfave/cli/v3"
)

/* ------------------------------------------------------------------------- */
/* Command Setup                                                             */
/* ------------------------------------------------------------------------- */

func setupComponentNewSubCommand(cmdCtx *app.AppContext) *cli.Command {
	flags := getNewFlags()
	return &cli.Command{
		Name:                   "new",
		Description:            "Generate a component instance from a template",
		UseShortOptionHandling: true,
		Flags:                  flags,
		ArgsUsage:              "[--package value | -p] [--assets value | -a] [--name value | -n] [--js] [--force] [--dry-run]",
		Before:                 validateComponentNewPrerequisites(cmdCtx.Config),
		Action:                 runComponentNewSubCommand(cmdCtx),
	}
}

/* ------------------------------------------------------------------------- */
/* Flag Generation                                                           */
/* ------------------------------------------------------------------------- */

// getNewFlags generates the core CLI flags shared across subcommands.
func getNewFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "package",
			Aliases: []string{"p"},
			Usage:   "The Go package name where components will be generated (default: components)",
		},
		&cli.StringFlag{
			Name:    "assets",
			Aliases: []string{"a"},
			Usage:   "The directory where asset files (e.g., CSS, JS) will be generated (default: assets)",
		},
		&cli.StringFlag{
			Name:     "name",
			Aliases:  []string{"n"},
			Usage:    "Name of the component",
			Required: true,
		},
		&cli.BoolFlag{
			Name:  "js",
			Usage: "Whether or not JS is needed for the component",
		},
		&cli.BoolFlag{
			Name:  "force",
			Usage: "Force overwriting if already exists",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "Preview actions without making changes",
		},
	}
}

/* ------------------------------------------------------------------------- */
/* Command Runner                                                            */
/* ------------------------------------------------------------------------- */

func runComponentNewSubCommand(cmdCtx *app.AppContext) func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		helpers.EnableLoggerIndentation(cmdCtx.Logger)

		// Step 1: Create template data
		data, err := createComponentData(cmd, cmdCtx.Config)
		if err != nil {
			return errors.Wrap("Failed to create template data for component", err)
		}

		if data.DryRun {
			cmdCtx.Logger.Info("Dry Run Mode: No changes will be made.\n")
			cmdCtx.Logger.Reset()
			return nil
		}

		// Step 2: Check if "define component" command has been executed
		pathToComponentActionsFile := filepath.Join(data.ActionsDir, "component.json")
		exists, err := utils.FileExistsFunc(pathToComponentActionsFile)
		if err != nil {
			return err
		}
		if !exists {
			return errors.Wrap("Cannot find actions folder. Did you run 'tempo define component' before?")
		}

		// Step 3: Check if the component already exists
		// Display a warning and stop if `--force` is not set
		outputPath := filepath.Join(data.GoPackage, data.ComponentName)
		if exists, err = utils.DirExists(outputPath); err != nil {
			return err
		} else if exists {
			utils.CheckEntityForNew("component", data.ComponentName, data.GoPackage, data.Force, cmdCtx.Logger)

			if !data.Force {
				return nil
			}
		}

		// Step 4: Retrieve and process actions
		if err := processEntityActions(cmdCtx.Logger, pathToComponentActionsFile, data, cmdCtx.Config); err != nil {
			return errors.Wrap("failed to process actions for component", err, data.ComponentName)
		}

		// Step 5: Log success and asset information
		componentPath := filepath.Join(data.GoPackage, data.ComponentName)
		assetPath := filepath.Join(data.AssetsDir, data.ComponentName)

		cmdCtx.Logger.Success("Templ component files have been created").
			WithAttrs(
				"component", data.ComponentName,
				"component_path", componentPath,
				"asset_path", assetPath,
			)

		cmdCtx.Logger.Reset()

		return nil
	}
}

/* ------------------------------------------------------------------------- */
/* Prerequisites Validation                                                  */
/* ------------------------------------------------------------------------- */

// validateComponentNewPrerequisites checks prerequisites for the "new component" subcommand, including:
// - Initialized Tempo project (inherited from the main define command).
// - Existence of the component templates folder.
func validateComponentNewPrerequisites(cfg *config.Config) func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	return func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		foldersToCheck := map[string]string{
			"templates_directory": filepath.Join(cfg.Paths.TemplatesDir, "component"),
		}

		missingFolders, err := utils.CheckMissingFolders(foldersToCheck)
		if err != nil {
			return nil, err
		}

		if len(missingFolders) > 0 {
			return nil, helpers.BuildMissingFoldersError(
				missingFolders,
				"Have you run 'tempo component define' or 'tempo component new' to set up your components?\nMake sure your templates, actions, and implementations exist before creating a new component.",
				[]string{"tempo component -h"},
			)
		}

		return ctx, nil
	}
}

/* ------------------------------------------------------------------------- */
/* Helper Functions                                                          */
/* ------------------------------------------------------------------------- */

// processEntityActions retrieves and processes actions from a JSON file.
func processEntityActions(logger logger.LoggerInterface, pathToActionsFile string, data *generator.TemplateData, cfg *config.Config) error {
	// Retrieve user actions
	userActions, err := retrieveActionsFile(logger, pathToActionsFile, cfg)
	if err != nil {
		return errors.Wrap("failed to get component actions file", err)
	}

	// Convert to built-in actions
	builtInActions := userActions.ToActions(generator.RenderActionId)

	if data.Force {
		for i := range builtInActions {
			builtInActions[i].Force = true
		}
	}

	// Process actions
	if err := generator.ProcessActions(logger, builtInActions, data); err != nil {
		return errors.Wrap("failed to process actions", err)
	}

	return nil
}

// retrieveActionsFile retrieves actions from a JSON file.
func retrieveActionsFile(logger logger.LoggerInterface, actionFilePath string, cfg *config.Config) (generator.JSONActionList, error) {
	var userActions generator.JSONActionList

	// Step 1: Resolve action file path
	resolvedPath, err := resolveActionFilePath(cfg.Paths.ActionsDir, actionFilePath)
	if err != nil {
		return nil, errors.Wrap("failed to resolve action file path", err)
	}

	// Step 2: Check if the action file exists
	if resolvedPath != "" {
		exists, err := utils.FileExistsFunc(resolvedPath)
		if err != nil {
			return nil, err
		} else if exists {
			// Step 3: Load user-defined actions
			userActions, err = generator.LoadUserActions(resolvedPath)
			if err != nil {
				return nil, errors.Wrap("failed to load user-defined actions from:", err, resolvedPath)
			}
			logger.Info("Actions loaded").
				WithAttrs(
					"action_file", resolvedPath,
					"num_actions", len(userActions),
				)
		} else {
			logger.Info("No user-defined actions found, proceeding with built-in actions only")
		}
	}

	return userActions, nil
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

// createComponentData initializes TemplateData for a component.
func createComponentData(cmd *cli.Command, cfg *config.Config) (*generator.TemplateData, error) {
	data, err := createBaseTemplateData(cmd, cfg)
	if err != nil {
		return nil, err
	}

	// Add component-specific fields
	data.ComponentName = gonameprovider.ToGoPackageName(cmd.String("name"))

	return data, nil
}

// createBaseTemplateData initializes common fields for TemplateData.
func createBaseTemplateData(cmd *cli.Command, cfg *config.Config) (*generator.TemplateData, error) {
	goPackage, err := resolver.ResolveString(
		cmd.String("package"),
		cfg.App.GoPackage,
		"package",
		config.DefaultGoPackage,
		nil,
	)
	if err != nil {
		return nil, err
	}

	assetsDir, err := resolver.ResolveString(
		cmd.String("assets"),
		cfg.App.AssetsDir,
		"assets folder",
		config.DefaultAssetsDir,
		nil,
	)
	if err != nil {
		return nil, err
	}

	TemplatesDir, ActionsDir := config.DerivedFolderPaths(cfg.TempoRoot)
	isWithJs := resolver.ResolveBool(cmd.Bool("js"), cfg.App.WithJs)
	isForce := cmd.Bool("force")
	isDryRun := cmd.Bool("dry-run")

	// Initialize common fields
	return &generator.TemplateData{
		TemplatesDir: TemplatesDir,
		ActionsDir:   ActionsDir,
		GoModule:     cfg.App.GoModule,
		GoPackage:    goPackage,
		AssetsDir:    assetsDir,
		WithJs:       isWithJs,
		CssLayer:     cfg.App.CssLayer,
		GuardMarker:  cfg.Templates.GuardMarker,
		Force:        isForce,
		DryRun:       isDryRun,
	}, nil
}
