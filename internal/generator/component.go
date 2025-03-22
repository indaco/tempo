package generator

// BuildComponentActions generates the list of actions required to scaffold a new component.
func BuildComponentActions(actionType string, force, withJs bool) ([]Action, error) {
	actions := []Action{
		// [Templ] - Main component
		{
			Type:         actionType,
			Item:         "file",
			TemplateFile: "component/templ/component.templ.gotxt",
			Path:         "{{ .GoPackage }}/{{ .ComponentName | goPackageName }}/{{ .ComponentName | goPackageName }}.templ",
		},

		// [CSS] - Asset base
		{
			Type:         actionType,
			Item:         "file",
			TemplateFile: "component/assets/css/base.css.gotxt",
			Path:         "{{ .AssetsDir }}/{{ .ComponentName | goPackageName }}/css/base.css",
		},

		// [CSS] - Templ base
		{
			Type:         actionType,
			Item:         "file",
			TemplateFile: "component/templ/css/base-css.templ.gotxt",
			Path:         "{{ .GoPackage }}/{{ .ComponentName | goPackageName }}/css/base.templ",
		},

		// [CSS] - Asset themes folder
		{
			Type:        actionType,
			Item:        "folder",
			Source:      "component/assets/css/themes",
			Destination: "{{ .AssetsDir }}/{{ .ComponentName | goPackageName }}/css/themes",
		},

		// [CSS] - Templ themes folder
		{
			Type:        actionType,
			Item:        "folder",
			Source:      "component/templ/css/themes",
			Destination: "{{ .GoPackage }}/{{ .ComponentName | goPackageName }}/css/themes",
		},
	}

	// Apply force if requested
	if force {
		for i := range actions {
			actions[i].Force = true
		}
	}

	// [JS] - Optional files
	if withJs {
		actions = append(actions,
			// [JS] - Asset
			Action{
				Type:         actionType,
				Item:         "file",
				TemplateFile: "component/assets/js/script.js.gotxt",
				Path:         "{{ .AssetsDir }}/{{ .ComponentName | goPackageName }}/js/script.js",
				OnlyIfJs:     withJs,
			},
			// [JS] - Templ
			Action{
				Type:         actionType,
				Item:         "file",
				TemplateFile: "component/templ/js/script.templ.gotxt",
				Path:         "{{ .GoPackage }}/{{ .ComponentName | goPackageName }}/js/script.templ",
				OnlyIfJs:     withJs,
			},
		)
	}

	return actions, nil
}
