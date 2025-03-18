package componentcmd

import (
	"testing"

	"github.com/indaco/tempo/internal/app"
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/logger"
)

// Test Setup Component Define SubCommand
func TestSetupComponentCommand(t *testing.T) {
	cliCtx := &app.AppContext{
		Logger: logger.NewDefaultLogger(),
		Config: config.DefaultConfig(),
		CWD:    t.TempDir(),
	}

	command := SetupComponentCommand(cliCtx)
	if command == nil {
		t.Fatal("SetupComponentCommand returned nil")
	}

	// Check Command Name
	if command.Name != "component" {
		t.Errorf("Expected command name 'component', got '%s'", command.Name)
	}

	// Check Command Description
	expectedUsage := "Define reusable component templates and generate instances from them"
	if command.Usage != expectedUsage {
		t.Errorf("Expected description '%s', got '%s'", expectedUsage, command.Description)
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
