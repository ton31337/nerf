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

// OauthOrganization compile-time derived from -X github.com/ton31337/nerf.OauthOrganization
// E.g.: example which will be used to retrieve teams by username from GitHub in this org.
var OauthOrganization string

// DNSAutoDiscoverZone compile-time derived from -Xgithub.com/ton31337/nerf.DNSAutoDiscoverZone
// E.g.: example.com which will be combined to _vpn._udp.example.com SRV query
var DNSAutoDiscoverZone string

// Config struct to store all the relevant data for both client and server
type Config struct {
	Logger     *zap.Logger
	OAuth      *oauth2.Config
	Token      string
	ListenAddr string
	Login      string
	Nebula     *Nebula
	Teams      *Teams
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

// Teams struct to store all the relevant data about Github Teams.
type Teams struct {
	Members   map[string][]string
	UpdatedAt int64
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

func teamsByUser(login string) []string {
	var teams []string

	for team, users := range Cfg.Teams.Members {
		for _, user := range users {
			if login == user {
				teams = append(teams, team)
			}
		}
	}

	return teams
}

// SyncTeams sync Github Teams with local cache
// Scheduled every 10 seconds and updated every hour.
func SyncTeams() {
	token := &TokenSource{
		AccessToken: OauthMasterToken,
	}
	oclient := oauth2.NewClient(context.Background(), token)
	client := github.NewClient(oclient)

	teamOptions := github.ListOptions{PerPage: 500}

	for {
		teams, respTeams, _ := client.Teams.ListTeams(
			context.Background(),
			OauthOrganization,
			&teamOptions,
		)
		for _, team := range teams {
			Cfg.Teams.Members[*team.Name] = make([]string, 0)
			usersOptions := &github.TeamListTeamMembersOptions{
				ListOptions: github.ListOptions{PerPage: 500},
			}
			for {
				users, respUsers, _ := client.Teams.ListTeamMembers(
					context.Background(),
					*team.ID,
					usersOptions,
				)
				for _, user := range users {
					Cfg.Teams.Members[*team.Name] = append(
						Cfg.Teams.Members[*team.Name],
						*user.Login,
					)
				}
				if respUsers.NextPage == 0 {
					break
				}
				usersOptions.ListOptions.Page = respUsers.NextPage
			}
		}
		if respTeams.NextPage == 0 {
			break
		}
		teamOptions.Page = respTeams.NextPage
	}
	Cfg.Teams.UpdatedAt = time.Now().Unix()
}

// GetNebulaConfig generates config.yml for Nebula
func (s *Server) GetNebulaConfig(ctx context.Context, in *Request) (*Response, error) {
	if *in.Login == "" {
		return nil, fmt.Errorf("Failed gRPC certificate request")
	}

	if Cfg.Verbose {
		Cfg.Logger.Info("Got certificate request", zap.String("Login", *in.Login))
	}

	token := &TokenSource{
		AccessToken: *in.Token,
	}
	oclient := oauth2.NewClient(context.Background(), token)
	client := github.NewClient(oclient)
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		return nil, fmt.Errorf("Failed validate login %s(%s): %s\n", user, *in.Login, err)
	}

	userTeams := teamsByUser(*user.Login)
	if len(userTeams) == 0 {
		if Cfg.Verbose {
			Cfg.Logger.Info("Teams not found", zap.String("Login", *user.Login))
			return nil, fmt.Errorf("No teams founds")
		}
	}

	Cfg.Login = *user.Login
	clientIP := NebulaClientIP()

	config, err := NebulaGenerateConfig(userTeams)
	if err != nil {
		Cfg.Logger.Error(
			"Can't generate config for Nebula",
			zap.String("Login", *user.Login),
			zap.Strings("Teams", userTeams),
		)
	}

	if Cfg.Verbose {
		Cfg.Logger.Info("Teams found",
			zap.String("Login", *user.Login),
			zap.String("ClientIP", clientIP),
			zap.Strings("Teams", userTeams))
	}

	return &Response{Config: &config, ClientIP: &clientIP, Teams: userTeams}, nil
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

	_, srvRecords, err := r.LookupSRV(context.Background(), "vpn", "udp", DNSAutoDiscoverZone)
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
		if Cfg.Verbose {
			Cfg.Logger.Info(
				"Probing endpoint",
				zap.String("RemoteIP", e.RemoteIP),
				zap.String("RemoteHost", e.RemoteHost),
				zap.String("Description", e.Description),
				zap.Int64("Latency (ms)", e.Latency),
			)
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
		Logger: &zap.Logger{},
		OAuth: &oauth2.Config{
			ClientID:     OauthClientID,
			ClientSecret: OauthClientSecret,
			Scopes:       []string{"user:email"},
			Endpoint:     githuboauth.Endpoint,
		},
		Token:      "",
		ListenAddr: "127.0.0.1:1337",
		Login:      "",
		Nebula: &Nebula{
			Subnet:      "172.17.0.0/12",
			Certificate: &Certificate{},
			LightHouse:  &LightHouse{},
		},
		Endpoints: map[string]Endpoint{},
		Teams: &Teams{
			Members:   make(map[string][]string),
			UpdatedAt: time.Now().Unix() - 24*3600,
		},
		Verbose: false,
	}
}
