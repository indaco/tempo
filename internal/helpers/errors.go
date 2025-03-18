package helpers

import (
	"errors"
	"fmt"
	"sort"
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

	// Collect keys and sort them
	keys := make([]string, 0, len(missingFolders))
	for name := range missingFolders {
		keys = append(keys, name)
	}
	sort.Strings(keys) // Ensure consistent ordering

	// Append each missing folder entry in sorted order
	for _, name := range keys {
		fmt.Fprintf(&sb, "  - %s: %s\n", textprovider.SnakeToTitle(name), missingFolders[name])
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
