package main

import (
	"context"
	"fmt"
	"time"

	"github.com/getlantern/systray"
	"github.com/ton31337/nerf"
	"github.com/ton31337/nerf/cmd/nerf/icons"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

var mStatus, mRemoteIP, mConnect, mDisconnect, mQuitOrig *systray.MenuItem
var connectionTime time.Time
var connectionTicker *time.Ticker

func main() {
	systray.Run(onReady, nil)
}

func onReady() {
	systray.SetIcon(icons.Disconnected)
	systray.SetTooltip("Hostinger Network")

	mStatus = systray.AddMenuItem("Status: Not connected", "Connection status")
	mStatus.Disable()
	mRemoteIP = systray.AddMenuItem("Remote IP:", "Remote IP address")
	mRemoteIP.Disable()
	mRemoteIP.Hide()
	systray.AddSeparator()
	mConnect = systray.AddMenuItem("Connect", "Connect to Hostinger Network")
	mDisconnect = systray.AddMenuItem("Disconnect", "Disconnect from Hostinger Network")
	mQuitOrig = systray.AddMenuItem("Quit", "Quit")

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
				guiConnecting()
				connect()
				if cfg.Connected {
					guiConnected()
				} else {
					guiDisconnected()
				}
			case <-mDisconnect.ClickedCh:
				disconnect()
				if !cfg.Connected {
					guiDisconnected()
				}
			case <-mQuitOrig.ClickedCh:
				disconnect()
				systray.Quit()
				return
			}
		}
	}(&nerf.Cfg)
}

func guiConnected() {
	systray.SetIcon(icons.Connected)
	mStatus.SetTitle("Status: Connected")
	mRemoteIP.SetTitle("Remote IP: " + nerf.Cfg.CurrentEndpoint.RemoteIP)
	mDisconnect.SetTitle("Disconnect")
	mRemoteIP.Show()
	mDisconnect.Show()
	connectionTicker = time.NewTicker(1 * time.Second)
	go func() {
		for {
			<-connectionTicker.C
			if !nerf.Cfg.Connected {
				connectionTicker.Stop()
				return
			}
			connectionDuration := int(time.Since(connectionTime).Seconds())
			mStatus.SetTitle("Status: Connected (" + fmt.Sprintf("%02d:%02d:%02d", connectionDuration/3600, (connectionDuration%3600)/60, connectionDuration%60) + ")")
		}
	}()
}

func guiDisconnected() {
	systray.SetIcon(icons.Disconnected)
	mStatus.SetTitle("Status: Not connected")
	mRemoteIP.Hide()
	mConnect.Show()
	mConnect.Enable()
	mDisconnect.Hide()
}

func guiConnecting() {
	systray.SetIcon(icons.Connecting)
	mStatus.SetTitle("Status: Connecting")
	mConnect.Hide()
}

func connect() {
	if nerf.Cfg.Connected {
		return
	}

	nerf.Auth()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx,
		"unix:/tmp/nerf.sock",
		grpc.WithInsecure())
	if err != nil {
		nerf.Cfg.Logger.Error("can't connect to gRPC UNIX socket", zap.Error(err))
		return
	}

	defer conn.Close()

	client := nerf.NewApiClient(conn)

	request := &nerf.Request{Login: &nerf.Cfg.Login, Token: &nerf.Cfg.Token}
	response, err := client.Connect(context.Background(), request)
	if err != nil {
		nerf.Cfg.Logger.Error("can't connect to gRPC UNIX socket", zap.Error(err))
		return
	}

	connectionTime = time.Now()
	nerf.Cfg.Connected = true
	nerf.Cfg.CurrentEndpoint.RemoteIP = *response.RemoteIP
	nerf.Cfg.ClientIP = *response.ClientIP
	guiConnected()
}

func disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx,
		"unix:/tmp/nerf.sock",
		grpc.WithInsecure())
	if err != nil {
		nerf.Cfg.Logger.Error("can't connect to gRPC UNIX socket", zap.Error(err))
		return
	}

	defer conn.Close()

	client := nerf.NewApiClient(conn)

	request := &nerf.Notify{Login: &nerf.Cfg.Login}
	_, err = client.Disconnect(context.Background(), request)
	if err != nil {
		nerf.Cfg.Logger.Error("can't connect to gRPC UNIX socket", zap.Error(err))
		return
	}

	guiDisconnected()
	nerf.Cfg.Connected = false
}
