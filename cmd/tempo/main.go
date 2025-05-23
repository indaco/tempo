package main

import (
	"context"
	"fmt"
	"log"
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
	usage       = "A lightweight CLI for managing assets and scaffolding components in templ-based projects."
	description = "tempo simplifies asset management in templ-based projects, providing a seamless workflow for handling CSS and JS files. It automatically extracts and injects styles and scripts into .templ components while preserving the original source files, ensuring a smooth developer experience. Additionally, it offers a lightweight scaffolding system to quickly generate component and variant templates with predefined structures."
	author      = "indaco"
	email       = "github@mircoveltri.me"
)

// main is the CLI application's entry point.
func main() {
	if err := runCLI(os.Args); err != nil {
		errors.LogErrorChain(err)
		log.Fatal(err)
	}
}

// runCLI sets up and runs the CLI application, returning any errors encountered during execution.
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

	appCmd := newCLI(cliCtx)

	// Run application.
	return appCmd.Run(context.Background(), args)
}

// newCLI creates and returns the root CLI command and its subcommands.
func newCLI(cliCtx *app.AppContext) *cli.Command {
	return &cli.Command{
		Name:        appName,
		Version:     fmt.Sprintf("v%s", version.GetVersion()),
		Usage:       usage,
		UsageText:   "tempo <subcommand> [options] [arguments]",
		Description: description,
		Commands: []*cli.Command{
			initcmd.SetupInitCommand(cliCtx),
			componentcmd.SetupComponentCommand(cliCtx),
			variantcmd.SetupVariantCommand(cliCtx),
			registercmd.SetupRegisterCommand(cliCtx),
			synccmd.SetupSyncCommand(cliCtx),
		},
	}
}
