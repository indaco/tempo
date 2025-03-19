package componentcmd

import (
	"context"
	"path/filepath"

	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/errors"
	"github.com/indaco/tempo/internal/generator"
	"github.com/indaco/tempo/internal/helpers"
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
		Usage:                  "Generate a component instance from a template",
		UsageText:              "tempo component new [options]",
		UseShortOptionHandling: true,
		Flags:                  flags,
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
		if err := generator.ProcessEntityActions(cmdCtx.Logger, pathToComponentActionsFile, data, cmdCtx.Config); err != nil {
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
