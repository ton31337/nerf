package nerf

import (
	"context"
	math "math"
	"net"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
	grpc "google.golang.org/grpc"
)

// ClientCfg is a global configuration for Nerf client
var ClientCfg ClientConfig

// OauthClientID compile-time derived from -X github.com/ton31337/nerf.OauthClientID
var OauthClientID string

// OauthClientSecret compile-time derived from -X github.com/ton31337/nerf.OauthClientSecret
var OauthClientSecret string

// DNSAutoDiscoverZone compile-time derived from -Xgithub.com/ton31337/nerf.DNSAutoDiscoverZone
// E.g.: example.com which will be combined to _vpn._udp.example.com SRV query
var DNSAutoDiscoverZone string

// ClientConfig struct to store all the relevant data for a client
type ClientConfig struct {
	Logger           *zap.Logger
	OAuth            *oauth2.Config
	Token            string
	ListenAddr       string
	Login            string
	Endpoints        map[string]Endpoint
	SavedNameServers []string
}

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
		ClientCfg.Logger.Error("failed connecting to gRPC",
			zap.String("RemoteHost", remoteHost),
			zap.Error(err))
	}
	defer conn.Close()

	client := NewServerClient(conn)
	data := start.UnixNano()
	request := &PingRequest{Data: &data, Login: &ClientCfg.Login}
	response, err := client.Ping(ctx, request)
	if err != nil || *response.Data == 0 {
		return math.MaxInt64
	}

	return time.Since(start).Milliseconds()
}

// GetVPNEndpoints construct Endpoints map
func GetVPNEndpoints() {
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
		ClientCfg.Logger.Fatal("no available gRPC endpoints found (DNS SRV)", zap.Error(err))
	}

	for _, record := range srvRecords {
		txtRecords, err := r.LookupTXT(context.Background(), record.Target)
		if err != nil || len(txtRecords) == 0 {
			ClientCfg.Logger.Fatal("no available endpoint's data found (DNS TXT)", zap.Error(err))
		}
		aRecords, err := r.LookupHost(context.Background(), record.Target)
		if err != nil || len(aRecords) == 0 {
			ClientCfg.Logger.Fatal("no available endpoint's data found (DNS A)", zap.Error(err))
		}
		endpoint := Endpoint{
			Description: txtRecords[0],
			RemoteHost:  record.Target,
			RemoteIP:    aRecords[0],
			Latency:     probeEndpoint(record.Target),
		}
		ClientCfg.Endpoints[record.Target] = endpoint
	}
}

// GetFastestEndpoint returns fastest gRPC endpoint
func GetFastestEndpoint() Endpoint {
	GetVPNEndpoints()

	var fastestEndpoint Endpoint
	var latency int64 = math.MaxInt64

	for _, e := range ClientCfg.Endpoints {
		if e.Latency < latency {
			fastestEndpoint = e
		}
		ClientCfg.Logger.Debug(
			"probing endpoint",
			zap.String("RemoteIP", e.RemoteIP),
			zap.String("RemoteHost", e.RemoteHost),
			zap.String("Description", e.Description),
			zap.Int64("Latency (ms)", e.Latency),
		)
	}

	return fastestEndpoint
}

// NewClientConfig initializes ClientConfig
func NewClientConfig() ClientConfig {
	return ClientConfig{
		Logger: &zap.Logger{},
		OAuth: &oauth2.Config{
			ClientID:     OauthClientID,
			ClientSecret: OauthClientSecret,
			Scopes:       []string{"user:email"},
			Endpoint:     githuboauth.Endpoint,
		},
		Token:            "",
		ListenAddr:       "127.0.0.1:1337",
		Login:            "",
		Endpoints:        map[string]Endpoint{},
		SavedNameServers: []string{},
	}
}
