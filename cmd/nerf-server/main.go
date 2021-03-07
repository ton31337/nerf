package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/ton31337/nerf"
	"google.golang.org/grpc"
)

func main() {
	lightHouse := flag.String("lighthouse", "", "Set the lighthouse. E.g.: <NebulaIP>:<PublicIP>")
	printUsage := flag.Bool("help", false, "Print command line usage")

	flag.Parse()

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

	if *printUsage {
		flag.Usage()
		os.Exit(0)
	}

	nerf.Cfg = nerf.NewConfig()
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
}
