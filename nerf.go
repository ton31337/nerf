package nerf

import (
	"context"
	math "math"
	"net"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	grpc "google.golang.org/grpc"
)

// Endpoint struct to store all the relevant data about gRPC server,
// which generates and returns data for Nebula.
type Endpoint struct {
	Description string
	RemoteHost  string
	RemoteIP    string
	Latency     int64
}

func probeEndpoint(remoteHost string) int64 {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	conn, err := grpc.Dial(remoteHost+":9000", grpc.WithInsecure())
	if err != nil {
		Cfg.Logger.Error("failed connecting to gRPC",
			zap.String("RemoteHost", remoteHost),
			zap.Error(err))
	}
	defer conn.Close()

	client := NewServerClient(conn)
	data := start.UnixNano()
	request := &PingRequest{Data: &data, Login: &Cfg.Login}
	response, err := client.Ping(ctx, request)
	if err != nil || *response.Data == 0 {
		return math.MaxInt64
	}

	return time.Since(start).Milliseconds()
}

func getVPNEndpoints() {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 3 * time.Second,
			}
			return d.DialContext(ctx, "udp", "1.1.1.1:53")
		},
	}

	_, srvRecords, err := r.LookupSRV(context.Background(), "vpn", "udp", DNSAutoDiscoverZone)
	if err != nil {
		Cfg.Logger.Fatal("no available gRPC endpoints found (DNS SRV)", zap.Error(err))
	}

	for _, record := range srvRecords {
		txtRecords, err := r.LookupTXT(context.Background(), record.Target)
		if err != nil || len(txtRecords) == 0 {
			Cfg.Logger.Fatal("no available endpoint's data found (DNS TXT)", zap.Error(err))
		}
		aRecords, err := r.LookupHost(context.Background(), record.Target)
		if err != nil || len(aRecords) == 0 {
			Cfg.Logger.Fatal("no available endpoint's data found (DNS A)", zap.Error(err))
		}
		endpoint := Endpoint{
			Description: txtRecords[0],
			RemoteHost:  record.Target,
			RemoteIP:    aRecords[0],
			Latency:     probeEndpoint(record.Target),
		}
		Cfg.Endpoints[record.Target] = endpoint
	}
}

// GetFastestEndpoint returns fastest gRPC endpoint
func GetFastestEndpoint() Endpoint {
	getVPNEndpoints()

	var fastestEndpoint Endpoint
	var latency int64 = math.MaxInt64

	for _, e := range Cfg.Endpoints {
		if e.Latency < latency {
			fastestEndpoint = e
			latency = e.Latency
		}
		Cfg.Logger.Debug(
			"probing endpoint",
			zap.String("RemoteIP", e.RemoteIP),
			zap.String("RemoteHost", e.RemoteHost),
			zap.String("Description", e.Description),
			zap.Int64("Latency (ms)", e.Latency),
		)
	}

	return fastestEndpoint
}

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
