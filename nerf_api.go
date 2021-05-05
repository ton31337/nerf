package nerf

import (
	"context"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
	grpc "google.golang.org/grpc"

	"github.com/golang/protobuf/ptypes/empty"
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
)

// Cfg is a global configuration for Nerf client
var Cfg Config

// OauthClientID compile-time derived from -X github.com/ton31337/nerf.OauthClientID
var OauthClientID string

// OauthClientSecret compile-time derived from -X github.com/ton31337/nerf.OauthClientSecret
var OauthClientSecret string

// DNSAutoDiscoverZone compile-time derived from -Xgithub.com/ton31337/nerf.DNSAutoDiscoverZone
// E.g.: example.com which will be combined to _vpn._udp.example.com SRV query
var DNSAutoDiscoverZone string

// Config struct to store all the relevant data for a client
type Config struct {
	Logger           *zap.Logger
	OAuth            *oauth2.Config
	Token            string
	ListenAddr       string
	Login            string
	Endpoints        map[string]Endpoint
	CurrentEndpoint  *Endpoint
	SavedNameServers []string
	NebulaPid        *int
	Connected        bool
}

// Api interface for Protobuf service
type Api struct {
}

// StopApi handled for disconnect and quit. Or even nerf-api crash interruption.
func StopApi() {
	Cfg.Logger.Debug("disconnect", zap.String("Login", Cfg.Login))

	if Cfg.NebulaPid == nil {
		return
	}

	conn, err := grpc.Dial(Cfg.CurrentEndpoint.RemoteHost+":9000", grpc.WithInsecure())
	if err != nil {
		Cfg.Logger.Fatal(
			"can't connect to gRPC server",
			zap.Error(err),
		)
	}

	defer conn.Close()
	client := NewServerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = client.Disconnect(ctx, &Notify{Login: &Cfg.Login})
	if err != nil {
		Cfg.Logger.Error(
			"disconnect",
			zap.String("Login", Cfg.Login),
		)
	}
	if err = NebulaSetNameServers(Cfg.CurrentEndpoint, Cfg.SavedNameServers, false); err != nil {
		Cfg.Logger.Fatal("can't revert name servers", zap.Error(err))
	}

	if err = syscall.Kill(*Cfg.NebulaPid, syscall.SIGKILL); err != nil {
		Cfg.Logger.Fatal("can't stop Nebula", zap.Error(err))
	}

	Cfg.NebulaPid = nil
	Cfg.CurrentEndpoint = &Endpoint{}
}

func startApi() {
	e := GetFastestEndpoint()
	Cfg.CurrentEndpoint = &e

	if Cfg.NebulaPid != nil {
		Cfg.Logger.Fatal("Nebula instance already running")
	}

	if Cfg.CurrentEndpoint.RemoteHost == "" {
		Cfg.Logger.Fatal("no available gRPC endpoints found")
	}

	if err := NebulaAddLightHouseStaticRoute(Cfg.CurrentEndpoint); err != nil {
		Cfg.Logger.Fatal(
			"can't create route",
			zap.String("destination", e.RemoteIP),
			zap.Error(err),
		)
	}

	Cfg.Logger.Debug("authorized", zap.String("login", Cfg.Login))

	conn, err := grpc.Dial(Cfg.CurrentEndpoint.RemoteHost+":9000", grpc.WithInsecure())
	if err != nil {
		Cfg.Logger.Fatal(
			"can't connect to gRPC server",
			zap.Error(err),
			zap.String("remoteHost", Cfg.CurrentEndpoint.RemoteHost),
			zap.String("description", Cfg.CurrentEndpoint.Description),
		)
	}

	defer conn.Close()

	client := NewServerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	request := &Request{Token: &Cfg.Token, Login: &Cfg.Login}
	response, err := client.Connect(ctx, request)
	if err != nil {
		Cfg.Logger.Fatal(
			"can't connect to gRPC server",
			zap.Error(err),
			zap.String("remoteHost", Cfg.CurrentEndpoint.RemoteHost),
			zap.String("description", Cfg.CurrentEndpoint.Description),
		)
	}

	Cfg.Logger.Debug("connected to LightHouse",
		zap.String("ClientIP", *response.ClientIP),
		zap.String("LightHouseIP", *response.LightHouseIP),
		zap.Strings("Teams", response.Teams))

	out, err := os.Create(path.Join(NebulaDir(), "config.yml"))
	if err != nil {
		Cfg.Logger.Fatal("can't create Nebula config", zap.Error(err))
	}

	if _, err := out.WriteString(*response.Config); err != nil {
		Cfg.Logger.Fatal("can't write Nebula config", zap.Error(err))
	}
	defer out.Close()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	if err := NebulaSetNameServers(&e, []string{*response.LightHouseIP}, true); err != nil {
		Cfg.Logger.Fatal("can't set custom DNS servers", zap.Error(err))
	}

	pid, err := NebulaStart()
	if err != nil {
		Cfg.Logger.Fatal("can't start Nebula client", zap.Error(err))
	}

	Cfg.NebulaPid = &pid

	<-done

	StopApi()
}

// Connect used to notify API about initiated connect
func (s *Api) Connect(ctx context.Context, in *Request) (*google_protobuf.Empty, error) {
	var err error

	Cfg.Login = *in.Login
	Cfg.Token = *in.Token

	go startApi()

	return &empty.Empty{}, err
}

// Disconnect used to notify API about initiated disconnect
func (s *Api) Disconnect(ctx context.Context, in *Notify) (*google_protobuf.Empty, error) {
	var err error

	Cfg.Login = *in.Login

	go StopApi()

	return &empty.Empty{}, err
}

// NewConfig initializes Config
func NewConfig() Config {
	return Config{
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
		CurrentEndpoint:  &Endpoint{},
		SavedNameServers: []string{},
		NebulaPid:        nil,
		Connected:        false,
	}
}
