package loader

import (
	"fmt"

	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/templatefuncs/registry"
)

// RegisterFunctionsFromPath registers functions from a local Go package.
func RegisterFunctionsFromPath(packagePath string, logger logger.LoggerInterface) error {
	logger.Info("Loading functions from").WithAttrs("package", packagePath)

	// Step 1: Locate the provider.go file dynamically
	providerMetadata, err := findProviderFile(packagePath)
	if err != nil {
		return fmt.Errorf("invalid function provider: %s", err)
	}

	// Step 2: Extract provider functions
	provider, err := loadDynamicProvider(providerMetadata, logger)
	if err != nil {
		return fmt.Errorf("failed to import package: %s", err)
	}

	// Step 3: Register functions
	functions := provider.GetFunctions()
	if len(functions) == 0 {
		return fmt.Errorf("invalid function provider: No functions found in Provider")
	}

	for name, fn := range functions {
		logger.Success("Registered function", name)
		registry.RegisterFunction(name, fn)
	}

	return nil
}
