package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ton31337/nerf"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	server := flag.Bool("server", false, "Start gRPC server to generate config for Nebula")
	lightHouse := flag.String("lighthouse", "", "Set the lighthouse. E.g.: <NebulaIP>:<PublicIP>")
	verbose := flag.Bool("verbose", false, "Print verbose output")
	printUsage := flag.Bool("help", false, "Print command line usage")

	flag.Parse()
	nerf.Cfg = nerf.NewConfig()

	logger, _ := zap.NewProduction()
	nerf.Cfg.Logger = logger
	defer func() {
		if err := nerf.Cfg.Logger.Sync(); err != nil {
			log.Fatalln("Failed Logger sync")
		}
	}()

	if *verbose {
		nerf.Cfg.Verbose = true
	}

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

		if nerf.Cfg.Verbose {
			nerf.Cfg.Logger.Info("Nerf server started")
		}

		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 9000))
		if err != nil {
			log.Fatalf("Failed to listen gRPC server: %v\n", err)
		}

		grpcServer := grpc.NewServer()
		nerf.RegisterServerServer(grpcServer, &nerf.Server{})

		go func() {
			for range time.Tick(10 * time.Second) {
				if (time.Now().Unix() - nerf.Cfg.Teams.UpdatedAt) > 3600 {
					if nerf.Cfg.Verbose {
						nerf.Cfg.Logger.Info(
							"Begin-Of-Sync Github Teams with local cache")
					}
					nerf.SyncTeams()
					if nerf.Cfg.Verbose {
						nerf.Cfg.Logger.Info(
							"End-Of-Sync Github Teams with local cache")
					}
				}
			}
		}()

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC server: %v\n", err)
		}
	} else {
		err := nerf.NebulaDownload()
		if err != nil {
			if _, err := os.Stat(nerf.NebulaExecutable()); err != nil {
				log.Fatalf("Failed installing Nebula: %s\n", err)
			}
		}

		// Before probing all gRPC endpoints we MUST be authenticated.
		// Otherwise, we can't continue and nothing happens.
		nerf.Auth()
		e := nerf.GetFastestEndpoint()
		if e.RemoteHost == "" {
			log.Fatalln("No available gRPC endpoints found")
		}

		if err := nerf.NebulaAddLightHouseStaticRoute(&e); err != nil {
			log.Fatalf("Failed creating a static route to %s: %s\n", e.RemoteIP, err)
		}

		if nerf.Cfg.Verbose {
			nerf.Cfg.Logger.Info("Authorized", zap.String("Login", nerf.Cfg.Login))
			nerf.Cfg.Logger.Info("Using fastest gRPC endpoint",
				zap.String("RemoteIP", e.RemoteIP),
				zap.String("RemoteHost", e.RemoteHost),
				zap.String("Description", e.Description))
		}

		conn, err := grpc.Dial(e.RemoteHost+":9000", grpc.WithInsecure())
		if err != nil {
			log.Fatalf("Failed conneting to gRPC %s(%s): %s\n", e.RemoteHost, e.Description, err)
		}
		defer conn.Close()

		client := nerf.NewServerClient(conn)
		request := &nerf.Request{Token: &nerf.Cfg.Token, Login: &nerf.Cfg.Login}
		response, err := client.GetNebulaConfig(context.Background(), request)
		if err != nil {
			log.Fatalf("Failed calling remote gRPC %s(%s): %s\n", e.RemoteHost, e.Description, err)
		}

		if nerf.Cfg.Verbose {
			nerf.Cfg.Logger.Info("Connected",
				zap.String("RemoteIP", e.RemoteIP),
				zap.String("RemoteHost", e.RemoteHost),
				zap.String("Description", e.Description),
				zap.String("ClientIP", *response.ClientIP),
				zap.Strings("Teams", response.Teams))
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
	}

	if *printUsage {
		flag.Usage()
		os.Exit(0)
	}
}
