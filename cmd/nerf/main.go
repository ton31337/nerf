package main

import (
	"fmt"

	"github.com/ton31337/nerf"
)

func main() {
	err := nerf.NebulaDownload()
	if err != nil {
		fmt.Printf("Failed to install Nebula: %s\n", err)
	}

	nerf.Cfg = nerf.NewNerfConfig()
	nerf.Auth()

	if nerf.Cfg.Email != "" {
		fmt.Printf("email: %s\n", nerf.Cfg.Email)
		for _, team := range nerf.Cfg.Teams {
			fmt.Printf("team: %s\n", team)
		}
	}
}
