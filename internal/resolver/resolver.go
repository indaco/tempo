// GoPackage resolver provides helper functions to resolve values by prioritizing
// CLI-provided values over configuration-defined values, ensuring proper validation
// and error handling when necessary
package resolver

import (
	"strconv"

	"github.com/indaco/tempo/internal/errors"
	"github.com/indaco/tempo/internal/templatefuncs/providers/textprovider"
)

// ResolveString determines the final string value by preferring the CLI flag over
// the configuration value. If both are empty, it returns an error indicating that
// the field is required.
//
// Parameters:
//   - cliValue: The value provided via CLI flag (if any).
//   - configValue: The value from the configuration file.
//   - fieldName: The name of the field being resolved (used for error messages).
//
// Returns:
//   - The resolved string value (either cliValue or configValue).
//   - An error if both values are empty, indicating that the field is required.
func ResolveString(cliValue, configValue, fieldName string) (string, error) {
	if textprovider.IsEmptyString(cliValue) && textprovider.IsEmptyString(configValue) {
		return "", errors.Wrap("%s is required. Ensure to pass it as a flag or set its value in .tempo.yaml file", fieldName)
	}

	if !textprovider.IsEmptyString(cliValue) {
		return cliValue, nil
	}

	return configValue, nil
}

// ResolveInt determines the final integer value by prioritizing the CLI flag over
// the configuration value. If both values are zero, it returns an error indicating
// that the field is required.
//
// Parameters:
//   - cliValue: The value provided via CLI flag (if any).
//   - configValue: The integer value from the configuration file.
//   - fieldName: The name of the field being resolved (used for error messages).
//
// Returns:
//   - The resolved integer value (either cliValue or configValue).
//   - An error if both values are zero, indicating that the field is required.
func ResolveInt(cliValue string, configValue int, fieldName string) (int, error) {
	// If the CLI value is provided, attempt to parse it
	if cliValue != "" {
		num, err := strconv.Atoi(cliValue)
		if err != nil {
			return 0, errors.Wrap("invalid value for %s: expected an integer but got '%s'", fieldName, cliValue)
		}
		return num, nil
	}

	// If CLI value is empty, fallback to config
	if configValue == 0 {
		return 0, errors.Wrap("%s is required. Ensure to pass it as a flag or set its value in .tempo.yaml file", fieldName)
	}

	return configValue, nil
}

// ResolveBool determines the final boolean value by preferring the CLI flag over
// the configuration value.
//
// Parameters:
//   - cliValue: The boolean value provided via CLI flag (if any).
//   - configValue: The boolean value from the configuration file.
//
// Returns:
//   - The resolved boolean value, defaulting to configValue if cliValue is false.
func ResolveBool(cliValue, configValue bool) bool {
	if cliValue {
		return cliValue
	}
	return configValue
}
