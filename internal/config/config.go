package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/indaco/tempo/internal/errors"
	"gopkg.in/yaml.v3"
)

/* ------------------------------------------------------------------------- */
/* TYPES & CONSTANTS                                                         */
/* ------------------------------------------------------------------------- */

// App contains application-specific settings.
type App struct {
	GoModule  string `yaml:"go_module,omitempty"`
	GoPackage string `yaml:"go_package,omitempty"`
	WithJs    bool   `yaml:"with_js,omitempty"`
	CssLayer  string `yaml:"css_layer,omitempty"`
	AssetsDir string `yaml:"assets_dir,omitempty"`
}

// Paths defines paths used in the application.
type Paths struct {
	TemplatesDir string `yaml:"-"`
	ActionsDir   string `yaml:"-"`
}

// Processor defines settings for the files processing.
type Processor struct {
	Workers       int    `yaml:"workers"`
	SummaryFormat string `yaml:"summary_format"`
}

// TemplateFuncProvider represents a function provider that can be loaded from a local path or a remote URL.
type TemplateFuncProvider struct {
	Name  string `yaml:"name,omitempty"`
	Type  string `yaml:"type,omitempty"` // "path" or "url"
	Value string `yaml:"value,omitempty"`
}

// Templates defines settings related to template files and processing.
type Templates struct {
	Extensions        []string               `yaml:"extensions,omitempty"`
	GuardMarker       string                 `yaml:"guard_marker,omitempty"`
	WatermarkTip      bool                   `yaml:"watermark,omitempty"`
	FunctionProviders []TemplateFuncProvider `yaml:"function_providers,omitempty"`
}

// Config represents the configuration settings for the application.
type Config struct {
	TempoRoot string    `yaml:"tempo_root"`
	App       App       `yaml:"app,omitempty"`
	Paths     Paths     `yaml:"-"`
	Processor Processor `yaml:"processor,omitempty"`
	Templates Templates `yaml:"templates,omitempty"`
}

// Default values for the configuration.
const (
	DefaultBaseDir          = ".tempo-files"
	DefaultGoPackage        = "components"
	DefaultAssetsDir        = "assets"
	DefaultCssLayer         = "components"
	DefaultSummaryFormat    = "long"
	DefaultGuardMarkText    = "tempo"
	DefaultWatermarkTipFlag = true
)

var (
	DefaultNumWorkers         = runtime.NumCPU() * 2
	TempoConfigFiles          = []string{"tempo.yaml", "tempo.yml"}
	DefaultTemplateExtensions = []string{".gotxt", ".gotmpl", ".tpl"}
)

/* ------------------------------------------------------------------------- */
/* DEFAULT CONFIG & LOADING                                                  */
/* ------------------------------------------------------------------------- */

// DefaultConfig generates a default configuration with the paths dynamically updated
// based on DefaultBaseFolder.
func DefaultConfig() *Config {
	templatesDir, actionsDir := DerivedFolderPaths(DefaultBaseDir)

	return &Config{
		TempoRoot: DefaultBaseDir,
		App: App{
			GoModule:  "github.com/example/demotempo",
			GoPackage: DefaultGoPackage,
			WithJs:    false,
			CssLayer:  DefaultCssLayer,
			AssetsDir: DefaultAssetsDir,
		},
		Paths: Paths{
			TemplatesDir: templatesDir,
			ActionsDir:   actionsDir,
		},
		Processor: Processor{
			Workers:       DefaultNumWorkers,
			SummaryFormat: DefaultSummaryFormat,
		},
		Templates: Templates{
			GuardMarker:       DefaultGuardMarkText,
			WatermarkTip:      DefaultWatermarkTipFlag,
			Extensions:        DefaultTemplateExtensions,
			FunctionProviders: []TemplateFuncProvider{},
		},
	}
}

// LoadConfig loads the application configuration from a file or uses default values.
func LoadConfig() (*Config, error) {
	defaultConfig := DefaultConfig()

	for _, file := range TempoConfigFiles {
		if _, err := os.Stat(file); err == nil {
			data, err := os.ReadFile(file)
			if err != nil {
				return nil, errors.Wrap("failed to read config file:", err, file)
			}

			var fileConfig Config
			if err := yaml.Unmarshal(data, &fileConfig); err != nil {
				return nil, errors.Wrap("failed to parse config file:", err, file)
			}

			return ensureDefaults(defaultConfig, &fileConfig), nil
		}
	}

	return defaultConfig, nil
}

// DerivedFolderPaths returns the derived folder paths based on the base folder.
func DerivedFolderPaths(baseFolder string) (TemplatesDir, ActionsDir string) {
	TemplatesDir = filepath.Join(baseFolder, "templates")
	ActionsDir = filepath.Join(baseFolder, "actions")
	return
}

/* ------------------------------------------------------------------------- */
/* UTILITY HELPERS                                                           */
/* ------------------------------------------------------------------------- */

// ensureDefaults merges a partial configuration (from a file) with the default configuration.
//
// This function updates the default configuration with any non-empty values
// provided in the fileConfig.
func ensureDefaults(defaultConfig, fileConfig *Config) *Config {
	if fileConfig.TempoRoot != "" {
		defaultConfig.TempoRoot = fileConfig.TempoRoot
		defaultConfig.Paths.TemplatesDir = filepath.Join(fileConfig.TempoRoot, "templates")
		defaultConfig.Paths.ActionsDir = filepath.Join(fileConfig.TempoRoot, "actions")
	}

	if fileConfig.App.GoModule != "" {
		defaultConfig.App.GoModule = fileConfig.App.GoModule
	}
	if fileConfig.App.GoPackage != "" {
		defaultConfig.App.GoPackage = fileConfig.App.GoPackage
	}
	if fileConfig.App.WithJs {
		defaultConfig.App.WithJs = fileConfig.App.WithJs
	}
	if fileConfig.App.CssLayer != "" {
		defaultConfig.App.CssLayer = fileConfig.App.CssLayer
	}
	if fileConfig.App.AssetsDir != "" {
		defaultConfig.App.AssetsDir = fileConfig.App.AssetsDir
	}

	if fileConfig.Processor.Workers != 0 {
		defaultConfig.Processor.Workers = fileConfig.Processor.Workers
	}
	if fileConfig.Processor.SummaryFormat != "" {
		defaultConfig.Processor.SummaryFormat = fileConfig.Processor.SummaryFormat
	}

	if len(fileConfig.Templates.Extensions) > 0 {
		defaultConfig.Templates.Extensions = fileConfig.Templates.Extensions
	}
	if fileConfig.Templates.GuardMarker != "" {
		defaultConfig.Templates.GuardMarker = fileConfig.Templates.GuardMarker
	}
	if fileConfig.Templates.FunctionProviders != nil {
		defaultConfig.Templates.FunctionProviders = fileConfig.Templates.FunctionProviders
	} else {
		defaultConfig.Templates.FunctionProviders = []TemplateFuncProvider{} // Ensure it's an empty slice, not nil
	}

	return defaultConfig
}
