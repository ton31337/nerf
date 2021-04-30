package nerf

import (
	"net"
	"os/exec"
	"path"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
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

// NebulaSetNameServers set name server for the client to self
func NebulaSetNameServers(e *Endpoint, NameServer string) error {
	var err error

	routes, err := netlink.RouteGet(net.ParseIP(e.RemoteIP))
	if err != nil {
		Cfg.Logger.Error("Can't get route for gRPC server",
			zap.String("RemoteIP", e.RemoteIP),
			zap.Error(err))
		return err
	}

	device, err := netlink.LinkByIndex(routes[0].LinkIndex)
	if err != nil {
		Cfg.Logger.Error("Can't get default interface by ifIndex",
			zap.String("Dst", routes[0].Gw.String()),
			zap.Int("ifIndex", routes[0].LinkIndex),
			zap.Error(err))
		return err
	}

	cmd := exec.Command(
		"systemd-resolve",
		"--interface",
		device.Attrs().Name,
		"--set-dns",
		NameServer,
		"--set-domain",
		DNSAutoDiscoverZone,
	)
	err = cmd.Run()
	if err != nil {
		Cfg.Logger.Error("Can't set name servers",
			zap.String("NameServer1", Cfg.Nebula.LightHouse.NebulaIP),
			zap.String("Domain", DNSAutoDiscoverZone),
			zap.Error(err))
		return err
	}

	return err
}

// NebulaAddLightHouseStaticRoute add static route towards fastest gRPC server via default route
func NebulaAddLightHouseStaticRoute(e *Endpoint) error {
	routes, err := netlink.RouteGet(net.ParseIP(e.RemoteIP))
	if err != nil {
		Cfg.Logger.Error("Can't get route for gRPC server",
			zap.String("RemoteIP", e.RemoteIP),
			zap.Error(err))
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
		Cfg.Logger.Error("Can't delete a static route for gRPC server",
			zap.String("RemoteIP", e.RemoteIP),
			zap.Error(err))
	}

	return netlink.RouteAdd(&nr)
}
