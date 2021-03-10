package nerf

import (
	"os/exec"
	"path"
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
	return exec.Command("pkexec", NebulaExecutable(), "-config", path.Join(NebulaDir(), "config.yml")).Run()
}
