package nerf

import (
	"go.uber.org/zap/zapcore"
)

// OauthOrganization compile-time derived from -X github.com/ton31337/nerf.OauthOrganization
// E.g.: example which will be used to retrieve teams by username from GitHub in this org.
var OauthOrganization string

// DNSAutoDiscoverZone compile-time derived from -Xgithub.com/ton31337/nerf.DNSAutoDiscoverZone
// E.g.: example.com which will be combined to _vpn._udp.example.com SRV query
var DNSAutoDiscoverZone string

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
