package nerf

import (
	"net"
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
		return err
	}

	if err := netlink.RouteAdd(&nr); err != nil {
		Cfg.Logger.Error("Can't create a static route for gRPC server",
			zap.String("RemoteIP", e.RemoteIP),
			zap.Error(err))
		return err
	}

	return nil
}
