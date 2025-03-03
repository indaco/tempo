package generator

// TemplateData represents the data used to populate templates during file generation.
//
// Fields:
// - TemplatesDir: The root directory containing template files.
// - ActionsDir: The root directory containing actions files.
// - GoModule: The name of the Go module being worked on.
// - GoPackage: The Go package name where components will be organized and generated.
// - ComponentName: The name of the component being generated.
// - VariantName: The name of the variant being generated (if applicable).
// - AssetsDir: The directory where asset files (e.g., CSS, JS) will be generated.
// - WithJs: Indicates whether or not JavaScript is required for the component.
// - CssLayer: The name of the CSS layer to associate with component styles.
// - GuardMarker: A text placeholder or sentinel used in template files to mark auto-generated sections.
type TemplateData struct {
	TemplatesDir  string
	ActionsDir    string
	GoModule      string
	GoPackage     string
	ComponentName string
	VariantName   string
	AssetsDir     string
	WithJs        bool
	CssLayer      string
	GuardMarker   string
	WatermarkTip  bool
	Force         bool
	DryRun        bool
}
