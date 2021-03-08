package nerf

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

// OauthMasterToken compile-time derived from -X github.com/ton31337/nerf.OauthMasterToken
var OauthMasterToken string

// Config struct to store all the relevant data for both client and server
type Config struct {
	OAuth      *oauth2.Config
	Token      string
	ListenAddr string
	Teams      []string
	Login      string
	Nebula     *Nebula
}

// Server interface for Protobuf service
type Server struct {
}

// TokenSource defines Access Token for Github
type TokenSource struct {
	AccessToken string
}

// Token initializes Access Token for Github
func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

// GetNebulaConfig generates config.yml for Nebula
func (s *Server) GetNebulaConfig(ctx context.Context, in *Request) (*Response, error) {
	if *in.Login == "" {
		return nil, fmt.Errorf("Failed gRPC request")
	}

	fmt.Printf("Got certificate request from: %s\n", *in.Login)

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
					Cfg.Teams = append(Cfg.Teams, *team.Name)
				}
			}
		}
	}

	out, err := os.Create(path.Join(NebulaDir(), "config.yml"))
	if err != nil {
		return nil, err
	}
	defer out.Close()

	config, err := NebulaGenerateConfig()
	if err != nil {
		log.Fatalf("Failed creating configuration file for Nebula: %s", err)
	}

	return &Response{Config: &config}, nil
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
		OAuth:      &oauth2.Config{ClientID: OauthClientID, ClientSecret: OauthClientSecret, Scopes: []string{"user:email"}, Endpoint: githuboauth.Endpoint},
		Token:      "",
		ListenAddr: "127.0.0.1:1337",
		Teams:      []string{},
		Login:      "",
		Nebula: &Nebula{
			Subnet:      "172.17.0.0/12",
			Certificate: &Certificate{},
			LightHouse:  &LightHouse{},
		},
	}
}
