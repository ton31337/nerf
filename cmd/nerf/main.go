package main

import (
	"context"

	"github.com/getlantern/systray"
	"github.com/ton31337/nerf"
	"github.com/ton31337/nerf/cmd/nerf/icons"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

func main() {
	systray.Run(onReady, nil)
}

func onReady() {
	systray.SetIcon(icons.Disconnected)
	systray.SetTooltip("Hostinger Network")

	mConnect := systray.AddMenuItem("Connect", "Connect to Hostinger Network")
	mDisconnect := systray.AddMenuItem("Disconnect", "Disconnect from Hostinger Network")
	mQuitOrig := systray.AddMenuItem("Quit", "Quit")

	nerf.Cfg = nerf.NewConfig()

	logger, _ := zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:    "timestamp",
			EncodeTime: zapcore.ISO8601TimeEncoder,
			MessageKey: "message",
		},
	}.Build()

	nerf.Cfg.Logger = logger

	mDisconnect.Hide()

	go func(cfg *nerf.Config) {
		for {
			select {
			case <-mConnect.ClickedCh:
				systray.SetIcon(icons.Connecting)
				mConnect.SetTitle("Connecting")
				mConnect.Disable()
				connect()
				if cfg.Connected {
					mConnect.Hide()
					mDisconnect.Show()
					mDisconnect.SetTitle("Disconnect (" + cfg.CurrentEndpoint.RemoteIP + ")")
				}
			case <-mDisconnect.ClickedCh:
				disconnect()
				if !cfg.Connected {
					mConnect.SetTitle("Connect")
					mConnect.Show()
					mConnect.Enable()
					mDisconnect.Hide()
				}
			case <-mQuitOrig.ClickedCh:
				disconnect()
				systray.Quit()
				return
			}
		}
	}(&nerf.Cfg)
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
	response, err := client.Connect(context.Background(), request)
	if err != nil {
		nerf.Cfg.Logger.Fatal("can't connect to gRPC UNIX socket", zap.Error(err))
	}

	systray.SetIcon(icons.Connected)
	nerf.Cfg.Connected = true
	nerf.Cfg.CurrentEndpoint.RemoteIP = *response.RemoteIP
	nerf.Cfg.ClientIP = *response.ClientIP
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

	systray.SetIcon(icons.Disconnected)
	nerf.Cfg.Connected = false
}
