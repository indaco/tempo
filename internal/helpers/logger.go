package helpers

import (
	"github.com/indaco/tempo/internal/logger"
)

// EnableLoggerIndentation ensures consistent indentation.
func EnableLoggerIndentation(log logger.LoggerInterface) {
	log.WithIndent(true)
}

func ResetLogger(log logger.LoggerInterface) {
	log.Reset()
}
