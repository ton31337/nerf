package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path"
	"strings"
	"time"

	"github.com/ton31337/nerf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

func startClient(redirect bool) {
	err := nerf.NebulaDownload()
	if err != nil {
		if _, err := os.Stat(nerf.NebulaExecutable()); err != nil {
			nerf.ClientCfg.Logger.Fatal("can't install Nebula:", zap.Error(err))
		}
	}

	// Before probing all gRPC endpoints we MUST be authenticated.
	// Otherwise, we can't continue and nothing happens.
	nerf.Auth()
	e := nerf.GetFastestEndpoint()
	if e.RemoteHost == "" {
		nerf.ClientCfg.Logger.Fatal("no available gRPC endpoints found")
	}

	nerf.ClientCfg.Logger.Debug("found fastest endpoint",
		zap.String("RemoteIP", e.RemoteIP),
		zap.String("RemoteHost", e.RemoteHost),
		zap.String("Description", e.Description))

	if err := nerf.NebulaAddLightHouseStaticRoute(&e); err != nil {
		nerf.ClientCfg.Logger.Fatal(
			"can't create route",
			zap.String("destination", e.RemoteIP),
			zap.Error(err),
		)
	}

	nerf.ClientCfg.Logger.Debug("authorized", zap.String("login", nerf.ClientCfg.Login))

	conn, err := grpc.Dial(e.RemoteHost+":9000", grpc.WithInsecure())
	if err != nil {
		nerf.ClientCfg.Logger.Fatal(
			"can't connect to gRPC server",
			zap.Error(err),
			zap.String("remoteHost", e.RemoteHost),
			zap.String("description", e.Description),
		)
	}

	defer conn.Close()

	client := nerf.NewServerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	request := &nerf.Request{Token: &nerf.ClientCfg.Token, Login: &nerf.ClientCfg.Login}
	response, err := client.Connect(ctx, request)
	if err != nil {
		nerf.ClientCfg.Logger.Fatal(
			"can't connect to gRPC server",
			zap.Error(err),
			zap.String("remoteHost", e.RemoteHost),
			zap.String("description", e.Description),
		)
	}

	nerf.ClientCfg.Logger.Debug("connected to LightHouse",
		zap.String("ClientIP", *response.ClientIP),
		zap.String("LightHouseIP", *response.LightHouseIP),
		zap.Strings("Teams", response.Teams))

	config := *response.Config

	if !redirect {
		configLines := strings.Split(config, "\n")
		var filteredConfigLines []string
		skipNext := false

		for _, line := range configLines {

			if skipNext {
				skipNext = false
				continue
			}

			if strings.Contains(line, "0.0.0.0/1") {
				skipNext = true
				continue
			}

			if strings.Contains(line, "128.0.0.0/1") {
				skipNext = true
				continue
			}
			filteredConfigLines = append(filteredConfigLines, line)
		}
		config = strings.Join(filteredConfigLines, "\n")
	}

	out, err := os.Create(path.Join(nerf.NebulaDir(), "config.yml"))
	if err != nil {
		nerf.ClientCfg.Logger.Fatal("can't create Nebula config", zap.Error(err))
	}

	if _, err := out.WriteString(config); err != nil {
		nerf.ClientCfg.Logger.Fatal("can't write Nebula config", zap.Error(err))
	}
	defer out.Close()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	if err := nerf.NebulaSetNameServers(&e, []string{*response.LightHouseIP}, true); err != nil {
		nerf.ClientCfg.Logger.Fatal("can't set custom DNS servers", zap.Error(err))
	}

	if err := nerf.NebulaStart(); err != nil {
		nerf.ClientCfg.Logger.Fatal("can't start Nebula client", zap.Error(err))
	}

	<-done

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	notify, err := client.Disconnect(ctx, &nerf.Notify{Login: &nerf.ClientCfg.Login})
	if err != nil {
		nerf.ClientCfg.Logger.Error(
			"disconnect",
			zap.String("Login", nerf.ClientCfg.Login),
			zap.String("Response", notify.String()),
		)
	}
	if err = nerf.NebulaSetNameServers(&e, nerf.ClientCfg.SavedNameServers, false); err != nil {
		nerf.ClientCfg.Logger.Fatal("can't revert name servers", zap.Error(err))
	}
}

func main() {
	logLevel := flag.String(
		"log-level",
		"info",
		"Set the logging level - values are 'debug', 'info', 'warn', and 'error'",
	)
	redirectAll := flag.Bool("redirect-all", true, "Redirect all traffic through Nebula")
	printUsage := flag.Bool("help", false, "Print command line usage")

	flag.Parse()

	if *printUsage {
		flag.Usage()
		os.Exit(0)
	}

	nerf.ClientCfg = nerf.NewClientConfig()

	logger, _ := zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(nerf.StringToLogLevel(*logLevel)),
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",
		},
	}.Build()

	nerf.ClientCfg.Logger = logger

	defer func() {
		_ = nerf.ClientCfg.Logger.Sync()
	}()

	startClient(*redirectAll)
}
