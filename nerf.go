package nerf

import (
	"os"

	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

// Config structure
type Config struct {
	OAuth        *oauth2.Config
	Organization string
	Email        string
	Teams        []string
	ListenAddr   string
}

// Cfg is a global configuration for Nerf internals
var Cfg Config

// NewNerfConfig initializes NerfCfg
func NewNerfConfig() Config {
	return Config{
		OAuth: &oauth2.Config{
			ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
			ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
			Scopes:       []string{"user:email", "repo"},
			Endpoint:     githuboauth.Endpoint,
		},
		Organization: "Hostinger International",
		ListenAddr:   "127.0.0.1:1337",
	}
}
