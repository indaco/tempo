package main

import (
	"context"
	"fmt"
	"os"

	"github.com/indaco/tempo/cmd/tempo/cleancmd"
	"github.com/indaco/tempo/cmd/tempo/definecmd"
	"github.com/indaco/tempo/cmd/tempo/initcmd"
	"github.com/indaco/tempo/cmd/tempo/newcmd"
	"github.com/indaco/tempo/cmd/tempo/registercmd"
	"github.com/indaco/tempo/cmd/tempo/synccmd"
	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/errors"
	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/utils"
	"github.com/indaco/tempo/internal/version"
	"github.com/urfave/cli/v3"
)

const (
	appName     = "tempo"
	description = "A simple CLI for scaffolding components and managing assets in templ-based projects"
	author      = "indaco"
	email       = "github@mircoveltri.me"
)

func main() {
	// Get the current working directory
	cwd := utils.GetCWD()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize CLI context
	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    cwd,
	}

	// Define CLI application
	app := &cli.Command{
		Name:        appName,
		Version:     fmt.Sprintf("v%s", version.GetVersion()),
		Description: description,
		Commands: []*cli.Command{
			initcmd.SetupInitCommand(cliCtx),
			definecmd.SetupDefineCommand(cliCtx),
			newcmd.SetupNewCommand(cliCtx),
			synccmd.SetupSyncCommand(cliCtx),
			registercmd.SetupRegisterCommand(cliCtx),
			cleancmd.SetupCleanCommand(cliCtx),
		},
	}

	// Run application
	if err := app.Run(context.Background(), os.Args); err != nil {
		errors.LogErrorChain(err)
		os.Exit(1)
	}
}
