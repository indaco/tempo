package testutils

import (
	"context"
	"testing"

	"github.com/urfave/cli/v3"
)

func SetupComponentDefine(app *cli.Command, t *testing.T) (string, error) {
	output, err := CaptureStdout(func() {
		args := []string{"tempo", "component", "define"}
		if err := app.Run(context.Background(), args); err != nil {
			t.Fatalf("Failed to set up component structure with 'component define': %v", err)
		}
	})

	return output, err
}

func SetupVariantDefine(app *cli.Command, t *testing.T) (string, error) {
	output, err := CaptureStdout(func() {
		args := []string{"tempo", "variant", "define"}
		if err := app.Run(context.Background(), args); err != nil {
			t.Fatalf("Failed to set up component variant structure with 'variant define': %v", err)
		}
	})

	return output, err
}
