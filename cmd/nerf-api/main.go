package main

import (
	"flag"
	"net"
	"os"
	"os/signal"

	"github.com/ton31337/nerf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

const UnixSockAddr = "/tmp/nerf.sock"

func main() {
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

	nerf.Cfg = nerf.NewConfig()

	logger, _ := zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(nerf.StringToLogLevel(*logLevel)),
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",
		},
	}.Build()

	nerf.Cfg.Logger = logger

	defer func() {
		_ = nerf.Cfg.Logger.Sync()
	}()

	// Check if nerf-api is already running or not. Fail to start if exists.
	listen_check, err := net.Listen("unix", UnixSockAddr)
	if err != nil {
		nerf.Cfg.Logger.Fatal("nerf-api instance is already running", zap.Error(err))
	}
	defer listen_check.Close()

	if err := os.RemoveAll(UnixSockAddr); err != nil {
		nerf.Cfg.Logger.Fatal("can't remove UNIX socket", zap.Error(err))
	}

	lis, err := net.Listen("unix", UnixSockAddr)
	if err != nil {
		nerf.Cfg.Logger.Fatal("can't listen UNIX socket", zap.Error(err))
	}

	if os.Chmod(UnixSockAddr, 0777) != nil {
		nerf.Cfg.Logger.Fatal("can't set write permissions for UNIX socket", zap.Error(err))
	}

	defer lis.Close()

	grpcServer := grpc.NewServer()
	nerf.RegisterApiServer(grpcServer, &nerf.Api{})

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			nerf.Cfg.Logger.Fatal("can't serve gRPC", zap.Error(err))
		}
	}()

	<-done

	grpcServer.Stop()
	nerf.StopApi()
}
