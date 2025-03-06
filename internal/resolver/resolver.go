// GoPackage resolver provides helper functions to resolve values by prioritizing
// CLI-provided values over configuration-defined values, ensuring proper validation
// and error handling when necessary
package resolver

import (
	"strconv"

	"github.com/indaco/tempo/internal/errors"
	"github.com/indaco/tempo/internal/templatefuncs/providers/textprovider"
)

// ResolveString determines the final string value by prioritizing the CLI flag over
// the configuration value. It validates the resolved value against allowed options
// and falls back to a default if necessary.
//
// If both the CLI and configuration values are empty, an error is returned, indicating
// that the field is required. If an invalid value is provided, the default value
// is silently returned.
//
// Parameters:
//   - cliValue: The value provided via CLI flag (if any).
//   - configValue: The value from the configuration file (if any).
//   - fieldName: The name of the field being resolved (used for error messages).
//   - defaultValue: The fallback value if neither cliValue nor configValue are valid.
//   - allowedValues: A list of permitted values for validation.
//
// Returns:
//   - The resolved string value (cliValue, configValue, or defaultValue).
//   - An error if both values are empty or an invalid value is provided.
func ResolveString(cliValue, configValue, fieldName, defaultValue string, allowedValues []string) (string, error) {
	// If both CLI and config values are empty, return the default if it's provided
	if textprovider.IsEmptyString(cliValue) && textprovider.IsEmptyString(configValue) {
		if textprovider.IsEmptyString(defaultValue) {
			// If no default value exists, return an error
			return "", errors.Wrap("%s is required. Ensure to pass it as a flag or set its value in .tempo.yaml file", fieldName)
		}
	}

	// If allowedValues is empty/nil, we skip validation (used for unrestricted fields like directories)
	shouldValidate := len(allowedValues) > 0

	// Use CLI value if provided
	if !textprovider.IsEmptyString(cliValue) {
		if shouldValidate && !textprovider.IsValidValue(cliValue, allowedValues) {
			return defaultValue, nil // Use default if invalid
		}
		return cliValue, nil
	}

	// Use config value if provided
	if !textprovider.IsEmptyString(configValue) {
		if shouldValidate && !textprovider.IsValidValue(configValue, allowedValues) {
			return defaultValue, nil // Use default if invalid
		}
		return configValue, nil
	}

	// Default fallback
	return defaultValue, nil
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
