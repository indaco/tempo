package newcmd

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
	"github.com/indaco/tempo/internal/templatefuncs/providers/gonameprovider"
	"github.com/indaco/tempo/internal/templatefuncs/providers/textprovider"
	"github.com/indaco/tempo/internal/utils"
	"github.com/urfave/cli/v3"
)

/* ------------------------------------------------------------------------- */
/* Command Setup                                                             */
/* ------------------------------------------------------------------------- */

func SetupNewCommand(cmdCtx *app.AppContext) *cli.Command {
	coreFlags := getCoreFlags()
	componentFlags := getComponentFlags()
	variantFlags := getVariantFlags()

	return &cli.Command{
		Name:        "new",
		Aliases:     []string{"n"},
		Description: "Generate a component or variant based on defined templates",
		Before:      validateNewPrerequisites(cmdCtx.CWD),
		Commands: []*cli.Command{
			setupNewComponentSubCommand(cmdCtx, append(coreFlags, componentFlags...)),
			setupNewVariantSubCommand(cmdCtx, append(coreFlags, variantFlags...)),
		},
	}
}

/* ------------------------------------------------------------------------- */
/* Flag Generation                                                           */
/* ------------------------------------------------------------------------- */

// getCoreFlags generates the core CLI flags shared across subcommands.
func getCoreFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "module",
			Aliases: []string{"m"},
			Usage:   "The name of the Go module being worked on",
		},
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
		&cli.BoolFlag{
			Name:  "watermark",
			Usage: "Whether or not include comments as watermark in generated files",
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

// getComponentFlags generates flags specific to the "component" subcommand.
func getComponentFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Usage: "Name of the component", Required: true},
		&cli.BoolFlag{Name: "js", Usage: "Whether or not JS is needed for the component"},
	}
}

// getVariantFlags generates flags specific to the "variant" subcommand.
func getVariantFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Usage: "The name of the variant being generated", Required: true},
		&cli.StringFlag{Name: "component", Aliases: []string{"c"}, Usage: "Name of the component or entity", Required: true},
	}
}

/* ------------------------------------------------------------------------- */
/* Subcommand Setup                                                          */
/* ------------------------------------------------------------------------- */

// setupNewComponentSubCommand creates the "component" subcommand for generating files for a new component.
func setupNewComponentSubCommand(cmdCtx *app.AppContext, flags []cli.Flag) *cli.Command {
	return &cli.Command{
		Name:                   "component",
		Description:            "Uses the templates and actions created with `tempo define` to generate a real component",
		Aliases:                []string{"c"},
		UseShortOptionHandling: true,
		Flags:                  flags,
		ArgsUsage:              "[--module value | -m] [--package value | -p] [--assets value | -a] [--name value | -n] [--js] [--watermark] [--force] [--dry-run]",
		Before:                 validateNewComponentPrerequisites(cmdCtx.Config),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			helpers.EnableLoggerIndentation(cmdCtx.Logger)

			// Step 1: Create template data
			data, err := createComponentData(cmd, cmdCtx.Config)
			if err != nil {
				return errors.Wrap("Failed to create template data", err)
			}

			if data.DryRun {
				cmdCtx.Logger.Info("Dry Run Mode: No changes will be made.\n")
			}

			// Step 2: Check if "define component" command has been executed
			pathToComponentActionsFile := filepath.Join(data.ActionsDir, "component.json")
			exists, err := utils.FileExists(pathToComponentActionsFile)
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
				handleEntityExistence("component", data.ComponentName, data.GoPackage, data.Force, cmdCtx.Logger)

				if !data.Force {
					return nil
				}
			}

			// Step 4: Retrieve and process actions
			if err := processEntityActions(cmdCtx.Logger, pathToComponentActionsFile, data, cmdCtx.Config); err != nil {
				return errors.Wrap("failed to process actions for component", err, data.ComponentName)
			}

			// Step 5: Log success and asset information
			if !data.DryRun {
				componentPath := filepath.Join(data.GoPackage, data.ComponentName)
				assetPath := filepath.Join(data.AssetsDir, data.ComponentName)

				cmdCtx.Logger.Success("Templ component files have been created").
					WithAttrs(
						"component", data.ComponentName,
						"component_path", componentPath,
						"asset_path", assetPath,
					)
			}
			cmdCtx.Logger.Reset()

			return nil
		},
	}
}

// setupNewVariantSubCommand creates the "variant" subcommand for generating files for a variant.
func setupNewVariantSubCommand(cmdCtx *app.AppContext, flags []cli.Flag) *cli.Command {
	return &cli.Command{
		Name:                   "variant",
		Description:            "Uses the templates and actions created with `tempo define` to generate a real component variant",
		Aliases:                []string{"v"},
		UseShortOptionHandling: true,
		Flags:                  flags,
		ArgsUsage:              "[--module value | -m] [--package value | -p] [--assets value | -a] [--name value | -n] [--component value | -c] [--watermark] [--force] [--dry-run]",
		Before:                 validateNewVariantPrerequisites(cmdCtx.Config),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			helpers.EnableLoggerIndentation(cmdCtx.Logger)

			// Step 1: Create variant data
			data, err := createVariantData(cmd, cmdCtx.Config)
			if err != nil {
				return errors.Wrap("failed to create variant data", err)
			}

			if data.DryRun {
				cmdCtx.Logger.Info("Dry Run Mode: No changes will be made.\n")
			}

			// Step 2: Check if "define variant" command has been executed
			pathToVariantActionsFile := filepath.Join(data.ActionsDir, "variant.json")
			exists, err := utils.FileExists(pathToVariantActionsFile)
			if err != nil {
				return err
			}
			if !exists {
				return errors.Wrap("Cannot find actions folder. Did you run 'tempo define variant' before?")
			}

			// Step 3: Ensure the component folder exists before adding a variant
			componentFolderPath := filepath.Join(data.GoPackage, data.ComponentName)
			if exists, err := utils.DirExists(componentFolderPath); err != nil {
				return errors.Wrap("Error checking component folder", err, data.ComponentName)
			} else if !exists {
				cmdCtx.Logger.Error("Cannot create variant: Component does not exist").
					WithAttrs(
						"variant", data.VariantName,
						"component", data.ComponentName,
					)
				return errors.Wrap("Cannot create variant: Component does not exist", data.ComponentName)
			}

			// Step 4: Check if the component variant already exists with the same name
			// Display a warning and stop if `--force` is not set
			outputPath := filepath.Join(data.GoPackage, data.ComponentName, "css", "variants", data.VariantName+".templ")
			if exists, err := utils.FileExists(outputPath); err != nil {
				return err
			} else if exists {
				handleEntityExistence("variant", data.VariantName, outputPath, data.Force, cmdCtx.Logger)

				if !data.Force {
					return nil
				}
			}

			// Step 5: Retrieve and process actions
			if err := processEntityActions(cmdCtx.Logger, pathToVariantActionsFile, data, cmdCtx.Config); err != nil {
				return errors.Wrap("failed to process actions for variant", err, data.ComponentName)
			}

			// Step 6: Log success and asset information
			if !data.DryRun {
				// Define paths for components and assets
				componentPath := filepath.Join(data.GoPackage, data.ComponentName, "css", "variant")
				assetPath := filepath.Join(data.AssetsDir, data.ComponentName, "css", "variants")

				// Log the success message with structured attributes
				cmdCtx.Logger.Success("Templ component for the variant and asset files (CSS) have been created").
					WithAttrs(
						"variant", data.VariantName,
						"component", data.ComponentName,
						"component_path", componentPath,
						"asset_path", assetPath,
					)

				cmdCtx.Logger.Blank()
				cmdCtx.Logger.Hint(fmt.Sprintf("Update %s/css/base.templ to conditionally load the variant's styles.", data.ComponentName))
			}
			cmdCtx.Logger.Reset()

			return nil
		},
	}
}

/* ------------------------------------------------------------------------- */
/* Prerequisites Validation                                                  */
/* ------------------------------------------------------------------------- */

// validateNewPrerequisites checks if the required configuration file exists for the "new" command,including:
// - Initialized Tempo project (inherit from the main define command).
func validateNewPrerequisites(folderPathToConfig string) func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	return func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		return ctx, app.IsTempoProject(folderPathToConfig)
	}
}

// validateNewComponentPrerequisites checks prerequisites for the "new component" subcommand, including:
// - Initialized Tempo project (inherited from the main define command).
// - Existence of the component templates folder.
func validateNewComponentPrerequisites(cfg *config.Config) func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	return func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		foldersToCheck := map[string]string{
			"Templates directory": filepath.Join(cfg.Paths.TemplatesDir, "component"),
		}

		missingFolders, err := utils.CheckMissingFolders(foldersToCheck)
		if err != nil {
			return nil, err
		}

		if len(missingFolders) > 0 {
			return nil, helpers.BuildMissingFoldersError(
				missingFolders,
				"Have you run 'tempo define' or 'tempo create' to set up your components?\nMake sure your templates, actions, and implementations exist before creating a new component.",
				[]string{"tempo define -h", "tempo create -h"},
			)
		}

		return ctx, nil
	}
}

// validateNewVariantPrerequisites checks prerequisites for the "new variant" subcommand, including:
// - Initialized Tempo project (inherited from the main define command).
// - Existence of the component templates folder.
// - Existence of the variant templates folder.
func validateNewVariantPrerequisites(cfg *config.Config) func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	return func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		foldersToCheck := map[string]string{
			"Component directory": filepath.Join(cfg.Paths.TemplatesDir, "component"),
			"Variant directory":   filepath.Join(cfg.Paths.TemplatesDir, "component-variant"),
		}

		missingFolders, err := utils.CheckMissingFolders(foldersToCheck)
		if err != nil {
			return nil, err
		}

		if len(missingFolders) > 0 {
			return nil, helpers.BuildMissingFoldersError(
				missingFolders,
				"Have you run 'tempo define' or 'tempo create' to set up your components?\nMake sure your templates, actions, and implementations exist before creating a new variant.",
				[]string{"tempo define -h", "tempo create -h"},
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

// createVariantData initializes TemplateData for a variant.
func createVariantData(cmd *cli.Command, cfg *config.Config) (*generator.TemplateData, error) {
	data, err := createBaseTemplateData(cmd, cfg)
	if err != nil {
		return nil, err
	}

	// Add variant-specific fields
	data.VariantName = cmd.String("name")
	data.ComponentName = gonameprovider.ToGoPackageName(cmd.String("component"))
	return data, nil
}

// createBaseTemplateData initializes common fields for TemplateData.
func createBaseTemplateData(cmd *cli.Command, cfg *config.Config) (*generator.TemplateData, error) {
	moduleName, err := resolver.ResolveString(cmd.String("module"), cfg.App.GoModule, "module name")
	if err != nil {
		return nil, err
	}

	packageName, err := resolver.ResolveString(cmd.String("package"), cfg.App.GoPackage, "package")
	if err != nil {
		return nil, err
	}

	assetsDir, err := resolver.ResolveString(cmd.String("assets"), cfg.App.AssetsDir, "assets folder")
	if err != nil {
		return nil, err
	}

	TemplatesDir, ActionsDir := config.DerivedFolderPaths(cfg.TempoRoot)
	isWithJs := resolver.ResolveBool(cmd.Bool("js"), cfg.App.WithJs)
	isWatermarkTip := resolver.ResolveBool(cmd.Bool("watermark"), cfg.Templates.WatermarkTip)
	isForce := cmd.Bool("force")
	isDryRun := cmd.Bool("dry-run")

	// Initialize common fields
	return &generator.TemplateData{
		TemplatesDir: TemplatesDir,
		ActionsDir:   ActionsDir,
		GoModule:     moduleName,
		GoPackage:    packageName,
		AssetsDir:    assetsDir,
		WithJs:       isWithJs,
		CssLayer:     cfg.App.CssLayer,
		WatermarkTip: isWatermarkTip,
		GuardMarker:  cfg.Templates.GuardMarker,
		Force:        isForce,
		DryRun:       isDryRun,
	}, nil
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
		exists, err := utils.FileExists(resolvedPath)
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
		exists, err := utils.FileExists(resolvedPath)
		if err != nil {
			return "", err
		} else if exists {
			return resolvedPath, nil
		}
	}

	// Step 2: Check if the provided actionFileFlag is a valid full path
	// Check the actionFileFlag as an absolute path
	exists, err := utils.FileExists(actionFileFlag)
	if err != nil {
		return "", errors.Wrap("error checking action file path", err, actionFileFlag)
	}
	if !exists {
		return "", errors.Wrap("action file does not exist", actionFileFlag)
	}

	return actionFileFlag, nil
}

func handleEntityExistence(entityType, entityName, outputPath string, force bool, logr logger.LoggerInterface) {
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
