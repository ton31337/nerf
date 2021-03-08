package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path"
	"strings"

	"github.com/getlantern/systray"
	"github.com/ton31337/nerf"
	"google.golang.org/grpc"
)

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
}

func main() {
	server := flag.Bool("server", false, "Start gRPC server to generate config for Nebula")
	lightHouse := flag.String("lighthouse", "", "Set the lighthouse. E.g.: <NebulaIP>:<PublicIP>")
	printUsage := flag.Bool("help", false, "Print command line usage")

	flag.Parse()
	nerf.Cfg = nerf.NewConfig()

	if *server {
		if *lightHouse == "" {
			fmt.Println("-lighthouse flag must be set")
			flag.Usage()
			os.Exit(1)
		}

		lightHouseIPS := strings.Split(*lightHouse, ":")
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

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9000))
		if err != nil {
			log.Fatalf("Failed to listen: %v\n", err)
		}

		grpcServer := grpc.NewServer()
		nerf.RegisterServerServer(grpcServer, &nerf.Server{})

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v\n", err)
		}
	} else {
		err := nerf.NebulaDownload()
		if err != nil {
			if _, err := os.Stat(nerf.NebulaExecutable()); err != nil {
				log.Fatalf("Failed installing Nebula: %s\n", err)
			}
		}

		systray.Run(onReady, onExit)
	}

	if *printUsage {
		flag.Usage()
		os.Exit(0)
	}
}
