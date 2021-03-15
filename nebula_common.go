package nerf

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/snksoft/crc"
	"go.uber.org/zap"
)

// Certificate struct for certificates generated for Nebula
type Certificate struct {
	Ca  string
	Crt string
	Key string
}

// LightHouse struct to define Nebula internal (overlay) IP address,
// and public (how to reach the real host in the mesh) IP address.
type LightHouse struct {
	NebulaIP string
	PublicIP string
}

// Nebula struct to store all the relevant data to generate config.yml for Nebula
type Nebula struct {
	Certificate *Certificate
	Subnet      string
	LightHouse  *LightHouse
}

// NewCertificate stores ca.crt, client.crt, client.key
func NewCertificate(Ca string, Crt string, Key string) *Certificate {
	return &Certificate{
		Ca:  Ca,
		Crt: Crt,
		Key: Key,
	}
}

// NebulaStart starts Nebula instance in foreground
func NebulaStart() error {
	return exec.Command(NebulaExecutable(), "-config", path.Join(NebulaDir(), "config.yml")).Run()
}

// NebulaGenerateConfig generate config.yml
func NebulaGenerateConfig(userTeams []string) (string, error) {
	var generatedConfig bytes.Buffer

	NebulaGenerateCertificate(userTeams)

	configTemplate := `# Generated by Nerf!

pki:
  ca: |
{{ .Certificate.Ca | indent 4 }}
  cert: |
{{ .Certificate.Crt | indent 4 }}
  key: |
{{ .Certificate.Key | indent 4 }}
static_host_map:
  "{{ .LightHouse.NebulaIP }}": ["{{ .LightHouse.PublicIP }}:4242"]

lighthouse:
  am_lighthouse: false
  interval: 60
  hosts:
    - {{ .LightHouse.NebulaIP }}

listen:
  host: 0.0.0.0
  port: 4242

local_range: {{ .Subnet }}

tun:
  disabled: false
  dev: nebula1
  unsafe_routes:
    - route: 0.0.0.0/1
      via: {{ .LightHouse.NebulaIP }}
    - route: 128.0.0.0/1
      via: {{ .LightHouse.NebulaIP }}

firewall:
  outbound:
    - port: any
      proto: any
      host: any
`
	nebulaConfigTemplate := template.New("Nebula configuration file")
	_, err := nebulaConfigTemplate.Funcs(sprig.HermeticTxtFuncMap()).Parse(configTemplate)
	if err != nil {
		return "", err
	}

	if err := nebulaConfigTemplate.Execute(&generatedConfig, Cfg.Nebula); err != nil {
		return "", err
	}

	return generatedConfig.String(), nil
}

func nebulaIP2Int(ip string) uint32 {
	var long uint32
	if err := binary.Read(bytes.NewBuffer(net.ParseIP(ip).To4()), binary.BigEndian, &long); err != nil {
		log.Fatalf("Failed converting Nebula IP to integer: %s\n", err)
	}
	return long
}

func int2NebulaIP(ip int64) string {
	b0 := strconv.FormatInt((ip>>24)&0xff, 10)
	b1 := strconv.FormatInt((ip>>16)&0xff, 10)
	b2 := strconv.FormatInt((ip>>8)&0xff, 10)
	b3 := strconv.FormatInt((ip & 0xff), 10)
	return b0 + "." + b1 + "." + b2 + "." + b3
}

// NebulaClientIP returns client's IP generated from Github login
func NebulaClientIP() string {
	clientIPHash := crc.CalculateCRC(crc.CCITT, []byte(Cfg.Login))
	clientIP := int64(nebulaIP2Int(nebulaSubnet()) + uint32(clientIPHash))
	return int2NebulaIP(clientIP)
}

func nebulaSubnet() string {
	return strings.Split(Cfg.Nebula.Subnet, "/")[0]
}

func nebulaSubnetLen() string {
	return strings.Split(Cfg.Nebula.Subnet, "/")[1]
}

// NebulaGenerateCertificate generate ca.crt, client.crt, client.key for Nebula
func NebulaGenerateCertificate(userTeams []string) {
	crtPath := "/etc/nebula/certs/" + Cfg.Login + ".crt"
	keyPath := "/etc/nebula/certs/" + Cfg.Login + ".key"

	if _, err := os.Stat(crtPath); err == nil {
		os.Remove(crtPath)
	}

	if _, err := os.Stat(keyPath); err == nil {
		os.Remove(keyPath)
	}

	err := exec.Command("/usr/local/nebula/nebula-cert",
		"sign", "-name", Cfg.Login,
		"-out-crt", crtPath,
		"-out-key", keyPath,
		"-ca-crt", "/etc/nebula/certs/ca.crt",
		"-ca-key", "/etc/nebula/certs/ca.key",
		"-ip", NebulaClientIP()+"/"+nebulaSubnetLen(), "-groups", strings.Join(userTeams, ","),
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

	Cfg.Nebula.Certificate.Ca = string(ca)
	Cfg.Nebula.Certificate.Crt = string(crt)
	Cfg.Nebula.Certificate.Key = string(key)
}

// NebulaDownload used to download Nebula binary
func NebulaDownload() (err error) {
	err = os.Mkdir(NebulaDir(), 0755)
	if err != nil {
		return err
	}

	out, err := os.Create(NebulaExecutable())
	if err != nil {
		Cfg.Logger.Error("Can't create Nebula binary",
			zap.String("Path", NebulaExecutable()),
			zap.Error(err))
		return err
	}
	defer out.Close()

	err = os.Chmod(NebulaExecutable(), 0755)
	if err != nil {
		Cfg.Logger.Error("Can't change permissions for Nebula binary",
			zap.String("Path", NebulaExecutable()),
			zap.Error(err))
		return err
	}

	resp, err := http.Get(nebulaDownloadLink())
	if err != nil {
		Cfg.Logger.Error("Can't download Nebula binary",
			zap.String("Url", nebulaDownloadLink()),
			zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed download Nebula: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		Cfg.Logger.Error("Can't write Nebula binary",
			zap.String("Url", nebulaDownloadLink()),
			zap.String("Path", NebulaExecutable()),
			zap.Error(err))
		return err
	}

	return nil
}
