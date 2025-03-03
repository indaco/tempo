package definecmd

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/errors"
	"github.com/indaco/tempo/internal/generator"
	"github.com/indaco/tempo/internal/helpers"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/resolver"
	"github.com/indaco/tempo/internal/utils"
	"github.com/urfave/cli/v3"
)

/* ------------------------------------------------------------------------- */
/* Define Command Setup                                                      */
/* ------------------------------------------------------------------------- */

// SetupDefineCommand creates the "define" command for combining "add" and "generate action".
func SetupDefineCommand(cmdCtx *app.AppContext) *cli.Command {
	coreFlags := getCoreFlags()
	componentFlags := getComponentFlags()

	return &cli.Command{
		Name:        "define",
		Aliases:     []string{"d"},
		Description: "Define templates and actions for component or variant",
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			return ctx, app.IsTempoProject(cmdCtx.CWD)
		},
		Commands: []*cli.Command{
			setupDefineComponentSubCommand(cmdCtx, append(coreFlags, componentFlags...)),
			setupDefineVariantSubCommand(cmdCtx, coreFlags),
		},
	}
}

/* ------------------------------------------------------------------------- */
/* Flag Generation                                                           */
/* ------------------------------------------------------------------------- */

// getCoreFlags generates the core CLI flags shared across subcommands.
func getCoreFlags() []cli.Flag {
	return []cli.Flag{
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

// getComponentFlags generates the flags specific to the "component" subcommand.
func getComponentFlags() []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:  "js",
			Usage: "Whether or not JS is needed for the component",
		},
	}
}

/* ------------------------------------------------------------------------- */
/* Component Subcommand Setup                                                */
/* ------------------------------------------------------------------------- */

func setupDefineComponentSubCommand(cmdCtx *app.AppContext, flags []cli.Flag) *cli.Command {
	return &cli.Command{
		Name:                   "component",
		Description:            "Define templates and action for component",
		Aliases:                []string{"c"},
		UseShortOptionHandling: true,
		Flags:                  flags,
		ArgsUsage:              "[--js] [--force] [--dryrun]",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			helpers.EnableLoggerIndentation(cmdCtx.Logger)

			// Step 1: Create template data
			data, err := createTemplateData(cmd, cmdCtx.Config)
			if err != nil {
				return errors.Wrap("Failed to create template data", err)
			}

			if data.DryRun {
				cmdCtx.Logger.Info("Dry Run Mode: No changes will be made.")
			}

			// Step 2: Check if templates folder for component already exists
			// Display a warning and stop if `--force` is not set
			outputPath := filepath.Join(data.TemplatesDir, "component")
			exists, err := utils.DirExists(outputPath)
			if err != nil {
				return err
			} else if exists {
				handleEntityExistence("component", outputPath, data.Force, cmdCtx.Logger)

				if !data.Force {
					return nil
				}
			}

			// Step 3: Retrieve component actions
			builtInActions, err := generator.BuildComponentActions(generator.CopyActionId, data.Force, false)
			if err != nil {
				return errors.Wrap("Failed to build component actions", err)
			}

			// Step 4: Process actions
			if err := generator.ProcessActions(cmdCtx.Logger, builtInActions, data); err != nil {
				return errors.Wrap("Failed to process actions for component", err)
			}

			if !data.DryRun {
				// Step 5: Log success and asset information
				logSuccessMessages(cmdCtx.Config, "component", cmdCtx.Logger)

				// Step 6: Generate JSON action file
				if err := generateActionFile(data, builtInActions, "component", cmdCtx.Logger); err != nil {
					return err
				}
			}
			helpers.ResetLogger(cmdCtx.Logger)

			return nil
		},
	}
}

/* ------------------------------------------------------------------------- */
/* Variant Subcommand Setup                                                  */
/* ------------------------------------------------------------------------- */

func setupDefineVariantSubCommand(cmdCtx *app.AppContext, flags []cli.Flag) *cli.Command {
	return &cli.Command{
		Name:                   "variant",
		Description:            "Define templates and action for component variant",
		Aliases:                []string{"v"},
		UseShortOptionHandling: true,
		Flags:                  flags,
		ArgsUsage:              "[--force] [--dryrun]",
		Before:                 validateDefineVariantPrerequisites(cmdCtx.Config),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			helpers.EnableLoggerIndentation(cmdCtx.Logger)

			// Step 1: Create template data
			data, err := createTemplateData(cmd, cmdCtx.Config)
			if err != nil {
				return errors.Wrap("Failed to create template data", err)
			}

			if data.DryRun {
				cmdCtx.Logger.Info("Dry Run Mode: No changes will be made.")
			}

			// Step 2: Check if templates folder for component already exists
			// Display a warning and stop if `--force` is not set
			outputPath := filepath.Join(data.TemplatesDir, "component-variant")
			exists, err := utils.DirExists(outputPath)
			if err != nil {
				return err
			} else if exists {
				handleEntityExistence("variant", outputPath, data.Force, cmdCtx.Logger)

				if !data.Force {
					return nil
				}
			}

			// Step 3: Retrieve component actions
			builtInActions, err := generator.BuildVariantActions(generator.CopyActionId, false)
			if err != nil {
				return errors.Wrap("Failed to build variant actions", err)
			}

			// Step 4: Process actions
			if err := generator.ProcessActions(cmdCtx.Logger, builtInActions, data); err != nil {
				return errors.Wrap("Failed to process actions for variant", err)
			}

			if !data.DryRun {
				// Step 5: Log success and asset information
				logSuccessMessages(cmdCtx.Config, "variant", cmdCtx.Logger)

				// Step 6: Generate JSON action file
				if err := generateActionFile(data, builtInActions, "variant", cmdCtx.Logger); err != nil {
					return err
				}
			}
			helpers.ResetLogger(cmdCtx.Logger)

			return nil
		},
	}
}

/* ------------------------------------------------------------------------- */
/* Prerequisites Validation                                                  */
/* ------------------------------------------------------------------------- */

// validateDefineVariantPrerequisites checks prerequisites for the "define variant" subcommand, including:
// - Initialized Tempo project (inherit from the main define command).
// - Existence of the component templates folder.
func validateDefineVariantPrerequisites(cfg *config.Config) func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	return func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		pathToTemplatesComponent := filepath.Join(cfg.Paths.TemplatesDir, "component")
		exists, err := utils.DirExists(pathToTemplatesComponent)
		if err != nil {
			return nil, errors.Wrap("Failed to check component templates folder", err)
		}
		if !exists {
			return nil, errors.Wrap("Templates for component not found. Run 'tempo define component' first.")
		}
		return ctx, nil
	}
}

/* ------------------------------------------------------------------------- */
/* Helper Functions                                                          */
/* ------------------------------------------------------------------------- */

// createTemplateData initializes the common fields of TemplateData
// by resolving configuration values and CLI flags.
func createTemplateData(cmd *cli.Command, cfg *config.Config) (*generator.TemplateData, error) {
	TemplatesDir, ActionsDir := config.DerivedFolderPaths(cfg.TempoRoot)
	isWithJs := resolver.ResolveBool(cmd.Bool("js"), cfg.App.WithJs)
	isForce := cmd.Bool("force")
	isDryRun := cmd.Bool("dry-run")

	// Initialize common fields
	return &generator.TemplateData{
		TemplatesDir: TemplatesDir,
		ActionsDir:   ActionsDir,
		WithJs:       isWithJs,
		Force:        isForce,
		DryRun:       isDryRun,
	}, nil
}

// logSuccessMessages logs the success messages for generated template files and assets.
func logSuccessMessages(cfg *config.Config, entityType string, logger logger.LoggerInterface) {
	var basePath, assetsPath, message string

	switch entityType {
	case "component":
		basePath = filepath.Join(cfg.Paths.TemplatesDir, "component", "templ")
		assetsPath = filepath.Join(cfg.Paths.TemplatesDir, "component", "assets", "css")
		message = "Templates for the component and assets (CSS and JS) have been created"

	case "variant":
		basePath = filepath.Join(cfg.Paths.TemplatesDir, "component-variant", "templ")
		assetsPath = filepath.Join(cfg.Paths.TemplatesDir, "component-variant", "assets", "css")
		message = "Templates for the component variant and assets (CSS) have been created"

	default:
		logger.Warning("Unknown entity type:", entityType)
		return
	}

	logger.Success(message).
		WithAttrs(
			"templates_path", basePath,
			"assets_path", assetsPath,
		)
}

// generateActionFile generates an action file for the given entity type.
func generateActionFile(data *generator.TemplateData, actions []generator.Action, entityType string, logger logger.LoggerInterface) error {
	actionFileName := fmt.Sprintf("%s.json", entityType)
	actionsPath := filepath.Join(data.ActionsDir, actionFileName)

	if err := generator.GenerateActionJSONFile(actionsPath, actions); err != nil {
		return errors.Wrap("Failed to generate action file for %s", entityType, err)
	}

	logger.Success(fmt.Sprintf("Tempo action file for '%s' has been created", entityType)).
		WithAttrs("action_file_path", actionsPath)
	return nil
}

func handleEntityExistence(entityType, outputPath string, force bool, logr logger.LoggerInterface) {
	// Select logging function and action message based on `force` flag
	logFunc, action := logr.Warning, "Use '--force' to overwrite them. Any changes will be lost."
	if force {
		logFunc, action = logr.Info, "Overwriting due to '--force' flag."
	}

	msg := fmt.Sprintf("Templates for '%s' already exist.\n  %s", entityType, action)
	logFunc(msg).WithAttrs("path", outputPath)
}
