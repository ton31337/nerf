package nerf

import (
	"bufio"
	fmt "fmt"
	"net"
	"os"
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

func nebulaGetNameServers() error {
	var err error
	var nameServers []string

	file, err := os.Open("/etc/resolv.conf")
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 && line[0] == '#' {
			continue
		}

		f := strings.Fields(line)
		if len(f) < 1 {
			continue
		}

		if f[0] == "nameserver" {
			if net.ParseIP(f[1]) != nil {
				nameServers = append(nameServers, f[1])
			}
		}
	}

	if len(nameServers) == 0 {
		return fmt.Errorf("failed retrieving current name servers")
	}

	ClientCfg.Logger.Debug("saving current name servers", zap.Strings("NameServers", nameServers))

	ClientCfg.SavedNameServers = nameServers

	return err
}

// NebulaSetNameServers set name server for the client to self
func NebulaSetNameServers(e *Endpoint, NameServers []string, save bool) error {
	var err error

	if save {
		if err = nebulaGetNameServers(); err != nil {
			return err
		}
	}

	cmd := exec.Command(
		"networksetup",
		"listallnetworkservices",
	)
	lines, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed 'networksetup listallnetworkservices'")
	}

	for _, service := range strings.Split(string(lines), "\n") {
		if strings.Contains(service, "Wi-Fi") || strings.Contains(service, "USB") || strings.Contains(service, "Thunderbolt") ||
			strings.Contains(service, "Air") {
			set_dns := exec.Command(
				"networksetup",
			)
			set_dns.Args = append(
				[]string{"networksetup", "-setdnsservers", service},
				NameServers...)
			if set_dns.Run() != nil {
				ClientCfg.Logger.Error("can't set name servers",
					zap.Strings("NameServers", NameServers),
					zap.String("Domain", DNSAutoDiscoverZone),
					zap.Error(err))
				return err
			}
		}
	}

	ClientCfg.Logger.Debug("setting name servers", zap.Strings("NameServers", NameServers))

	return err
}

// NebulaAddLightHouseStaticRoute add static route towards fastest gRPC server via default route
func NebulaAddLightHouseStaticRoute(e *Endpoint) error {
	defaultGw, err := nebulaDefaultGateway(e)
	if err != nil {
		ClientCfg.Logger.Error("can't get route for gRPC server",
			zap.String("RemoteIP", e.RemoteIP),
			zap.Error(err))
		return err
	}

	if err := exec.Command("/sbin/route", "-n", "delete", "-net", e.RemoteIP).Run(); err != nil {
		ClientCfg.Logger.Error("can't delete a static route for gRPC server",
			zap.String("RemoteIP", e.RemoteIP),
			zap.Error(err))
	}

	return exec.Command("/sbin/route", "-n", "add", "-net", e.RemoteIP, defaultGw).Run()
}
