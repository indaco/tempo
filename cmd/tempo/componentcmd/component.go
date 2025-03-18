package componentcmd

import (
	"context"

	"github.com/indaco/tempo/internal/app"
	"github.com/urfave/cli/v3"
)

// SetupComponentCommand creates the "component" command with its "define" and "new" subcommand.
func SetupComponentCommand(cmdCtx *app.AppContext) *cli.Command {
	return &cli.Command{
		Name:  "component",
		Usage: "Define component templates and generate instances from them",
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			return ctx, app.IsTempoProject(cmdCtx.CWD)
		},
		Commands: []*cli.Command{
			setupComponentDefineSubCommand(cmdCtx),
			setupComponentNewSubCommand(cmdCtx),
		},
	}
}
