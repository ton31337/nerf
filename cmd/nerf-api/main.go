package main

import (
	"flag"
	"log"
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

	if err := os.RemoveAll(UnixSockAddr); err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("unix", UnixSockAddr)
	if err != nil {
		nerf.Cfg.Logger.Fatal("can't listen UNIX socket", zap.Error(err))
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