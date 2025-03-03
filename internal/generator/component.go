package generator

// BuildComponentActions generates the list of actions required to scaffold a new component.
func BuildComponentActions(actionType string, force, withJs bool) ([]Action, error) {
	actions := []Action{
		{
			Type:         actionType,
			Item:         "file",
			TemplateFile: "component/templ/component.templ.gotxt",
			Path:         "{{ .GoPackage }}/{{ .ComponentName | goPackageName }}/{{ .ComponentName | goPackageName }}.templ",
		},
		{
			Type:         actionType,
			Item:         "file",
			TemplateFile: "component/templ/css/base-css.templ.gotxt",
			Path:         "{{ .GoPackage }}/{{ .ComponentName | goPackageName }}/css/base.templ",
		},
		{
			Type:        actionType,
			Item:        "folder",
			Source:      "component/templ/css/themes",
			Destination: "{{ .GoPackage }}/{{ .ComponentName | goPackageName }}/css/themes",
		},

		// CSS files
		{
			Type:         actionType,
			Item:         "file",
			TemplateFile: "component/assets/css/base.css.gotxt",
			Path:         "{{ .AssetsDir }}/{{ .ComponentName | goPackageName }}/css/base.css",
		},
		{
			Type:        actionType,
			Item:        "folder",
			Source:      "component/assets/css/themes",
			Destination: "{{ .AssetsDir }}/{{ .ComponentName | goPackageName }}/css/themes",
		},
	}

	if force {
		for i := range actions {
			actions[i].Force = true
		}
	}

	// Add the extra action if requested
	if withJs {
		actions = append(actions, Action{
			Type:         actionType,
			Item:         "file",
			TemplateFile: "component/assets/js/script.js.gotxt",
			Path:         "{{ .AssetsDir }}/{{ .ComponentName | goPackageName }}/js/script.js",
		})
	}

	return actions, nil
}
