package main

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/ton31337/nerf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

const UnixSockAddr = "/tmp/nerf.sock"

func main() {
	var d net.Dialer

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
			TimeKey:    "timestamp",
			EncodeTime: zapcore.ISO8601TimeEncoder,
			MessageKey: "message",
		},
	}.Build()

	nerf.Cfg.Logger = logger

	err := nerf.NebulaDownload()
	if err != nil {
		if _, err := os.Stat(nerf.NebulaExecutable()); err != nil {
			nerf.Cfg.Logger.Fatal("can't install Nebula", zap.Error(err))
		}
	}

	defer func() {
		_ = nerf.Cfg.Logger.Sync()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Check if nerf-api is already running or not.
	// If the socket is orphaned, check against it.
	// If an error is returned, delete it.
	_, err = d.DialContext(ctx, "unix", UnixSockAddr)
	if err != nil {
		if err := os.RemoveAll(UnixSockAddr); err != nil {
			nerf.Cfg.Logger.Fatal("can't remove UNIX socket", zap.Error(err))
		}
	}

	lis, err := net.Listen("unix", UnixSockAddr)
	if err != nil {
		nerf.Cfg.Logger.Fatal("nerf-api instance is already running", zap.Error(err))
	}
	defer lis.Close()

	if os.Chmod(UnixSockAddr, 0777) != nil {
		nerf.Cfg.Logger.Fatal("can't set write permissions for UNIX socket", zap.Error(err))
	}

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
