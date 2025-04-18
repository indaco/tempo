package variantcmd

import (
	"context"

	"github.com/indaco/tempo/internal/app"
	"github.com/urfave/cli/v3"
)

// SetupVariantCommand creates the "component" command with its "define" and "new" subcommand.
func SetupVariantCommand(cmdCtx *app.AppContext) *cli.Command {
	return &cli.Command{
		Name:      "variant",
		Usage:     "Define variant templates and generate instances from them",
		UsageText: "tempo variant <subcommand> [options] [arguments]",
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			return ctx, app.IsTempoProject(cmdCtx.CWD)
		},
		Commands: []*cli.Command{
			setupVariantDefineSubCommand(cmdCtx),
			setupVariantNewSubCommand(cmdCtx),
		},
	}
}
