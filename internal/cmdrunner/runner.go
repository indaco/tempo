package cmdrunner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/indaco/tempo/internal/validation"
)

// RunCommandWithTimeout executes a command with a specified timeout.
// It validates the directory to prevent command execution in unsafe locations.
func RunCommandWithTimeout(dir string, timeout time.Duration, command string, args ...string) error {
	// Validate directory to prevent command injection
	if err := validation.ValidateDirectory(dir); err != nil {
		return fmt.Errorf("invalid directory: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("command timed out after %v: %w", timeout, err)
		}
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

// RunCommand executes a command with a default timeout of 30 seconds.
func RunCommand(dir string, command string, args ...string) error {
	return RunCommandWithTimeout(dir, 30*time.Second, command, args...)
}

// RunCommandOutput executes a command and returns its output while enforcing a timeout.
// It validates the directory to prevent command execution in unsafe locations.
func RunCommandOutput(dir string, command string, args ...string) (string, error) {
	// Validate directory to prevent command injection
	if err := validation.ValidateDirectory(dir); err != nil {
		return "", fmt.Errorf("invalid directory: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("command timed out: %w", err)
	}

	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}
