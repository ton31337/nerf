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
    <div style="width:400px; margin:0 auto; margin-top: 10%; font-size: 30px; font-family: Times New Roman, Times, serif; text-align: center;">Authorized.</div>
    <div style="width:400px; margin:0 auto; font-size: 18px; font-family: Times New Roman, Times, serif; text-align: center;">You can close this window.</div>
</body>
</html>
`

var server *http.Server
var authCodeState = uuid.NewString()

// TokenSource defines Access Token for Github
type TokenSource struct {
	AccessToken string
}

// Token initializes Access Token for Github
func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

func handleAuthMain(w http.ResponseWriter, r *http.Request) {
	url := Cfg.OAuth.AuthCodeURL(authCodeState, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleAuthDone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(authorizedHTML)); err != nil {
		log.Fatalf("Failed writing response to a browser: %s\n", err)
	}
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
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		// Firefox does not work properly because:
		// % sudo xdg-open http://example.org
		// Running Firefox as root in a regular user's session is not supported.
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
			log.Fatalf("Failed starting a web server: %s\n", err)
		}
	}()

	fmt.Printf("Your browser has been opened to visit:\n\thttp://%s\n", Cfg.ListenAddr)

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Web server shutdown failed: %s\n", err)
	}
}
