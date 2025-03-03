package generator

import (
	"reflect"
	"testing"
)

func TestBuildComponentActions(t *testing.T) {
	actionType := "render"
	withJs := true

	t.Run("Force", func(t *testing.T) {
		actions, err := BuildComponentActions(actionType, true, false)
		if err != nil {
			t.Fatalf("BuildComponentActions() returned an error: %v", err)
		}

		// Expected actions count (including JS file)
		expectedCount := 5

		if len(actions) != expectedCount {
			t.Errorf("BuildComponentActions() = %d actions; want %d actions", len(actions), expectedCount)
		}

		// Verify the last action is the JS action
		lastAction := actions[0]
		expectedLastAction := Action{
			Type:         actionType,
			Item:         "file",
			TemplateFile: "component/templ/component.templ.gotxt",
			Path:         "{{ .GoPackage }}/{{ .ComponentName | goPackageName }}/{{ .ComponentName | goPackageName }}.templ",
			Force:        true,
		}

		if !reflect.DeepEqual(lastAction, expectedLastAction) {
			t.Errorf("Last action = %v; want %v", lastAction, expectedLastAction)
		}
	})

	t.Run("WithoutJS", func(t *testing.T) {
		actions, err := BuildComponentActions(actionType, false, false)
		if err != nil {
			t.Fatalf("BuildComponentActions() returned an error: %v", err)
		}

		// Expected actions count (excluding JS file)
		expectedCount := 5

		if len(actions) != expectedCount {
			t.Errorf("BuildComponentActions() = %d actions; want %d actions", len(actions), expectedCount)
		}

		// Verify that no JS action is included
		for _, action := range actions {
			if action.TemplateFile == "component/assets/js/script.js.gotxt" {
				t.Errorf("Unexpected JS action found: %v", action)
			}
		}
	})

	t.Run("WithJS", func(t *testing.T) {
		actions, err := BuildComponentActions(actionType, false, withJs)
		if err != nil {
			t.Fatalf("BuildComponentActions() returned an error: %v", err)
		}

		// Expected actions count (including JS file)
		expectedCount := 6

		if len(actions) != expectedCount {
			t.Errorf("BuildComponentActions() = %d actions; want %d actions", len(actions), expectedCount)
		}

		// Verify the last action is the JS action
		lastAction := actions[len(actions)-1]
		expectedLastAction := Action{
			Type:         actionType,
			Item:         "file",
			TemplateFile: "component/assets/js/script.js.gotxt",
			Path:         "{{ .AssetsDir }}/{{ .ComponentName | goPackageName }}/js/script.js",
		}

		if !reflect.DeepEqual(lastAction, expectedLastAction) {
			t.Errorf("Last action = %v; want %v", lastAction, expectedLastAction)
		}
	})

	t.Run("ActionType", func(t *testing.T) {
		actions, err := BuildComponentActions("copy", false, false)
		if err != nil {
			t.Fatalf("BuildComponentActions() returned an error: %v", err)
		}

		// Verify that all actions have the correct action type
		for _, action := range actions {
			if action.Type != "copy" {
				t.Errorf("Action.Type = %s; want 'copy'", action.Type)
			}
		}
	})
}
