// Package helpers provides CLI-specific helper functions for Tempo commands.
//
// This package contains utilities that are specifically designed for CLI command
// implementations and should only be imported by packages in cmd/tempo/*.
//
// # Logger Helpers (logger.go)
//
// Functions for managing logger state in CLI commands:
//   - EnableLoggerIndentation - Enable indented output for nested logging
//   - ResetLogger - Reset logger to default state after command completion
//   - LogSuccessMessages - Log standardized success messages for entity creation
//
// # Error Builders (errors.go)
//
// Functions for building user-friendly error messages:
//   - BuildMissingFoldersError - Build error message for missing folder validation
//
// # Entity Helpers (entity.go)
//
// Functions for handling entity existence checks in CLI commands:
//   - CheckEntityForNew - Log warning/info when creating entities that exist
//   - CheckEntityForDefine - Log warning/info when defining templates that exist
//
// # Usage
//
// These helpers are designed to be used in CLI command implementations:
//
//	func runCommand(cmdCtx *app.AppContext) func(ctx context.Context, cmd *cli.Command) error {
//	    return func(ctx context.Context, cmd *cli.Command) error {
//	        helpers.EnableLoggerIndentation(cmdCtx.Logger)
//	        defer helpers.ResetLogger(cmdCtx.Logger)
//
//	        // Command implementation...
//
//	        return nil
//	    }
//	}
package helpers
