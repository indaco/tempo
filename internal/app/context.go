package app

import (
	"github.com/indaco/tempo/internal/config"
	"github.com/indaco/tempo/internal/logger"
)

type AppContext struct {
	Logger logger.LoggerInterface
	Config *config.Config
	CWD    string
}
