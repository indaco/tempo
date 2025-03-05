package testhelpers

import (
	"strings"
	"testing"
)

func ValidateCLIOutput(t *testing.T, output string, expectedMessages []string) {
	for _, msg := range expectedMessages {
		if !strings.Contains(output, msg) {
			t.Errorf("Expected message not found in output: %s\nOutput: %s", msg, output)
		}
	}
}
