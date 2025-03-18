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
	"github.com/indaco/tempo/internal/utils"
	"github.com/urfave/cli/v3"
)

/* ------------------------------------------------------------------------- */
/* Command Setup                                                             */
/* ------------------------------------------------------------------------- */

func setupComponentDefineSubCommand(cmdCtx *app.AppContext) *cli.Command {
	flags := getDefineFlags()

	return &cli.Command{
		Name:                   "define",
		Usage:                  "Define a new component template",
		UseShortOptionHandling: true,
		Flags:                  flags,
		ArgsUsage:              "[--js] [--force] [--dryrun]",
		Action:                 runComponentDefineSubCommand(cmdCtx),
	}
}

/* ------------------------------------------------------------------------- */
/* Flag Generation                                                           */
/* ------------------------------------------------------------------------- */

// getDefineFlags generates the core CLI flags shared across subcommands.
func getDefineFlags() []cli.Flag {
	return []cli.Flag{
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

func runComponentDefineSubCommand(cmdCtx *app.AppContext) func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		helpers.EnableLoggerIndentation(cmdCtx.Logger)

		// Step 1: Create template data
		data, err := createTemplateData(cmd, cmdCtx.Config)
		if err != nil {
			return errors.Wrap("Failed to create template data for component", err)
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
			utils.CheckEntityForDefine("component", outputPath, data.Force, cmdCtx.Logger)

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
			utils.LogSuccessMessages("component", cmdCtx.Config, cmdCtx.Logger)

			// Step 6: Generate JSON action file
			if err := generator.GenerateActionFile("component", data, builtInActions, cmdCtx.Logger); err != nil {
				return err
			}
		}
		helpers.ResetLogger(cmdCtx.Logger)

		return nil
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
