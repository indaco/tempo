package resolver

import (
	"testing"
)

func TestResolveString(t *testing.T) {
	tests := []struct {
		name         string
		cliValue     string
		configValue  string
		fieldName    string
		defaultValue string
		allowed      []string
		expected     string
		expectError  bool
	}{
		// Case to cover the missing line
		{
			name:         "CLI and config empty, returns default value",
			cliValue:     "",
			configValue:  "",
			fieldName:    "input directory",
			defaultValue: "./default-input",
			allowed:      nil, // No validation, so default should be used
			expected:     "./default-input",
			expectError:  false,
		},
		{
			name:         "CLI and config empty, no default, should return error",
			cliValue:     "",
			configValue:  "",
			fieldName:    "required-field",
			defaultValue: "",
			allowed:      nil, // No validation
			expected:     "",
			expectError:  true, // Should return an error
		},
		{
			name:         "CLI and config empty, returns default value",
			cliValue:     "",
			configValue:  "",
			fieldName:    "input directory",
			defaultValue: "./default-input",
			allowed:      nil, // No validation, so default should be used
			expected:     "./default-input",
			expectError:  false,
		},
		{
			name:         "CLI and Config both invalid, falls back to default",
			cliValue:     "invalidCLI",
			configValue:  "invalidConfig",
			fieldName:    "summary",
			defaultValue: "long",
			allowed:      []string{"long", "compact", "json", "none"}, // Validation enabled
			expected:     "long",                                      // Should return default
			expectError:  false,
		},
		// Tests with allowed values
		{
			name:         "CLI value provided and valid",
			cliValue:     "compact",
			configValue:  "long",
			fieldName:    "summary",
			defaultValue: "long",
			allowed:      []string{"long", "compact", "json", "none"},
			expected:     "compact",
			expectError:  false,
		},
		{
			name:         "Config value provided and valid",
			cliValue:     "",
			configValue:  "json",
			fieldName:    "summary",
			defaultValue: "long",
			allowed:      []string{"long", "compact", "json", "none"},
			expected:     "json",
			expectError:  false,
		},
		{
			name:         "Both CLI and config values empty",
			cliValue:     "",
			configValue:  "",
			fieldName:    "summary",
			defaultValue: "long",
			allowed:      []string{"long", "compact", "json", "none"},
			expected:     "long",
			expectError:  false,
		},
		{
			name:         "CLI value invalid, falls back to default",
			cliValue:     "invalidValue",
			configValue:  "json",
			fieldName:    "summary",
			defaultValue: "long",
			allowed:      []string{"long", "compact", "json", "none"},
			expected:     "long",
			expectError:  false,
		},
		{
			name:         "Config value invalid, falls back to default",
			cliValue:     "",
			configValue:  "invalidValue",
			fieldName:    "summary",
			defaultValue: "long",
			allowed:      []string{"long", "compact", "json", "none"},
			expected:     "long",
			expectError:  false,
		},
		{
			name:         "CLI and Config both invalid, uses default",
			cliValue:     "invalidCLI",
			configValue:  "invalidConfig",
			fieldName:    "summary",
			defaultValue: "long",
			allowed:      []string{"long", "compact", "json", "none"},
			expected:     "long",
			expectError:  false,
		},
		{
			name:         "CLI value valid, config invalid (uses CLI)",
			cliValue:     "json",
			configValue:  "invalidConfig",
			fieldName:    "summary",
			defaultValue: "long",
			allowed:      []string{"long", "compact", "json", "none"},
			expected:     "json",
			expectError:  false,
		},
		{
			name:         "Empty CLI value with valid config (uses config)",
			cliValue:     "",
			configValue:  "compact",
			fieldName:    "summary",
			defaultValue: "long",
			allowed:      []string{"long", "compact", "json", "none"},
			expected:     "compact",
			expectError:  false,
		},
		{
			name:         "Allowed values list is empty (accepts any value)",
			cliValue:     "json",
			configValue:  "compact",
			fieldName:    "summary",
			defaultValue: "long",
			allowed:      []string{}, // No validation → should accept any value
			expected:     "json",     // Should return "json" instead of "long"
			expectError:  false,
		},
		// Tests with no validation (e.g., input directory)
		{
			name:         "CLI value provided with no validation",
			cliValue:     "/home/user/data",
			configValue:  "",
			fieldName:    "input directory",
			defaultValue: "./default-input",
			allowed:      nil, // No validation → accept any string
			expected:     "/home/user/data",
			expectError:  false,
		},
		{
			name:         "Config value provided with no validation",
			cliValue:     "",
			configValue:  "/var/logs",
			fieldName:    "log directory",
			defaultValue: "./logs",
			allowed:      nil, // No validation → accept any string
			expected:     "/var/logs",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveString(tt.cliValue, tt.configValue, tt.fieldName, tt.defaultValue, tt.allowed)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Unexpected result. Got: %q, Want: %q", result, tt.expected)
				}
			}
		})
	}
}

func TestResolveInt(t *testing.T) {
	tests := []struct {
		name        string
		cliValue    string
		configValue int
		fieldName   string
		expected    int
		expectError bool
	}{
		{
			name:        "Valid CLI integer",
			cliValue:    "10",
			configValue: 5,
			fieldName:   "TestIntField",
			expected:    10,
			expectError: false,
		},
		{
			name:        "Empty CLI value, valid config integer",
			cliValue:    "",
			configValue: 5,
			fieldName:   "TestIntField",
			expected:    5,
			expectError: false,
		},
		{
			name:        "Empty CLI value, zero config integer",
			cliValue:    "",
			configValue: 0,
			fieldName:   "TestIntField",
			expected:    0,
			expectError: true,
		},
		{
			name:        "Invalid CLI value (non-integer)",
			cliValue:    "invalid",
			configValue: 5,
			fieldName:   "TestIntField",
			expected:    0,
			expectError: true,
		},
		{
			name:        "CLI value is zero, valid config integer",
			cliValue:    "0",
			configValue: 10,
			fieldName:   "TestIntField",
			expected:    0,
			expectError: false,
		},
		{
			name:        "CLI value is negative integer",
			cliValue:    "-3",
			configValue: 10,
			fieldName:   "TestIntField",
			expected:    -3,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveInt(tt.cliValue, tt.configValue, tt.fieldName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Unexpected result. Got: %d, Want: %d", result, tt.expected)
				}
			}
		})
	}
}

func TestResolveBool(t *testing.T) {
	tests := []struct {
		name        string
		cliValue    bool
		configValue bool
		expected    bool
	}{
		{
			name:        "CLI value true, config value false",
			cliValue:    true,
			configValue: false,
			expected:    true,
		},
		{
			name:        "CLI value false, config value true",
			cliValue:    false,
			configValue: true,
			expected:    true,
		},
		{
			name:        "Both values false",
			cliValue:    false,
			configValue: false,
			expected:    false,
		},
		{
			name:        "Both values true",
			cliValue:    true,
			configValue: true,
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveBool(tt.cliValue, tt.configValue)

			if result != tt.expected {
				t.Errorf("Unexpected result. Got: %v, Want: %v", result, tt.expected)
			}
		})
	}
}
