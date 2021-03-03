package nerf

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

func nebulaDownloadLink() string {
	os := runtime.GOOS
	switch os {
	case "windows":
		return "https://github.com/hostinger/packages/releases/download/v1.0.0/nebula-1.3.0-windows-amd64"
	case "darwin":
		return "https://github.com/hostinger/packages/releases/download/v1.0.0/nebula-1.3.0-darwin-amd64"
	case "linux":
		return "https://github.com/hostinger/packages/releases/download/v1.0.0/nebula-1.3.0-linux-amd64"
	default:
		return ""
	}
}

func nebulaDir() string {
	os := runtime.GOOS
	switch os {
	case "windows":
		return "C:/Nebula"
	case "darwin":
		return "/opt/nebula"
	case "linux":
		return "/opt/nebula"
	default:
		return ""
	}
}

// NebulaExecutable show full path of Nebula executable
func NebulaExecutable() string {
	os := runtime.GOOS
	switch os {
	case "windows":
		return path.Join(nebulaDir(), "nebula.exe")
	case "darwin":
		return path.Join(nebulaDir(), "nebula")
	case "linux":
		return path.Join(nebulaDir(), "nebula")
	default:
		return ""
	}
}

// NebulaGenerateCertificate generate ca.crt, client.crt, client.key for Nebula
func NebulaGenerateCertificate(groups []string, login string) (string, string, string) {
	crtPath := "/etc/nebula/certs/" + login + ".crt"
	keyPath := "/etc/nebula/certs/" + login + ".key"

	if _, err := os.Stat(crtPath); err == nil {
		os.Remove(crtPath)
	}

	if _, err := os.Stat(keyPath); err == nil {
		os.Remove(keyPath)
	}

	err := exec.Command("/usr/local/bin/nebula-cert",
		"sign", "-name", login,
		"-out-crt", crtPath,
		"-out-key", keyPath,
		"-ca-crt", "/etc/nebula/certs/ca.crt",
		"-ca-key", "/etc/nebula/certs/ca.key",
		"-ip", "172.17.0.2/12", "-groups", strings.Join(groups, ","),
		"-duration", "48h").Run()
	if err != nil {
		log.Fatalf("Failed generating certificate for Nebula: %v", err)
	}

	ca, err := ioutil.ReadFile("/etc/nebula/certs/ca.crt")
	if err != nil {
		log.Fatalf("Failed retrieving CA certificate: %v", err)
	}

	crt, err := ioutil.ReadFile(crtPath)
	if err != nil {
		log.Fatalf("Failed retrieving client certificate: %v", err)
	}

	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Fatalf("Failed retrieving client key: %v", err)
	}

	return string(ca), string(crt), string(key)
}

// NebulaDownload used to download Nebula binary
func NebulaDownload() (err error) {
	err = os.Mkdir(nebulaDir(), 0755)
	if err != nil {
		return err
	}

	out, err := os.Create(NebulaExecutable())
	if err != nil {
		return err
	}
	defer out.Close()

	err = os.Chmod(NebulaExecutable(), 0755)
	if err != nil {
		return err
	}

	resp, err := http.Get(nebulaDownloadLink())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed download Nebula: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
