package testutils

import (
	"context"
	"testing"

	"github.com/indaco/tempo/internal/testhelpers"
	"github.com/urfave/cli/v3"
)

func SetupDefineComponent(app *cli.Command, t *testing.T) (string, error) {
	output, err := testhelpers.CaptureStdout(func() {
		args := []string{"tempo", "define", "component"}
		if err := app.Run(context.Background(), args); err != nil {
			t.Fatalf("Failed to set up component structure with 'define component': %v", err)
		}
	})

	return output, err
}

func SetupDefineVariant(app *cli.Command, t *testing.T) (string, error) {
	output, err := testhelpers.CaptureStdout(func() {
		args := []string{"tempo", "define", "variant"}
		if err := app.Run(context.Background(), args); err != nil {
			t.Fatalf("Failed to set up component variant structure with 'define variant': %v", err)
		}
	})

	return output, err
}
