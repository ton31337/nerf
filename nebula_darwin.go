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

	if err := exec.Command("/sbin/route", "-n", "add", "-net", e.RemoteIP, defaultGw).Run(); err != nil {
		Cfg.Logger.Error("Can't create a static route for gRPC server",
			zap.String("RemoteIP", e.RemoteIP),
			zap.Error(err))
		return err
	}

	return nil
}
