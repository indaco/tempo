package generator

import (
	"reflect"
	"testing"
)

func TestBuildVariantActions(t *testing.T) {
	actionType := "render"

	t.Run("GenerateVariantActions", func(t *testing.T) {
		actions, err := BuildVariantActions(actionType, false)
		if err != nil {
			t.Fatalf("BuildVariantActions() returned an error: %v", err)
		}

		// Expected actions
		expectedActions := []Action{
			{
				Type:         actionType,
				Item:         "file",
				TemplateFile: "component-variant/name.templ.gotxt",
				Path:         "{{ .GoPackage }}/{{ .ComponentName | goPackageName }}/css/variants/{{ .VariantName | goUnexportedName }}.templ",
			},
			{
				Type:         actionType,
				Item:         "file",
				TemplateFile: "component-variant/assets/css/name.css.gotxt",
				Path:         "{{ .AssetsDir }}/{{ .ComponentName | goPackageName }}/css/variants/{{ .VariantName | goUnexportedName }}.css",
			},
		}

		// Verify the number of actions
		if len(actions) != len(expectedActions) {
			t.Errorf("BuildVariantActions() = %d actions; want %d actions", len(actions), len(expectedActions))
		}

		// Verify each action
		for i, action := range actions {
			if !reflect.DeepEqual(action, expectedActions[i]) {
				t.Errorf("Action[%d] = %v; want %v", i, action, expectedActions[i])
			}
		}
	})

	t.Run("GenerateVariantActions_WithForce", func(t *testing.T) {
		actions, err := BuildVariantActions(actionType, true)
		if err != nil {
			t.Fatalf("BuildVariantActions() returned an error: %v", err)
		}

		// Expected actions
		expectedActions := []Action{
			{
				Type:         actionType,
				Item:         "file",
				TemplateFile: "component-variant/name.templ.gotxt",
				Path:         "{{ .GoPackage }}/{{ .ComponentName | goPackageName }}/css/variants/{{ .VariantName | goUnexportedName }}.templ",
				Force:        true,
			},
			{
				Type:         actionType,
				Item:         "file",
				TemplateFile: "component-variant/assets/css/name.css.gotxt",
				Path:         "{{ .AssetsDir }}/{{ .ComponentName | goPackageName }}/css/variants/{{ .VariantName | goUnexportedName }}.css",
				Force:        true,
			},
		}

		// Verify the number of actions
		if len(actions) != len(expectedActions) {
			t.Errorf("BuildVariantActions() = %d actions; want %d actions", len(actions), len(expectedActions))
		}

		// Verify each action
		for i, action := range actions {
			if !reflect.DeepEqual(action, expectedActions[i]) {
				t.Errorf("Action[%d] = %v; want %v", i, action, expectedActions[i])
			}
		}
	})

	t.Run("ActionTypeValidation", func(t *testing.T) {
		actions, err := BuildVariantActions(CopyActionId, false)
		if err != nil {
			t.Fatalf("BuildVariantActions() returned an error: %v", err)
		}

		// Verify that all actions have the correct action type
		for _, action := range actions {
			if action.Type != "copy" {
				t.Errorf("Action.Type = %s; want 'copy'", action.Type)
			}
		}
	})
}
