package nerf

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
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

func nebulaExecutable() string {
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

// NebulaDownload used to download Nebula binary
func NebulaDownload() (err error) {
	err = os.Mkdir(nebulaDir(), 0755)
	if err != nil {
		return err
	}

	out, err := os.Create(nebulaExecutable())
	if err != nil {
		return err
	}
	defer out.Close()

	err = os.Chmod(nebulaExecutable(), 0755)
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
