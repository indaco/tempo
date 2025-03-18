package main

import (
	"context"
	"fmt"
	"os"

	"github.com/indaco/tempo/cmd/tempo/componentcmd"
	"github.com/indaco/tempo/cmd/tempo/initcmd"
	"github.com/indaco/tempo/cmd/tempo/registercmd"
	"github.com/indaco/tempo/cmd/tempo/synccmd"
	"github.com/indaco/tempo/cmd/tempo/variantcmd"
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

// runCLI encapsulates the main cli logic and returns an error.
func runCLI(args []string) error {
	// Get current working directory.
	cwd := utils.GetCWD()

	// Load configuration.
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	// Initialize CLI context.
	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: cfg,
		CWD:    cwd,
	}

	// Define CLI application.
	appCmd := &cli.Command{
		Name:        appName,
		Version:     fmt.Sprintf("v%s", version.GetVersion()),
		Description: description,
		Commands: []*cli.Command{
			initcmd.SetupInitCommand(cliCtx),
			componentcmd.SetupComponentCommand(cliCtx),
			variantcmd.SetupVariantCommand(cliCtx),
			registercmd.SetupRegisterCommand(cliCtx),
			synccmd.SetupSyncCommand(cliCtx),
		},
	}

	// Run application.
	return appCmd.Run(context.Background(), args)
}

func main() {
	if err := runCLI(os.Args); err != nil {
		errors.LogErrorChain(err)
		os.Exit(1)
	}
}
