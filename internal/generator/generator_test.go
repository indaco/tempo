package generator

import (
	"errors"
	"strings"
	"testing"

	"github.com/indaco/tempo/internal/logger"
	"github.com/indaco/tempo/internal/testutils"
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

// TestHandleDryRun_File tests the "file" branch of handleDryRun using testutils.MockLogger.
func TestHandleDryRun_File(t *testing.T) {
	action := Action{
		Item:         "file",
		Path:         "output-{{.ComponentName}}.txt",
		TemplateFile: "template-{{.ComponentName}}.txt",
	}
	data := &TemplateData{
		ComponentName: "World",
	}
	mockLog := &testutils.MockLogger{}

	handleDryRun(mockLog, action, data)

	if len(mockLog.Logs) == 0 {
		t.Fatalf("Expected an info message, got none")
	}
	msg := mockLog.Logs[0]
	if !strings.Contains(msg, "output-World.txt") {
		t.Errorf("Expected resolved output path 'output-World.txt' in message, got: %s", msg)
	}
	if !strings.Contains(msg, "template-World.txt") {
		t.Errorf("Expected resolved template 'template-World.txt' in message, got: %s", msg)
	}
	if !strings.Contains(msg, "Dry Run: Would execute action:") {
		t.Errorf("Expected message to mention dry-run action, got: %s", msg)
	}
}

// TestHandleDryRun_Folder tests the "folder" branch of handleDryRun using testutils.MockLogger.
func TestHandleDryRun_Folder(t *testing.T) {
	action := Action{
		Item:        "folder",
		Source:      "base-{{.ComponentName}}",
		Destination: "dest-{{.ComponentName}}",
	}
	data := &TemplateData{
		ComponentName: "World",
	}
	mockLog := &testutils.MockLogger{}

	handleDryRun(mockLog, action, data)

	if len(mockLog.Logs) == 0 {
		t.Fatalf("Expected an info message, got none")
	}
	msg := mockLog.Logs[0]
	// Rather than checking an exact formatted string, we verify that the resolved values appear.
	if !strings.Contains(msg, "base-World") {
		t.Errorf("Expected output to contain 'base-World', got: %s", msg)
	}
	if !strings.Contains(msg, "dest-World") {
		t.Errorf("Expected output to contain 'dest-World', got: %s", msg)
	}
	if !strings.Contains(msg, "Dry Run: Would execute action") {
		t.Errorf("Expected message to mention dry-run action, got: %s", msg)
	}
}

// TestHandleDryRun_Unknown tests the default branch of handleDryRun with an unknown action type.
func TestHandleDryRun_Unknown(t *testing.T) {
	action := Action{
		Item: "unknown",
	}
	data := &TemplateData{}
	mockLog := &testutils.MockLogger{}

	handleDryRun(mockLog, action, data)

	if len(mockLog.Logs) == 0 {
		t.Fatalf("Expected a warning message, got none")
	}
	msg := mockLog.Logs[0]
	if !strings.Contains(msg, "Dry Run: Unknown action type") {
		t.Errorf("Expected warning message to mention unknown action type, got: %s", msg)
	}
}
