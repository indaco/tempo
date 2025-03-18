package registercmd

import (
	"context"
	"fmt"

	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/errors"
	"github.com/indaco/tempo/internal/helpers"
	"github.com/indaco/tempo/internal/resolver"
	"github.com/indaco/tempo/internal/templatefuncs/loader"
	"github.com/indaco/tempo/internal/utils"
	"github.com/urfave/cli/v3"
)

/* ------------------------------------------------------------------------- */
/* Command Setup                                                             */
/* ------------------------------------------------------------------------- */

// SetupRegisterCommand sets up the "register" command for registering external functions.
func SetupRegisterCommand(cmdCtx *app.AppContext) *cli.Command {
	return &cli.Command{
		Name:  "register",
		Usage: "Register is used to extend tempo.",
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			return ctx, app.IsTempoProject(cmdCtx.CWD)
		},
		Commands: []*cli.Command{
			setupRegisterFunctionsSubCommand(cmdCtx, getFlags()),
		},
	}
}

/* ------------------------------------------------------------------------- */
/* Flag Generation                                                           */
/* ------------------------------------------------------------------------- */

func getFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "name",
			Aliases: []string{"n"},
			Usage:   "Name for the function provider",
		},
		&cli.StringFlag{
			Name:    "url",
			Aliases: []string{"u"},
			Usage:   "Repository URL",
		},
		&cli.StringFlag{
			Name:    "path",
			Aliases: []string{"p"},
			Usage:   "Path to a local go module provider",
		},
		&cli.BoolFlag{
			Name:  "force",
			Usage: "Force re-cloning the repository and pull the latest changes",
		},
	}
}

/* ------------------------------------------------------------------------- */
/* Register Functions Subcommand                                             */
/* ------------------------------------------------------------------------- */

func setupRegisterFunctionsSubCommand(cmdCtx *app.AppContext, flags []cli.Flag) *cli.Command {
	return &cli.Command{
		Name:                   "functions",
		Description:            "Register a function provider from a local go module path or a remote repository",
		Aliases:                []string{"f"},
		UseShortOptionHandling: true,
		Flags:                  flags,
		ArgsUsage:              "[--name value | -n] [--url value | -u] [--path value | -p] [--force]",
		Action:                 runRegisterFunctionsSubCommand(cmdCtx),
	}
}

/* ------------------------------------------------------------------------- */
/* Command Runner                                                            */
/* ------------------------------------------------------------------------- */

func runRegisterFunctionsSubCommand(cmdCtx *app.AppContext) func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		helpers.EnableLoggerIndentation(cmdCtx.Logger)

		forceClone := cmd.Bool("force")

		providers, err := resolveFlagsToTemplateFuncProvider(cmd)
		if err != nil {
			return err
		}

		// Register each provider
		for _, provider := range providers {
			switch provider.Type {
			case "url":
				err = registerFunctionsFromRepo(cmdCtx, forceClone, provider)
			case "path":
				err = registerFunctionsFromLocal(cmdCtx, provider)
			default:
				err = fmt.Errorf("unknown provider type: %s", provider.Type)
			}

			if err != nil {
				return err
			}
		}

		cmdCtx.Logger.Success("Functions successfully registered!")
		helpers.ResetLogger(cmdCtx.Logger)
		return nil
	}
}

/* ------------------------------------------------------------------------- */
/* Helper Functions                                                          */
/* ------------------------------------------------------------------------- */

func resolveFlagsToTemplateFuncProvider(cmd *cli.Command) ([]config.TemplateFuncProvider, error) {
	var providers []config.TemplateFuncProvider

	// Resolve values with priority: CLI > Config
	name, _ := resolver.ResolveString(cmd.String("name"), "", "provider name", "", nil)
	url, _ := resolver.ResolveString(cmd.String("url"), "", "repository URL", "", nil)
	path, _ := resolver.ResolveString(cmd.String("path"), "", "local path", "", nil)

	if name == "" {
		if url != "" {
			name = utils.ExtractNameFromURL(url) // Extract repo name
		} else if path != "" {
			name = utils.ExtractNameFromPath(path) // Extract last folder name
		}
	}

	// If both URL and Path are provided, ensure unique names
	if url != "" && path != "" {
		providers = append(providers, config.TemplateFuncProvider{
			Name:  name + "-repo", // Differentiate the repo version
			Type:  "url",
			Value: url,
		})

		providers = append(providers, config.TemplateFuncProvider{
			Name:  name + "-local", // Differentiate the local version
			Type:  "path",
			Value: path,
		})

		return providers, nil
	}

	if url != "" {
		providers = append(providers, config.TemplateFuncProvider{
			Name:  name,
			Type:  "url",
			Value: url,
		})
	}

	if path != "" {
		providers = append(providers, config.TemplateFuncProvider{
			Name:  name,
			Type:  "path",
			Value: path,
		})
	}

	return providers, nil
}

func registerFunctionsFromRepo(cmdCtx *app.AppContext, forceClone bool, provider config.TemplateFuncProvider) error {
	cmdCtx.Logger.Info("Fetching functions from repository...").WithAttrs("url", provider.Value)

	if err := loader.InstallFunctionPackageFromRepo(
		cmdCtx.Config.TempoRoot,
		provider.Value,
		forceClone,
		cmdCtx.Logger,
	); err != nil {
		return errors.Wrap("Failed to install function package from repository", err)
	}

	return nil
}

func registerFunctionsFromLocal(cmdCtx *app.AppContext, provider config.TemplateFuncProvider) error {
	cmdCtx.Logger.Info("Registering functions from local package...").WithAttrs("path", provider.Value)

	if err := loader.RegisterFunctionsFromPath(provider.Value, cmdCtx.Logger); err != nil {
		return errors.Wrap("Failed to register function package from local path", err)
	}

	return nil
}
