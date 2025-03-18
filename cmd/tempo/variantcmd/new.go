package variantcmd

import (
	"context"
	"fmt"
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

// setupVariantNewSubCommand creates the "variant" subcommand for generating files for a variant.
func setupVariantNewSubCommand(cmdCtx *app.AppContext) *cli.Command {
	flags := getNewFlags()
	return &cli.Command{
		Name:                   "new",
		Description:            "Generate a variant instance from a template",
		UseShortOptionHandling: true,
		Flags:                  flags,
		ArgsUsage:              "[--package value | -p] [--assets value | -a] [--name value | -n] [--component value | -c] [--force] [--dry-run]",
		Before:                 validateVariantNewPrerequisites(cmdCtx.Config),
		Action:                 runVariantNewSubCommand(cmdCtx),
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
			Usage:    "The name of the variant being generated",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "component",
			Aliases:  []string{"c"},
			Usage:    "Name of the component or entity",
			Required: true,
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
func runVariantNewSubCommand(cmdCtx *app.AppContext) func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		helpers.EnableLoggerIndentation(cmdCtx.Logger)

		// Step 1: Create variant data
		data, err := createVariantData(cmd, cmdCtx.Config)
		if err != nil {
			return errors.Wrap("failed to create variant data", err)
		}

		if data.DryRun {
			cmdCtx.Logger.Info("Dry Run Mode: No changes will be made.\n")
			cmdCtx.Logger.Reset()
			return nil
		}

		// Step 2: Check if "define variant" command has been executed
		pathToVariantActionsFile := filepath.Join(data.ActionsDir, "variant.json")
		exists, err := utils.FileExistsFunc(pathToVariantActionsFile)
		if err != nil {
			return err
		}
		if !exists {
			return errors.Wrap("Cannot find actions folder. Did you run 'tempo variant define' before?")
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
		if exists, err := utils.FileExistsFunc(outputPath); err != nil {
			return err
		} else if exists {
			utils.CheckEntityForNew("variant", data.VariantName, outputPath, data.Force, cmdCtx.Logger)

			if !data.Force {
				return nil
			}
		}

		// Step 5: Retrieve and process actions
		if err := generator.ProcessEntityActions(cmdCtx.Logger, pathToVariantActionsFile, data, cmdCtx.Config); err != nil {
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
	}
}

/* ------------------------------------------------------------------------- */
/* Prerequisites Validation                                                  */
/* ------------------------------------------------------------------------- */

// validateVariantNewPrerequisites checks prerequisites for the "new variant" subcommand, including:
// - Initialized Tempo project (inherited from the main define command).
// - Existence of the component templates folder.
// - Existence of the variant templates folder.
func validateVariantNewPrerequisites(cfg *config.Config) func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	return func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		foldersToCheck := map[string]string{
			"component_directory": filepath.Join(cfg.Paths.TemplatesDir, "component"),
			"variant_directory":   filepath.Join(cfg.Paths.TemplatesDir, "component-variant"),
		}

		missingFolders, err := utils.CheckMissingFolders(foldersToCheck)
		if err != nil {
			return nil, err
		}

		if len(missingFolders) > 0 {
			return nil, helpers.BuildMissingFoldersError(
				missingFolders,
				"Have you run 'tempo component define' or 'tempo variant define' to set up your components?\nMake sure your templates, actions, and implementations exist before creating a new variant.",
				[]string{"tempo component -h", "tempo define -h"},
			)
		}

		return ctx, nil
	}
}

/* ------------------------------------------------------------------------- */
/* Helper Functions                                                          */
/* ------------------------------------------------------------------------- */

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
