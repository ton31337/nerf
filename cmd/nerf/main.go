package main

import (
	"context"
	"os"

	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	"github.com/ton31337/nerf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

func main() {
	systray.Run(onReady, nil)
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("Hostinger Network")

	mConnect := systray.AddMenuItem("Connect", "Connect to Hostinger Network")
	mDisconnect := systray.AddMenuItem("Disconnect", "Disconnect from Hostinger Network")
	mQuitOrig := systray.AddMenuItem("Quit", "Quit")

	nerf.Cfg = nerf.NewConfig()

	logger, _ := zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
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

	go func() {
		for {
			select {
			case <-mConnect.ClickedCh:
				connect()
			case <-mDisconnect.ClickedCh:
				disconnect()
			case <-mQuitOrig.ClickedCh:
				disconnect()
				systray.Quit()
				return
			}
		}
	}()
}

func connect() {
	if nerf.Cfg.Connected {
		return
	}

	nerf.Auth()

	conn, err := grpc.Dial(
		"unix:/tmp/nerf.sock",
		grpc.WithInsecure())
	if err != nil {
		nerf.Cfg.Logger.Fatal("can't connect to gRPC UNIX socket", zap.Error(err))
	}

	defer conn.Close()

	client := nerf.NewApiClient(conn)

	request := &nerf.Request{Login: &nerf.Cfg.Login, Token: &nerf.Cfg.Token}
	_, err = client.Connect(context.Background(), request)
	if err != nil {
		nerf.Cfg.Logger.Fatal("can't connect to gRPC UNIX socket", zap.Error(err))
	}

	nerf.Cfg.Connected = true
}

func disconnect() {
	conn, err := grpc.Dial(
		"unix:/tmp/nerf.sock",
		grpc.WithInsecure())
	if err != nil {
		nerf.Cfg.Logger.Fatal("can't connect to gRPC UNIX socket", zap.Error(err))
	}

	defer conn.Close()

	client := nerf.NewApiClient(conn)

	request := &nerf.Notify{Login: &nerf.Cfg.Login}
	_, err = client.Disconnect(context.Background(), request)
	if err != nil {
		nerf.Cfg.Logger.Fatal("can't connect to gRPC UNIX socket", zap.Error(err))
	}

	nerf.Cfg.Connected = false
}
