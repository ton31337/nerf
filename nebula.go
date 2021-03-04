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
	"text/template"

	"github.com/Masterminds/sprig"
)

// Certificate struct for certificates received from Nebula
type Certificate struct {
	Ca  string
	Crt string
	Key string
}

// NewCertificate stores ca.crt, client.crt, client.key
func NewCertificate(Ca string, Crt string, Key string) *Certificate {
	return &Certificate{
		Ca:  Ca,
		Crt: Crt,
		Key: Key,
	}
}

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

// NebulaStart starts Nebula instance in foreground
func NebulaStart() error {
	return exec.Command(NebulaExecutable(), "-config", path.Join(nebulaDir(), "config.yml")).Run()
}

// NebulaGenerateConfig generate config.yml
func NebulaGenerateConfig(cert *Certificate) error {
	configTemplate := `# Generated by Nerf!

pki:
  ca: |
{{ .Ca | indent 4 }}
  cert: |
{{ .Crt | indent 4 }}
  key: |
{{ .Key | indent 4 }}
static_host_map:
  "172.16.209.111": ["153.92.3.153:4242"]

lighthouse:
  am_lighthouse: false
  interval: 60
  hosts:
    - 172.16.209.111

listen:
  host: 0.0.0.0
  port: 4242

local_range: 172.16.0.0/12

tun:
  disabled: false
  dev: nebula1

firewall:
  outbound:
    - port: any
      proto: any
      host: any
`
	nebulaConfigTemplate := template.New("Nebula configuration file")
	_, err := nebulaConfigTemplate.Funcs(sprig.HermeticTxtFuncMap()).Parse(configTemplate)
	if err != nil {
		return err
	}

	out, err := os.Create(path.Join(nebulaDir(), "config.yml"))
	if err != nil {
		return err
	}
	defer out.Close()

	return nebulaConfigTemplate.Execute(out, cert)
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
		log.Fatalf("Failed generating certificate for Nebula: %v\n", err)
	}

	ca, err := ioutil.ReadFile("/etc/nebula/certs/ca.crt")
	if err != nil {
		log.Fatalf("Failed retrieving CA certificate: %v\n", err)
	}

	crt, err := ioutil.ReadFile(crtPath)
	if err != nil {
		log.Fatalf("Failed retrieving client certificate: %v\n", err)
	}

	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Fatalf("Failed retrieving client key: %v\n", err)
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
