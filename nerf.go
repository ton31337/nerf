package nerf

import (
	"context"
	"fmt"
	"log"
	math "math"
	"net"
	"time"

	"github.com/google/go-github/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
	grpc "google.golang.org/grpc"
)

// OauthMasterToken compile-time derived from -X github.com/ton31337/nerf.OauthMasterToken
var OauthMasterToken string

// Config struct to store all the relevant data for both client and server
type Config struct {
	Logger     *zap.Logger
	OAuth      *oauth2.Config
	Token      string
	ListenAddr string
	Login      string
	Nebula     *Nebula
	Endpoints  map[string]Endpoint
	Verbose    bool
}

// Server interface for Protobuf service
type Server struct {
}

// Endpoint struct to store all the relevant data about gRPC server,
// which generates and returns data for Nebula.
type Endpoint struct {
	Description string
	RemoteHost  string
	RemoteIP    string
	Latency     int64
}

// Ping get timestamp in milliseconds
func (s *Server) Ping(ctx context.Context, in *PingRequest) (*PingResponse, error) {
	if *in.Login == "" {
		return nil, fmt.Errorf("Failed gRPC ping request")
	}

	if Cfg.Verbose {
		Cfg.Logger.Info("Got ping request", zap.String("Login", *in.Login))
	}

	response := time.Now().Round(time.Millisecond).UnixNano() / 1e6
	return &PingResponse{Data: &response}, nil
}

// GetNebulaConfig generates config.yml for Nebula
func (s *Server) GetNebulaConfig(ctx context.Context, in *Request) (*Response, error) {
	var userTeams []string

	if *in.Login == "" {
		return nil, fmt.Errorf("Failed gRPC certificate request")
	}

	if Cfg.Verbose {
		Cfg.Logger.Info("Got certificate request", zap.String("Login", *in.Login))
	}

	originToken := &TokenSource{
		AccessToken: *in.Token,
	}
	originOauthClient := oauth2.NewClient(context.Background(), originToken)
	originClient := github.NewClient(originOauthClient)
	originUser, _, _ := originClient.Users.Get(context.Background(), "")

	if originUser != nil {
		Cfg.Login = *originUser.Login
		sudoToken := &TokenSource{
			AccessToken: OauthMasterToken,
		}
		sudoOauthClient := oauth2.NewClient(context.Background(), sudoToken)
		sudoClient := github.NewClient(sudoOauthClient)

		teams, _, _ := sudoClient.Teams.ListTeams(context.Background(), "hostinger", nil)

		for _, team := range teams {
			users, _, _ := sudoClient.Teams.ListTeamMembers(context.Background(), *team.ID, nil)
			for _, user := range users {
				if *user.Login == *originUser.Login {
					userTeams = append(userTeams, *team.Name)
				}
			}
		}
	}

	if len(userTeams) > 0 {
		if Cfg.Verbose {
			Cfg.Logger.Info("Teams found",
				zap.String("Login", *in.Login),
				zap.Strings("Teams", userTeams))
		}
	} else {
		if Cfg.Verbose {
			Cfg.Logger.Info("Teams not found", zap.String("Login", *in.Login))
		}
		return nil, fmt.Errorf("No teams founds")
	}

	config, err := NebulaGenerateConfig(userTeams)
	if err != nil {
		log.Fatalf("Failed creating configuration file for Nebula: %s\n", err)
	}

	return &Response{Config: &config}, nil
}

func probeEndpoint(remoteHost string) int64 {
	start := time.Now()
	conn, err := grpc.Dial(remoteHost+":9000", grpc.WithInsecure())
	if err != nil {
		Cfg.Logger.Error("Failed connecting to gRPC",
			zap.String("RemoteHost", remoteHost),
			zap.Error(err))
	}
	defer conn.Close()

	client := NewServerClient(conn)
	data := start.UnixNano()
	request := &PingRequest{Data: &data, Login: &Cfg.Login}
	response, err := client.Ping(context.Background(), request)
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
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, "udp", "1.1.1.1:53")
		},
	}

	_, srvRecords, err := r.LookupSRV(context.Background(), "vpn", "udp", "hostinger.io")
	if err != nil {
		log.Fatalf("Failed retrieving VPN endpoints: %s\n", err)
	}

	for _, record := range srvRecords {
		txtRecords, err := r.LookupTXT(context.Background(), record.Target)
		if err != nil || len(txtRecords) == 0 {
			log.Fatalf("Failed retrieving VPN endpoints (DNS TXT): %s\n", err)
		}
		aRecords, err := r.LookupHost(context.Background(), record.Target)
		if err != nil || len(aRecords) == 0 {
			log.Fatalf("Failed retrieving VPN endpoints (DNS A): %s\n", err)
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
	GetVPNEndpoints()

	var fastestEndpoint Endpoint
	var latency int64 = math.MaxInt64

	for _, e := range Cfg.Endpoints {
		if e.Latency < latency {
			fastestEndpoint = e
		}
	}

	return fastestEndpoint
}

// Cfg is a global configuration for Nerf internals
var Cfg Config

// OauthClientID compile-time derived from -X github.com/ton31337/nerf.OauthClientID
var OauthClientID string

// OauthClientSecret compile-time derived from -X github.com/ton31337/nerf.OauthClientSecret
var OauthClientSecret string

// NewConfig initializes NerfCfg
func NewConfig() Config {
	return Config{
		Logger:     &zap.Logger{},
		OAuth:      &oauth2.Config{ClientID: OauthClientID, ClientSecret: OauthClientSecret, Scopes: []string{"user:email"}, Endpoint: githuboauth.Endpoint},
		Token:      "",
		ListenAddr: "127.0.0.1:1337",
		Login:      "",
		Nebula: &Nebula{
			Subnet:      "172.17.0.0/12",
			Certificate: &Certificate{},
			LightHouse:  &LightHouse{},
		},
		Endpoints: map[string]Endpoint{},
		Verbose:   false,
	}
}
