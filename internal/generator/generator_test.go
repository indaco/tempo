package generator

import (
	"errors"
	"testing"

	"github.com/indaco/tempo/internal/logger"
)

// MockActionHandler is a mock implementation of ActionHandler for testing.
type MockActionHandler struct {
	ExecutedActions []Action
	ReturnError     bool
}

func (m *MockActionHandler) Execute(action Action, data *TemplateData) error {
	m.ExecutedActions = append(m.ExecutedActions, action)
	if m.ReturnError {
		return errors.New("mock error")
	}
	return nil
}

func TestProcessActions(t *testing.T) {
	logger := logger.NewDefaultLogger()
	tests := []struct {
		name        string
		actions     []Action
		dryRun      bool
		mockError   bool
		expectCount int
		expectError bool
	}{
		{
			name:        "No actions",
			actions:     []Action{},
			dryRun:      false,
			mockError:   false,
			expectCount: 0,
			expectError: false,
		},
		{
			name: "Single action success",
			actions: []Action{
				{Type: "file", Path: "path/to/file", TemplateFile: "template.gotxt"},
			},
			dryRun:      false,
			mockError:   false,
			expectCount: 1,
			expectError: false,
		},
		{
			name: "Multiple actions success",
			actions: []Action{
				{Type: "file", Path: "path/to/file1", TemplateFile: "template1.gotxt"},
				{Type: "folder", Source: "base/dir", Destination: "dest/dir"},
			},
			dryRun:      false,
			mockError:   false,
			expectCount: 2,
			expectError: false,
		},
		{
			name: "Handler returns error",
			actions: []Action{
				{Type: "file", Path: "path/to/file", TemplateFile: "template.gotxt"},
			},
			dryRun:      false,
			mockError:   true,
			expectCount: 1,
			expectError: true,
		},
		{
			name: "Dry run with actions",
			actions: []Action{
				{Type: "file", Path: "path/to/file", TemplateFile: "template.gotxt"},
			},
			dryRun:      true,
			mockError:   false,
			expectCount: 0,
			expectError: false,
		},
		{
			name: "Unknown action type",
			actions: []Action{
				{Type: "unknown", Path: "path/to/file", TemplateFile: "template.gotxt"},
			},
			dryRun:      false,
			mockError:   false,
			expectCount: 0,
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockHandler := &MockActionHandler{ReturnError: test.mockError}
			actionHandlers = map[string]ActionHandler{
				"file":   mockHandler,
				"folder": mockHandler,
			}

			data := &TemplateData{DryRun: test.dryRun}
			err := ProcessActions(logger, test.actions, data)

			if test.expectError {
				if err == nil {
					t.Fatalf("Expected error but got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
			}

			if test.dryRun {
				if len(mockHandler.ExecutedActions) > 0 {
					t.Fatalf("Dry run should not execute actions")
				}
			} else if !test.expectError {
				if len(mockHandler.ExecutedActions) != test.expectCount {
					t.Fatalf("Expected %d actions, got %d", test.expectCount, len(mockHandler.ExecutedActions))
				}
			}
		})
	}
}
