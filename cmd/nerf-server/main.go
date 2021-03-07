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

	lightHouseNebulaIP := strings.Split(*lightHouse, ":")[0]
	lightHousePublicIP := strings.Split(*lightHouse, ":")[1]

	if err := net.ParseIP(lightHouseNebulaIP); err == nil {
		fmt.Println("Overlay IP address is not IPv4")
		flag.Usage()
		os.Exit(1)
	}

	if err := net.ParseIP(lightHousePublicIP); err == nil {
		fmt.Println("Public IP address is not IPv4")
		flag.Usage()
		os.Exit(1)
	}

	if *printUsage {
		flag.Usage()
		os.Exit(0)
	}

	nerf.Cfg = nerf.NewConfig()
	nerf.Cfg.Nebula.LightHouse.NebulaIP = lightHouseNebulaIP
	nerf.Cfg.Nebula.LightHouse.PublicIP = lightHousePublicIP

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
