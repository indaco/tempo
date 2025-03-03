package initcmd

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/errors"
	"github.com/indaco/tempo/internal/helpers"
	"github.com/indaco/tempo/internal/utils"
	"github.com/urfave/cli/v3"
)

/* ------------------------------------------------------------------------- */
/* Command Setup                                                             */
/* ------------------------------------------------------------------------- */

// SetupInitCommand creates the "init" command for initializing a Tempo project.
func SetupInitCommand(cmdCtx *app.AppContext) *cli.Command {
	const configFileName = "tempo.yaml"

	return &cli.Command{
		Name:                   "init",
		Aliases:                []string{"i"},
		Description:            "Initialize a Tempo project",
		UseShortOptionHandling: true,
		Flags:                  getFlags(),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			helpers.EnableLoggerIndentation(cmdCtx.Logger)

			// Step 1: resolve base folder for tempo files & config
			userBaseFolder := cmd.String("base-folder")
			TempoRoot := filepath.Join(userBaseFolder, cmdCtx.Config.TempoRoot)
			tempoConfigPath := filepath.Join(userBaseFolder, configFileName)

			// Step 2: ensure configuration file does not already exist
			if err := validateInitPrerequisites(tempoConfigPath); err != nil {
				return err
			}

			// Step 3: Resolve derived folders
			TemplatesDir, ActionsDir := config.DerivedFolderPaths(userBaseFolder)

			// Step 4: Generate and write the configuration file
			cmdCtx.Logger.Info("Generating", tempoConfigPath)
			cfg := prepareConfig(TempoRoot, TemplatesDir, ActionsDir)
			if err := writeConfigFile(tempoConfigPath, cfg); err != nil {
				return errors.Wrap("Failed to write the configuration file", err, tempoConfigPath)
			}

			// Step 5: Log the successful initialization
			cmdCtx.Logger.Success("Done!", "Customize it to match your project needs.")
			helpers.ResetLogger(cmdCtx.Logger)

			return nil
		},
	}
}

/* ------------------------------------------------------------------------- */
/* Flag Generation                                                           */
/* ------------------------------------------------------------------------- */

// getFlags generates the CLI flags.
func getFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:  "base-folder",
			Usage: "Specify the base folder for Tempo files (default: current directory)",
		},
	}
}

/* ------------------------------------------------------------------------- */
/* Prerequisites Validation                                                  */
/* ------------------------------------------------------------------------- */

// validateInitPrerequisites ensures all the prerequisites for the init command are satisfied.
//
// - configuration file does not already exist.
func validateInitPrerequisites(configFilePath string) error {
	exists, err := utils.FileExists(configFilePath)
	if err != nil {
		return errors.Wrap("Error checking configuration file", err)
	}
	if exists {
		return errors.Wrap("Configuration file already exists", configFilePath)
	}
	return nil
}

/* ------------------------------------------------------------------------- */
/* Helper Functions                                                          */
/* ------------------------------------------------------------------------- */

// prepareConfig creates a new Config instance with the provided base folder, templates folder, and actions folder.
func prepareConfig(TempoRoot, TemplatesDir, ActionsDir string) *config.Config {
	return &config.Config{
		TempoRoot: path.Base(TempoRoot),
		App: config.App{
			GoPackage: config.DefaultGoPackage,
			AssetsDir: config.DefaultAssetsDir,
			CssLayer:  config.DefaultCssLayer,
		},
		Paths: config.Paths{
			TemplatesDir: TemplatesDir,
			ActionsDir:   ActionsDir,
		},
		Processor: config.Processor{
			Workers:       config.DefaultNumWorkers,
			SummaryFormat: config.DefaultSummaryFormat,
		},
		Templates: config.Templates{
			Extensions:   config.DefaultTemplateExtensions,
			WatermarkTip: config.DefaultWatermarkTipFlag,
			GuardMarker:  config.DefaultGuardMarkText,
		},
	}
}

// writeConfigFile writes the configuration to a YAML file with proper formatting and comments.
func writeConfigFile(filePath string, cfg *config.Config) error {
	var sb strings.Builder

	// Write header comment and base folder
	sb.WriteString("# The root folder for Tempo files\n")
	sb.WriteString(fmt.Sprintf("tempo_root: %s\n\n", cfg.TempoRoot))

	// Write app-specific configuration
	sb.WriteString("app:\n")
	sb.WriteString("  # The name of the Go module being worked on.\n")
	sb.WriteString(fmt.Sprintf("  # go_module: %s\n\n", cfg.App.GoModule))
	sb.WriteString("  # The Go package name where components will be organized and generated.\n")
	sb.WriteString(fmt.Sprintf("  go_package: %s\n\n", cfg.App.GoPackage))
	sb.WriteString("  # The directory where asset files (CSS, JS) will be generated.\n")
	sb.WriteString(fmt.Sprintf("  assets_dir: %s\n\n", cfg.App.AssetsDir))
	sb.WriteString("  # Indicates whether JavaScript is required for the component.\n")
	sb.WriteString(fmt.Sprintf("  # with_js: %s\n\n", strconv.FormatBool(cfg.App.WithJs)))
	sb.WriteString("  # The name of the CSS layer to associate with component styles.\n")
	sb.WriteString(fmt.Sprintf("  # css_layer: %s\n\n", cfg.App.CssLayer))

	// Write processor configuration
	sb.WriteString("processor:\n")
	sb.WriteString("  # Number of concurrent workers (numCPUs * 2).\n")
	sb.WriteString(fmt.Sprintf("  workers: %d\n\n", cfg.Processor.Workers))
	sb.WriteString("  # Summary format: text, json, none.\n")
	sb.WriteString(fmt.Sprintf("  summary_format: %s\n\n", cfg.Processor.SummaryFormat))

	// Write templates configuration
	sb.WriteString("templates:\n")
	sb.WriteString("  # Whether or not to include comments as watermarks in generated files.\n")
	sb.WriteString(fmt.Sprintf("  watermark: %s\n\n", strconv.FormatBool(cfg.Templates.WatermarkTip)))
	sb.WriteString("  # A placeholder in template files indicating auto-generated sections.\n")
	sb.WriteString(fmt.Sprintf("  guard_marker: %s\n\n", cfg.Templates.GuardMarker))
	sb.WriteString("  # File extensions used for template files.\n")
	sb.WriteString("  extensions:\n")

	extensions := cfg.Templates.Extensions
	if len(extensions) == 0 {
		extensions = config.DefaultTemplateExtensions // Use default if none are set
	}

	for _, ext := range extensions {
		sb.WriteString(fmt.Sprintf("    - %s\n", ext))
	}

	// Add function providers section
	formatFunctionProviders(&sb, cfg.Templates.FunctionProviders)

	// Write the final content to the file
	return utils.WriteStringToFile(filePath, sb.String())
}

/* ------------------------------------------------------------------------- */
/* Utility Helpers                                                           */
/* ------------------------------------------------------------------------- */

// formatFunctionProviders appends the function provider settings to the YAML config.
func formatFunctionProviders(sb *strings.Builder, providers []config.TemplateFuncProvider) {
	sb.WriteString("\n  # List of function providers for template processing.\n")
	sb.WriteString("  function_providers:\n")
	sb.WriteString("    # Example provider using a local path.\n")
	sb.WriteString("    # - name: default\n")
	sb.WriteString("    #   type: path\n")
	sb.WriteString("    #   value: ./providers/default\n")
	sb.WriteString("    #\n")
	sb.WriteString("    # Example provider from a Git repository.\n")
	sb.WriteString("    # - name: custom\n")
	sb.WriteString("    #   type: url\n")
	sb.WriteString("    #   value: https://github.com/user/custom-provider.git\n")

	// Append configured function providers
	if len(providers) > 0 {
		for _, provider := range providers {
			sb.WriteString(fmt.Sprintf("    - name: %s\n", provider.Name))
			sb.WriteString(fmt.Sprintf("      type: %s\n", provider.Type))
			sb.WriteString(fmt.Sprintf("      value: %s\n", provider.Value))
		}
	}
}
