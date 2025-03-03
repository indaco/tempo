# gonameprovider

## Available Template Functions

| Function Name        | Template Function Name | Description                                                                                                |
| :------------------- | :--------------------- | :--------------------------------------------------------------------------------------------------------- |
| `ToGoPackageName`    | `goPackageName`        | Converts a string into a valid Go package name. Handles kebab-case, snake_case, camelCase, and PascalCase. |
| `ToGoExportedName`   | `goExportedName`       | Converts a string to a valid exported Go function name (PascalCase).                                       |
| `ToGoUnexportedName` | `goUnexportedName`     | Converts a string to a valid unexported Go function name (camelCase).                                      |
