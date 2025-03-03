package config

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	expected := &Config{
		TempoRoot: DefaultBaseDir,
		App: App{
			GoModule:  "github.com/example/demotempo",
			GoPackage: "components",
			WithJs:    false,
			CssLayer:  "components",
			AssetsDir: "assets",
		},
		Paths: Paths{
			TemplatesDir: filepath.Join(DefaultBaseDir, "templates"),
			ActionsDir:   filepath.Join(DefaultBaseDir, "actions"),
		},
		Processor: Processor{
			Workers:       runtime.NumCPU() * 2,
			SummaryFormat: "long",
		},
		Templates: Templates{
			Extensions:        DefaultTemplateExtensions,
			WatermarkTip:      true,
			GuardMarker:       "tempo",
			FunctionProviders: []TemplateFuncProvider{},
		},
	}

	actual := DefaultConfig()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("DefaultConfig() = %+v, want %+v", actual, expected)
	}
}

func TestLoadConfig_NoFile(t *testing.T) {
	tempDir := t.TempDir()
	err := os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory to tempDir: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() returned an error: %v", err)
	}

	expected := DefaultConfig()
	if !reflect.DeepEqual(cfg, expected) {
		t.Errorf("LoadConfig() = %+v, want %+v", cfg, expected)
	}
}

func TestLoadConfig_WithFile(t *testing.T) {
	tempDir := t.TempDir()
	err := os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory to tempDir: %v", err)
	}

	configFile := "tempo.yaml"
	customConfig := Config{
		TempoRoot: "custom-tempo",
		App: App{
			GoModule:  "github.com/custom/module",
			GoPackage: "custom_package",
			WithJs:    true,
			CssLayer:  "custom_layer",
			AssetsDir: "custom_assets",
		},
		Paths: Paths{
			TemplatesDir: "custom-tempo/templates",
			ActionsDir:   "custom-tempo/actions",
		},
		Processor: Processor{
			Workers:       8,
			SummaryFormat: "json",
		},
		Templates: Templates{
			Extensions:        DefaultTemplateExtensions,
			WatermarkTip:      true,
			GuardMarker:       "custom-marker",
			FunctionProviders: []TemplateFuncProvider{},
		},
	}

	data, err := yaml.Marshal(&customConfig)
	if err != nil {
		t.Fatalf("Failed to marshal custom config: %v", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() returned an error: %v", err)
	}

	if !reflect.DeepEqual(cfg, &customConfig) {
		t.Errorf("LoadConfig() = %+v, want %+v", cfg, &customConfig)
	}
}

func TestEnsureDefaults(t *testing.T) {
	defaultConfig := DefaultConfig()
	customConfig := &Config{
		TempoRoot: "custom-tempo",
		App: App{
			GoModule: "github.com/custom/module",
		},
		Templates: Templates{
			GuardMarker: "custom-marker",
		},
	}

	updatedConfig := ensureDefaults(defaultConfig, customConfig)

	expected := &Config{
		TempoRoot: "custom-tempo",
		App: App{
			GoModule:  "github.com/custom/module",
			GoPackage: "components",
			WithJs:    false,
			CssLayer:  "components",
			AssetsDir: "assets",
		},
		Paths: Paths{
			TemplatesDir: filepath.Join("custom-tempo", "templates"),
			ActionsDir:   filepath.Join("custom-tempo", "actions"),
		},
		Processor: Processor{
			Workers:       DefaultNumWorkers,
			SummaryFormat: DefaultSummaryFormat,
		},
		Templates: Templates{
			Extensions:        DefaultTemplateExtensions,
			WatermarkTip:      true,
			GuardMarker:       "custom-marker",
			FunctionProviders: []TemplateFuncProvider{},
		},
	}

	if !reflect.DeepEqual(updatedConfig, expected) {
		t.Errorf("ensureDefaults() = %+v, want %+v", updatedConfig, expected)
	}
}

func TestLoadConfig_HybridApproach(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	expectedFile := filepath.Join(tempDir, "tempo.yaml")

	// Write to the most preferred file
	err := os.WriteFile(expectedFile, []byte("app:\n  go_module: custom_module\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Change working directory to tempDir
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Load the configuration
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify that the preferred file was loaded
	if config.App.GoModule != "custom_module" {
		t.Errorf("Expected 'custom_module', got '%s'", config.App.GoModule)
	}
}

func TestLoadConfig_Fallback(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	files := []string{"tempo.yaml", "tempo.yml"}

	// Write to the least preferred file
	lastFile := filepath.Join(tempDir, files[len(files)-1])
	err := os.WriteFile(lastFile, []byte("app:\n  go_module: fallback_module\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write fallback config file: %v", err)
	}

	// Write to a more preferred file
	preferredFile := filepath.Join(tempDir, files[0])
	err = os.WriteFile(preferredFile, []byte("app:\n  go_module: preferred_module\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to write preferred config file: %v", err)
	}

	// Change working directory to tempDir
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Load the configuration
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify that the most preferred file was loaded
	if config.App.GoModule != "preferred_module" {
		t.Errorf("Expected 'preferred_module', got '%s'", config.App.GoModule)
	}
}
