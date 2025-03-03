package generator

// BuildVariantActions generates the list of actions required to scaffold a new variant
// for an existing component.
func BuildVariantActions(actionType string, force bool) ([]Action, error) {
	actions := []Action{
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

	if force {
		for i := range actions {
			actions[i].Force = true
		}
	}

	return actions, nil
}
