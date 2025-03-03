package loader

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// findProviderFile searches for the `provider.go` file in the module and returns its path and package name.
func findProviderFile(modulePath string) (ProviderMetadata, error) {
	var metadata ProviderMetadata

	err := filepath.Walk(modulePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == "provider.go" {
			metadata.FilePath = path
			metadata.ModuleDir = modulePath

			// Extract the package name from the provider file
			packageName, err := extractPackageName(path)
			if err != nil {
				return fmt.Errorf("failed to determine package name: %s", err)
			}
			metadata.Package = packageName

			return filepath.SkipDir // Stop walking after finding it
		}
		return nil
	})

	if err != nil {
		return metadata, err
	}

	if metadata.FilePath == "" {
		return metadata, fmt.Errorf("missing required provider.go file in the module")
	}

	return metadata, nil
}

func extractPackageName(providerFilePath string) (string, error) {
	file, err := os.Open(providerFilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Ignore comments and blank lines
		if strings.HasPrefix(line, "//") || line == "" {
			continue
		}

		// Extract package name
		words := strings.Fields(line)
		if len(words) > 1 && words[0] == "package" {
			return words[1], nil
		}
	}

	return "", fmt.Errorf("failed to detect package name in provider.go")
}
