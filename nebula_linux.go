package nerf

import (
	"net"
	"os/exec"
	"path"

	"github.com/vishvananda/netlink"
)

func nebulaDownloadLink() string {
	return "https://github.com/hostinger/packages/releases/download/v1.0.0/nebula-1.3.0-linux-amd64"
}

// NebulaDir absolute paths to the directory of Nebula configurations and binaries
func NebulaDir() string {
	return "/opt/nebula"
}

// NebulaExecutable show full path of Nebula executable
func NebulaExecutable() string {
	return path.Join(NebulaDir(), "nebula")
}

// NebulaStart starts Nebula instance in foreground
func NebulaStart() error {
	return exec.Command(NebulaExecutable(), "-config", path.Join(NebulaDir(), "config.yml")).Run()
}

// NebulaAddLightHouseStaticRoute add static route towards fastest gRPC server via default route
func NebulaAddLightHouseStaticRoute(e *Endpoint) error {
	routes, err := netlink.RouteGet(net.ParseIP(e.RemoteIP))
	if err != nil {
		return err
	}

	dst := &net.IPNet{
		IP:   net.ParseIP(e.RemoteIP),
		Mask: net.CIDRMask(32, 32),
	}

	nr := netlink.Route{
		LinkIndex: routes[0].LinkIndex,
		Dst:       dst,
		Gw:        routes[0].Gw,
	}

	if err := netlink.RouteDel(&nr); err != nil {
		return err
	}

	if err := netlink.RouteAdd(&nr); err != nil {
		return err
	}

	return nil
}
