package helpers

import (
	"fmt"
	"strings"
)

// BuildMissingFoldersError constructs a formatted error message for missing folders.
func BuildMissingFoldersError(missingFolders []string, contextMsg string, helpCommands []string) error {
	helpText := ""
	if len(helpCommands) > 0 {
		helpText = "\n\n💡 Need help? Run:\n  - " + strings.Join(helpCommands, "\n  - ")
	}

	return fmt.Errorf(
		`oops! It looks like some required folders are missing.

%s

Missing folders:
%s%s`, contextMsg, strings.Join(missingFolders, "\n"), helpText)
}
