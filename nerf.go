package nerf

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

// Config structure
type Config struct {
	OAuth      *oauth2.Config
	Token      string
	ListenAddr string
	Teams      []string
	Login      string
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

// GetCertificates generates ca.crt, client.crt, client.key for Nebula
func (s *Server) GetCertificates(ctx context.Context, in *Request) (*Response, error) {
	fmt.Printf("Got request from %s\n", *in.Login)

	originToken := &TokenSource{
		AccessToken: *in.Token,
	}
	originOauthClient := oauth2.NewClient(oauth2.NoContext, originToken)
	originClient := github.NewClient(originOauthClient)
	originUser, _, _ := originClient.Users.Get(oauth2.NoContext, "")

	if originUser != nil {
		sudoToken := &TokenSource{
			AccessToken: os.Getenv("OAUTH_MASTER_TOKEN"),
		}
		sudoOauthClient := oauth2.NewClient(oauth2.NoContext, sudoToken)
		sudoClient := github.NewClient(sudoOauthClient)

		teams, _, _ := sudoClient.Teams.ListTeams(oauth2.NoContext, "hostinger", nil)

		if teams != nil {
			for _, team := range teams {
				users, _, _ := sudoClient.Teams.ListTeamMembers(oauth2.NoContext, *team.ID, nil)
				for _, user := range users {
					if *user.Login == *originUser.Login {
						Cfg.Teams = append(Cfg.Teams, *team.Name)
					}
				}
			}
		}
	}

	ca, crt, key := NebulaGenerateCertificate(Cfg.Teams, *originUser.Login)
	return &Response{Crt: &crt, Ca: &ca, Key: &key}, nil
}

// Cfg is a global configuration for Nerf internals
var Cfg Config

// NewConfig initializes NerfCfg
func NewConfig() Config {
	return Config{
		OAuth: &oauth2.Config{
			ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
			ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
			Scopes:       []string{"user:email"},
			Endpoint:     githuboauth.Endpoint,
		},
		ListenAddr: "127.0.0.1:1337",
	}
}
