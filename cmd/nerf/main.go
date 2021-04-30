package main

import (
	"context"
	"flag"
	"fmt"
	"net"
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

func stringToLogLevel(level string) zapcore.Level {
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

func startServer(lightHouse string) {
	if lightHouse == "" {
		fmt.Println("-lighthouse flag must be set")
		flag.Usage()
		os.Exit(1)
	}

	lightHouseIPS := strings.Split(lightHouse, ":")
	if len(lightHouseIPS) < 2 {
		fmt.Println("The format for lighthouse must be <NebulaIP>:<PublicIP>")
		flag.Usage()
		os.Exit(1)
	}

	if err := net.ParseIP(lightHouseIPS[0]); err == nil {
		fmt.Println("NebulaIP address is not IPv4")
		flag.Usage()
		os.Exit(1)
	}

	if err := net.ParseIP(lightHouseIPS[1]); err == nil {
		fmt.Println("PublicIP address is not IPv4")
		flag.Usage()
		os.Exit(1)
	}

	nerf.Cfg.Nebula.LightHouse.NebulaIP = lightHouseIPS[0]
	nerf.Cfg.Nebula.LightHouse.PublicIP = lightHouseIPS[1]

	nerf.Cfg.Logger.Debug("Nerf server started", zap.String("lightHouse", lightHouse))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9000))
	if err != nil {
		nerf.Cfg.Logger.Fatal("failed to listen gRPC server", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	nerf.RegisterServerServer(grpcServer, &nerf.Server{})

	go func() {
		for range time.Tick(10 * time.Second) {
			if (time.Now().Unix() - nerf.Cfg.Teams.UpdatedAt) > 3600 {
				nerf.Cfg.Logger.Debug(
					"begin-of-sync Github Teams with local cache")
				nerf.SyncTeams()
				nerf.Cfg.Logger.Debug(
					"end-of-sync Github Teams with local cache")
			}
		}
	}()

	if err := grpcServer.Serve(lis); err != nil {
		nerf.Cfg.Logger.Fatal("can't serve gRPC", zap.Error(err))
	}
}

func startClient(redirect bool) {
	err := nerf.NebulaDownload()
	if err != nil {
		if _, err := os.Stat(nerf.NebulaExecutable()); err != nil {
			nerf.Cfg.Logger.Fatal("can't install Nebula:", zap.Error(err))
		}
	}

	// Before probing all gRPC endpoints we MUST be authenticated.
	// Otherwise, we can't continue and nothing happens.
	nerf.Auth()
	e := nerf.GetFastestEndpoint()
	if e.RemoteHost == "" {
		nerf.Cfg.Logger.Fatal("no available gRPC endpoints found")
	}

	nerf.Cfg.Logger.Debug("found fastest endpoint",
		zap.String("RemoteIP", e.RemoteIP),
		zap.String("RemoteHost", e.RemoteHost),
		zap.String("Description", e.Description))

	if err := nerf.NebulaAddLightHouseStaticRoute(&e); err != nil {
		nerf.Cfg.Logger.Fatal("can't create route", zap.String("destination", e.RemoteIP), zap.Error(err))
	}

	nerf.Cfg.Logger.Debug("authorized", zap.String("login", nerf.Cfg.Login))

	conn, err := grpc.Dial(e.RemoteHost+":9000", grpc.WithInsecure())
	if err != nil {
		nerf.Cfg.Logger.Fatal("can't connect to gRPC server", zap.Error(err), zap.String("remoteHost", e.RemoteHost), zap.String("description", e.Description))
	}

	defer conn.Close()

	client := nerf.NewServerClient(conn)
	request := &nerf.Request{Token: &nerf.Cfg.Token, Login: &nerf.Cfg.Login}
	response, err := client.Connect(context.Background(), request)
	if err != nil {
		nerf.Cfg.Logger.Fatal("can't connect to gRPC server", zap.Error(err), zap.String("remoteHost", e.RemoteHost), zap.String("description", e.Description))
	}

	nerf.Cfg.Logger.Debug("connected to LightHouse",
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
		nerf.Cfg.Logger.Fatal("can't create Nebula config", zap.Error(err))
	}

	if _, err := out.WriteString(config); err != nil {
		nerf.Cfg.Logger.Fatal("can't write Nebula config", zap.Error(err))
	}
	defer out.Close()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	if err := nerf.NebulaSetNameServers(&e, *response.LightHouseIP); err != nil {
		nerf.Cfg.Logger.Fatal("can't set custom DNS servers", zap.Error(err))
	}

	if err := nerf.NebulaStart(); err != nil {
		nerf.Cfg.Logger.Fatal("can't start Nebula client", zap.Error(err))
	}

	<-done
	notify, err := client.Disconnect(context.Background(), &nerf.Notify{Login: &nerf.Cfg.Login})
	if err != nil {
		nerf.Cfg.Logger.Error("Disconnect", zap.String("Login", nerf.Cfg.Login), zap.String("Response", notify.String()))
	}
}

func main() {
	server := flag.Bool("server", false, "Start gRPC server to generate config for Nebula")
	lightHouse := flag.String("lighthouse", "", "Set the lighthouse. E.g.: <NebulaIP>:<PublicIP>")
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

	nerf.Cfg = nerf.NewConfig()

	logger, _ := zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(stringToLogLevel(*logLevel)),
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",
		},
	}.Build()

	nerf.Cfg.Logger = logger

	defer func() {
		if err := nerf.Cfg.Logger.Sync(); err != nil {
			panic("failed sync Logger")
		}
	}()

	if *server {
		startServer(*lightHouse)
	} else {
		startClient(*redirectAll)
	}
}
