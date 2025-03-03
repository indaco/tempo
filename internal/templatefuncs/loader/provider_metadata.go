package loader

// ProviderMetadata holds information about a provider file.
type ProviderMetadata struct {
	FilePath  string // Path to provider.go
	ModuleDir string // Root path of the provider module
	Package   string // The package name containing the Provider variable
}
