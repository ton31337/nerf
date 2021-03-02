package nerf

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const authorizedHtml = `
<html>
<body>
Authorized and connected to VPN. You can close this window.
</body>
</html>
`

var server *http.Server

func handleAuthMain(w http.ResponseWriter, r *http.Request) {
	url := Cfg.OAuth.AuthCodeURL(Cfg.Organization, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleAuthDone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(authorizedHtml))
	p, _ := os.FindProcess(os.Getpid())
	p.Signal(os.Interrupt)
}

func handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != Cfg.Organization {
		fmt.Printf("Invalid oauth state, expected '%s', got '%s'\n", Cfg.Organization, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := Cfg.OAuth.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	oauthClient := Cfg.OAuth.Client(oauth2.NoContext, token)
	client := github.NewClient(oauthClient)
	user, _, _ := client.Users.Get(oauth2.NoContext, "")
	teams, _, _ := client.Teams.ListUserTeams(oauth2.NoContext, nil)

	if teams != nil {
		Cfg.Email = *user.Email
		for _, team := range teams {
			if *team.Organization.Name == Cfg.Organization {
				Cfg.Teams = append(Cfg.Teams, *team.Name)
			}
		}
	}

	http.Redirect(w, r, "/done", http.StatusTemporaryRedirect)
}

func openBrowser(url string) error {
	var err error

	os := runtime.GOOS
	switch os {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	return err
}

// Auth handles OAuth authentication
func Auth() {
	router := http.NewServeMux()
	server = &http.Server{
		Addr:     Cfg.ListenAddr,
		Handler:  router,
		ErrorLog: nil,
	}

	router.HandleFunc("/", handleAuthMain)
	router.HandleFunc("/callback", handleAuthCallback)
	router.HandleFunc("/done", handleAuthDone)

	go func() {
		<-time.After(1000 * time.Millisecond)
		err := openBrowser("http://" + Cfg.ListenAddr)
		if err != nil {
			fmt.Println(err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		server.SetKeepAlivesEnabled(false)
		server.ListenAndServe()
	}()

	fmt.Print("Your browser has been opened to visit:\n\thttp://" + Cfg.ListenAddr + "\n")

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	server.Shutdown(ctx)
}
