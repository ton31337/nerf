package nerf

import (
	"net"
	"os/exec"
	"path"
	"strings"

	"go.uber.org/zap"
)

func nebulaDownloadLink() string {
	return "https://github.com/hostinger/packages/releases/download/v1.0.0/nebula-1.3.0-darwin-amd64"
}

// NebulaDir absolute paths to the directory of Nebula configurations and binaries
func NebulaDir() string {
	return "/opt/nebula"
}

// NebulaExecutable show full path of Nebula executable
func NebulaExecutable() string {
	return path.Join(NebulaDir(), "nebula")
}

func nebulaDefaultGateway(e *Endpoint) (string, error) {
	var defaultGw string

	cmd := exec.Command("/sbin/route", "-n", "get", e.RemoteIP)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "gateway:" {
			ip := net.ParseIP(fields[1])
			if ip != nil {
				defaultGw = ip.String()
			}
		}
	}
	return defaultGw, nil
}

// NebulaSetNameServers set name server for the client to self
func NebulaSetNameServers(e *Endpoint, NameServer string) error {
	var err error
	var lines []byte

	cmd := exec.Command(
		"networksetup",
		"listallnetworkservices",
	)
	lines, err = cmd.CombinedOutput()
	if err != nil {
		Cfg.Logger.Error("Can't get network services",
			zap.Error(err))
		return err
	}

	services := strings.Split(string(lines), "\n")
	for _, service := range services {
		if !strings.Contains(service, "Wi-Fi") || !strings.Contains(service, "Thunderbolt") ||
			!strings.Contains(service, "USB") {
			continue
		}
		if exec.Command("networksetup", "-setdnsservers", service, NameServer).Run() != nil {
			Cfg.Logger.Error("Can't set name servers",
				zap.String("NameServer1", NameServer),
				zap.String("Domain", DNSAutoDiscoverZone),
				zap.Error(err))
			return err
		}
	}

	return err
}

// NebulaAddLightHouseStaticRoute add static route towards fastest gRPC server via default route
func NebulaAddLightHouseStaticRoute(e *Endpoint) error {
	defaultGw, err := nebulaDefaultGateway(e)
	if err != nil {
		Cfg.Logger.Error("Can't get route for gRPC server",
			zap.String("RemoteIP", e.RemoteIP),
			zap.Error(err))
		return err
	}

	if err := exec.Command("/sbin/route", "-n", "delete", "-net", e.RemoteIP).Run(); err != nil {
		Cfg.Logger.Error("Can't delete a static route for gRPC server",
			zap.String("RemoteIP", e.RemoteIP),
			zap.Error(err))
	}

	return exec.Command("/sbin/route", "-n", "add", "-net", e.RemoteIP, defaultGw).Run()
}
