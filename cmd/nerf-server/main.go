package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/ton31337/nerf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

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

	nerf.ServerCfg.Nebula.LightHouse.NebulaIP = lightHouseIPS[0]
	nerf.ServerCfg.Nebula.LightHouse.PublicIP = lightHouseIPS[1]

	nerf.ServerCfg.Logger.Debug("Nerf server started", zap.String("lightHouse", lightHouse))

	go func() {
		for range time.Tick(10 * time.Second) {
			if (time.Now().Unix() - nerf.ServerCfg.Teams.UpdatedAt) > int64(time.Hour.Seconds()) {
				nerf.ServerCfg.Teams.Mutex.Lock()
				nerf.ServerCfg.Logger.Debug(
					"begin-of-sync Github Teams with local cache")
				nerf.ServerCfg.Teams.Sync()
				nerf.ServerCfg.Logger.Debug(
					"end-of-sync Github Teams with local cache")
				nerf.ServerCfg.Teams.Mutex.Unlock()
			}
		}
	}()

	// Start gRPC server only when Teams are synced initially.
	for {
		if nerf.ServerCfg.Teams != nil && !nerf.ServerCfg.Teams.Mutex.Locked() {
			lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9000))
			if err != nil {
				nerf.ServerCfg.Logger.Fatal("failed to listen gRPC server", zap.Error(err))
			}

			grpcServer := grpc.NewServer()
			nerf.RegisterServerServer(grpcServer, &nerf.Server{})

			if err = grpcServer.Serve(lis); err != nil {
				nerf.ServerCfg.Logger.Fatal("can't serve gRPC", zap.Error(err))
			}

			break
		}
	}
}

func main() {
	lightHouse := flag.String("lighthouse", "", "Set the lighthouse. E.g.: <NebulaIP>:<PublicIP>")
	logLevel := flag.String(
		"log-level",
		"info",
		"Set the logging level - values are 'debug', 'info', 'warn', and 'error'",
	)
	printUsage := flag.Bool("help", false, "Print command line usage")

	flag.Parse()

	if *printUsage {
		flag.Usage()
		os.Exit(0)
	}

	nerf.ServerCfg = nerf.NewServerConfig()

	logger, _ := zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(nerf.StringToLogLevel(*logLevel)),
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:    "timestamp",
			EncodeTime: zapcore.ISO8601TimeEncoder,
			MessageKey: "message",
		},
	}.Build()

	nerf.ServerCfg.Logger = logger

	defer func() {
		_ = nerf.ServerCfg.Logger.Sync()
	}()

	startServer(*lightHouse)
}
