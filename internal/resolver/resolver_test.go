package resolver

import (
	"testing"
)

func TestResolveString(t *testing.T) {
	tests := []struct {
		name        string
		cliValue    string
		configValue string
		fieldName   string
		expected    string
		expectError bool
	}{
		{
			name:        "CLI value provided",
			cliValue:    "cliValue",
			configValue: "configValue",
			fieldName:   "TestField",
			expected:    "cliValue",
			expectError: false,
		},
		{
			name:        "Config value provided",
			cliValue:    "",
			configValue: "configValue",
			fieldName:   "TestField",
			expected:    "configValue",
			expectError: false,
		},
		{
			name:        "Both values empty",
			cliValue:    "",
			configValue: "",
			fieldName:   "TestField",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Empty CLI value with non-empty config value",
			cliValue:    "",
			configValue: "fallbackValue",
			fieldName:   "AnotherField",
			expected:    "fallbackValue",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveString(tt.cliValue, tt.configValue, tt.fieldName)

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
