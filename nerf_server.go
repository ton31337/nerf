package nerf

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/go-github/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// ServerCfg is a global configuration for Nerf server
var ServerCfg ServerConfig

// OauthMasterToken compile-time derived from -X github.com/ton31337/nerf.OauthMasterToken
var OauthMasterToken string

// OauthOrganization compile-time derived from -X github.com/ton31337/nerf.OauthOrganization
// E.g.: example which will be used to retrieve teams by username from GitHub in this org.
var OauthOrganization string

// ServerConfig struct to store all the relevant data for a server
type ServerConfig struct {
	Logger    *zap.Logger
	Login     string
	Nebula    *Nebula
	Teams     *Teams
	GaidysUrl string
}

// Server interface for Protobuf service
type Server struct {
}

// Teams struct to store all the relevant data about Github Teams.
type Teams struct {
	Mutex     *NerfMutex
	Members   map[string][]string
	UpdatedAt int64
}

// Ping get timestamp in milliseconds
func (s *Server) Ping(ctx context.Context, in *PingRequest) (*PingResponse, error) {
	if in.Login == "" {
		return nil, fmt.Errorf("failed gRPC ping request")
	}

	ServerCfg.Logger.Debug("got ping request", zap.String("Login", in.Login))

	response := time.Now().Round(time.Millisecond).UnixNano() / 1e6
	return &PingResponse{Data: response}, nil
}

func (t *Teams) User(login string) []string {
	var teams []string

	for team, users := range ServerCfg.Teams.Members {
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
func (t *Teams) Sync() {
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
			ServerCfg.Teams.Members[*team.Name] = make([]string, 0)
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
					ServerCfg.Teams.Members[*team.Name] = append(
						ServerCfg.Teams.Members[*team.Name],
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
	ServerCfg.Teams.UpdatedAt = time.Now().Unix()
}

// Disconnect - notify the server about disconnection
func (s *Server) Disconnect(ctx context.Context, in *Notify) (*empty.Empty, error) {
	var err error

	if in.Login == "" {
		err = fmt.Errorf("failed gRPC disconnect request")
	}

	ServerCfg.Logger.Debug("disconnect", zap.String("Login", in.Login))

	return &empty.Empty{}, err
}

// Connect - connects to the server which generates config.yml for Nebula
func (s *Server) Connect(ctx context.Context, in *Request) (*Response, error) {
	if in.Login == "" {
		return nil, fmt.Errorf("failed gRPC certificate request")
	}

	ServerCfg.Logger.Debug("connect", zap.String("Login", in.Login))

	token := &TokenSource{
		AccessToken: in.Token,
	}
	oclient := oauth2.NewClient(context.Background(), token)
	client := github.NewClient(oclient)
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		return nil, fmt.Errorf("failed validate login %s(%s): %s", user, in.Login, err)
	}

	userTeams := ServerCfg.Teams.User(*user.Login)
	if len(userTeams) == 0 {
		ServerCfg.Logger.Debug("teams not found", zap.String("Login", *user.Login))
		return nil, fmt.Errorf("no teams founds")
	}

	ServerCfg.Login = *user.Login
	clientIP, err := NebulaClientIP()
	if err != nil {
		ServerCfg.Logger.Debug("IP address not found in IPAM", zap.String("Login", *user.Login))
		return nil, fmt.Errorf("no IP address")
	}

	config, err := NebulaGenerateConfig(userTeams)
	if err != nil {
		ServerCfg.Logger.Error(
			"can't generate config for Nebula",
			zap.String("Login", *user.Login),
			zap.Strings("Teams", userTeams),
			zap.Error(err),
		)
	}

	ServerCfg.Logger.Debug("teams found",
		zap.String("Login", *user.Login),
		zap.String("ClientIP", clientIP),
		zap.Strings("Teams", userTeams))

	return &Response{
		Config:       config,
		ClientIP:     clientIP,
		LightHouseIP: ServerCfg.Nebula.LightHouse.NebulaIP,
		Teams:        userTeams,
	}, nil
}

func NewServerConfig() ServerConfig {
	return ServerConfig{
		Logger: &zap.Logger{},
		Login:  "",
		Nebula: &Nebula{
			Certificate: &Certificate{},
			LightHouse:  &LightHouse{},
		},
		Teams: &Teams{
			Members:   make(map[string][]string),
			UpdatedAt: time.Now().Unix() - 24*3600,
			Mutex:     &NerfMutex{InUse: true},
		},
		GaidysUrl: os.Getenv("GAIDYS_URL"),
	}
}
