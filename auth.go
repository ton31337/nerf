package nerf

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"time"

	"github.com/google/go-github/github"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

const authorizedHTML = `
<html>
<body>
Authorized. You can close this window.
</body>
</html>
`

var server *http.Server
var authCodeState = uuid.NewString()

func handleAuthMain(w http.ResponseWriter, r *http.Request) {
	url := Cfg.OAuth.AuthCodeURL(authCodeState, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleAuthDone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(authorizedHTML))
	p, _ := os.FindProcess(os.Getpid())
	if err := p.Signal(os.Interrupt); err != nil {
		log.Fatalf("Failed shutting down a web server: %s\n", err)
	}
}

func handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != authCodeState {
		fmt.Printf("Invalid oauth state, expected '%s', got '%s'\n", authCodeState, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := Cfg.OAuth.Exchange(context.Background(), code)
	if err != nil {
		fmt.Printf("OAuth Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	oauthClient := Cfg.OAuth.Client(context.Background(), token)
	client := github.NewClient(oauthClient)
	user, _, _ := client.Users.Get(context.Background(), "")

	Cfg.Token = token.AccessToken
	Cfg.Login = *user.Login
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
	signal.Notify(done, os.Interrupt)

	go func() {
		server.SetKeepAlivesEnabled(false)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Failed starting a web server: %s", err)
		}
	}()

	fmt.Print("Your browser has been opened to visit:\n\thttp://" + Cfg.ListenAddr + "\n")

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Web server shutdown failed: %s\n", err)
	}
}
