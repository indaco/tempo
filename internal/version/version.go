package version //nolint:revive // package name matches the domain concept, not stdlib

import (
	_ "embed"
	"strings"
)

//go:embed .version
var version string

func GetVersion() string {
	return strings.TrimSpace(version)
}
