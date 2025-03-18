package helpers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/indaco/tempo/internal/templatefuncs/providers/textprovider"
)

// BuildMissingFoldersError constructs an error message for missing folders.
func BuildMissingFoldersError(missingFolders map[string]string, contextMsg string, helpCommands []string) error {
	if len(missingFolders) == 0 {
		return nil
	}

	var sb strings.Builder

	sb.WriteString("oops! It looks like some required folders are missing.\n\n")
	sb.WriteString(contextMsg)
	sb.WriteString("\n\nMissing folders:\n")

	// Append each missing folder entry
	for name, path := range missingFolders {
		fmt.Fprintf(&sb, "  - %s: %s\n", textprovider.SnakeToTitle(name), path)
	}

	// Append help commands if available
	if len(helpCommands) > 0 {
		sb.WriteString("\n💡 Need help? Run:\n")
		for _, cmd := range helpCommands {
			fmt.Fprintf(&sb, "  - %s\n", cmd)
		}
	}

	return errors.New(strings.TrimSpace(sb.String()))
}
