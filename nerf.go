package nerf

import (
	"go.uber.org/zap/zapcore"
)

// StringToLogLevel convert loglevel string into zapCore.Level enum
func StringToLogLevel(level string) zapcore.Level {
	switch string(level) {
	case "debug", "DEBUG":
		return zapcore.DebugLevel
	case "info", "INFO":
		return zapcore.InfoLevel
	case "warn", "WARN":
		return zapcore.WarnLevel
	case "error", "ERROR":
		return zapcore.ErrorLevel
	}

	return zapcore.InfoLevel
}
