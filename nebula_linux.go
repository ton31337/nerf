package nerf

import (
	"fmt"
	"net"
	"os/exec"
	"path"
	"strings"

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

func nebulaGetNameServers() error {
	var err error
	var nameServers []string

	cmd := exec.Command(
		"nmcli",
		"device",
		"show",
	)
	lines, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed 'nmcli device show'")
	}

	params := strings.Split(string(lines), "\n")
	for _, param := range params {
		if !strings.Contains(param, "IP4.DNS") {
			continue
		}

		ns := strings.TrimSpace(strings.Split(param, ":")[1])
		if ip := net.ParseIP(ns); ip != nil {
			nameServers = append(nameServers, ns)
		}
	}

	if len(nameServers) == 0 {
		return fmt.Errorf("failed retrieving current name servers")
	}

	Cfg.Logger.Debug("saving current name servers", zap.Strings("NameServers", nameServers))

	Cfg.SavedNameServers = nameServers

	return err
}

// NebulaSetNameServers set name server for the client to self
func NebulaSetNameServers(e *Endpoint, NameServers []string, save bool) error {
	var err error
	var dns []string

	if save {
		if err = nebulaGetNameServers(); err != nil {
			return err
		}
	}

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

	// systemd-resolve expects multiple `--set-dns X.X.X.X --set-dns Y.Y.Y.Y`,
	// thus crafting a slice for that.
	for _, ns := range NameServers {
		dns = append(dns, "--set-dns")
		dns = append(dns, ns)
	}

	cmd := exec.Command(
		"systemd-resolve",
	)
	cmd.Args = append([]string{
		"systemd-resolve",
		"--interface",
		device.Attrs().Name,
		"--set-domain",
		DNSAutoDiscoverZone,
	}, dns...)

	err = cmd.Run()
	if err != nil {
		Cfg.Logger.Error("Can't set name servers",
			zap.Strings("NameServers", NameServers),
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
