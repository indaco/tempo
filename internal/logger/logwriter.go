package logger

import "strings"

type LogWriter struct {
	Level  string
	Logger LoggerInterface
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	message := strings.TrimSpace(string(p))
	if lw.Logger == nil {
		lw.Logger = NewDefaultLogger()
	}

	switch lw.Level {
	case "info":
		lw.Logger.Info(message)
	case "success":
		lw.Logger.Success(message)
	case "warn":
		lw.Logger.Warning(message)
	case "error":
		lw.Logger.Error(message)
	default:
		lw.Logger.Default(message)
	}
	return len(p), nil
}
