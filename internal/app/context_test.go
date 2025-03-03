package app

import (
	"testing"

	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/logger"
)

func TestCliCmdContextInitialization(t *testing.T) {
	// Create mock dependencies
	mockLogger := logger.NewDefaultLogger()
	mockConfig := &config.Config{}
	mockCWD := "/mock/directory"

	// Initialize CLI context
	cliCtx := &AppContext{
		Logger: mockLogger,
		Config: mockConfig,
		CWD:    mockCWD,
	}

	// Check if fields are correctly assigned
	if cliCtx.Logger != mockLogger {
		t.Errorf("Expected Logger to be %v, but got %v", mockLogger, cliCtx.Logger)
	}

	if cliCtx.Config != mockConfig {
		t.Errorf("Expected Config to be %v, but got %v", mockConfig, cliCtx.Config)
	}

	if cliCtx.CWD != mockCWD {
		t.Errorf("Expected CWD to be %v, but got %v", mockCWD, cliCtx.CWD)
	}
}
