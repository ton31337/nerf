package main

import (
	"context"
	"log"
	"os"

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
	nerf.Auth()

	conn, err := grpc.Dial(":9000", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed conneting to gRPC: %s", err)
	}
	defer conn.Close()

	client := nerf.NewServerClient(conn)
	request := &nerf.Request{Token: &nerf.Cfg.Token, Login: &nerf.Cfg.Login}
	response, err := client.GetCertificates(context.Background(), request)
	if err != nil {
		log.Fatalf("Failed calling remote gRPC: %s", err)
	}

	nerf.Cfg.Certificate = nerf.NewCertificate(*response.Ca, *response.Crt, *response.Key)
	if err := nerf.NebulaGenerateConfig(nerf.Cfg.Certificate); err != nil {
		log.Fatalf("Failed creating configuration file for Nebula: %s", err)
	}
}
