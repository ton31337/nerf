package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/getlantern/systray"
	"github.com/ton31337/nerf"
	"google.golang.org/grpc"
)

func main() {
	err := nerf.NebulaDownload()
	if err != nil {
		if _, err := os.Stat(nerf.NebulaExecutable()); err != nil {
			log.Fatalf("Failed installing Nebula: %s\n", err)
		}
	}

	nerf.Cfg = nerf.NewConfig()
	systray.Run(onReady, onExit)
}

func getIcon(s string) []byte {
	b, err := ioutil.ReadFile(s)
	if err != nil {
		fmt.Print(err)
	}
	return b
}

func onReady() {
	systray.SetIcon(getIcon("favicon.ico"))
	systray.SetTitle(" Hostinger VPN")
	connectLT := systray.AddMenuItem("Lithuania", "Lithuania")
	systray.AddSeparator()
	quit := systray.AddMenuItem("Quit", "Quits this app")

	go func() {
		for {
			select {
			case <-connectLT.ClickedCh:
				nerf.Auth()

				conn, err := grpc.Dial(":9000", grpc.WithInsecure())
				if err != nil {
					log.Fatalf("Failed conneting to gRPC: %s\n", err)
				}
				defer conn.Close()

				client := nerf.NewServerClient(conn)
				request := &nerf.Request{Token: &nerf.Cfg.Token, Login: &nerf.Cfg.Login}
				response, err := client.GetNebulaConfig(context.Background(), request)
				if err != nil {
					log.Fatalf("Failed calling remote gRPC: %s\n", err)
				}

				out, err := os.Create(path.Join(nerf.NebulaDir(), "config.yml"))
				if err != nil {
					log.Fatalf("Failed creating config for Nebula: %s\n", err)
				}

				if _, err := out.WriteString(*response.Config); err != nil {
					log.Fatalf("Failed writing config for Nebula: %s\n", err)
				}
				defer out.Close()

				if err := nerf.NebulaStart(); err != nil {
					log.Fatalf("Failed starting Nebula client: %s\n", err)
				}
			case <-quit.ClickedCh:
				systray.Quit()
				os.Exit(0)
				return
			}
		}
	}()
}

func onExit() {
	// Cleaning stuff here.
}
