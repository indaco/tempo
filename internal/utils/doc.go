// Package utils provides shared utility functions for the Tempo CLI.
//
// This package contains utilities organized by domain:
//
// # File System Operations (fs.go)
//
// Functions for file and directory management:
//   - FileOrDirExists, FileExists, DirExists - Check existence
//   - EnsureDirExists, RemoveIfExists - Create/remove
//   - ReadFileAsString, WriteToFile, WriteStringToFile, WriteJSONToFile - Read/write
//   - CheckMissingFolders, ValidateFoldersExistence - Validation
//   - FileSystemOperations interface for dependency injection
//
// # Path Utilities (fs.go)
//
// Functions for path manipulation:
//   - GetCWD, ResolvePath - Current directory and path resolution
//   - ToTemplFilename, RebasePathToOutput - Template path conversion
//   - RemoveTemplatingExtension - Extension handling
//   - GetModuleName - Go module detection
//
// # Embedded Resources (embed.go)
//
// Functions for working with embedded filesystems:
//   - IsEmbedded - Check if path is embedded
//   - ReadEmbeddedFile, ReadEmbeddedDir - Read embedded content
//   - CopyDirFromEmbed, CopyFileFromEmbed - Copy embedded resources
//
// # Template Rendering (renderer.go)
//
// Functions for template processing:
//   - RenderTemplate - Render Go templates with custom functions
//
// # String Utilities (strings.go)
//
// Functions for string manipulation:
//   - ContainsSubstring - Case-insensitive substring check
//   - ExtractNameFromURL, ExtractNameFromPath - Name extraction
//
// # Type Conversion (numbers.go)
//
// Functions for type conversion:
//   - Int64ToInt - Safe int64 to int conversion
package utils //nolint:revive // package provides utility functions
