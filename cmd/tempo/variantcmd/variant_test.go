package variantcmd

import (
	"testing"

	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/logger"
)

func TestSetupVariantCommand(t *testing.T) {
	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: config.DefaultConfig(),
		CWD:    t.TempDir(),
	}

	command := SetupVariantCommand(cliCtx)
	if command == nil {
		t.Fatal("SetupVariantCommand returned nil")
	}

	// Check Command Name
	if command.Name != "variant" {
		t.Errorf("Expected command name 'variant', got '%s'", command.Name)
	}

	// Check Command Description
	expectedDesc := "Define component variant templates and create instances based on them"
	if command.Description != expectedDesc {
		t.Errorf("Expected description '%s', got '%s'", expectedDesc, command.Description)
	}

	// Check Subcommands Exist
	subcommands := map[string]bool{"define": false, "new": false}
	for _, sub := range command.Commands {
		if _, exists := subcommands[sub.Name]; exists {
			subcommands[sub.Name] = true
		}
	}

	for name, found := range subcommands {
		if !found {
			t.Errorf("Expected subcommand '%s' to be present", name)
		}
	}
}
