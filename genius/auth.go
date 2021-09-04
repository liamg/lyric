package genius

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/skratchdot/open-golang/open"
)

const (
	apiHost  = "api.genius.com"
	clientID = "YUzgT8ekyc89yyG3JDYGInAvmQ2vBg_b54QhbyQGlgL_iZa0fB2CuodWxhcYkLrH"
	scope    = "me"
)

var authTimeoutDuration = time.Second * 30

type AccessToken string

func testToken(token AccessToken) bool {
	_, err := NewClient(token).get("/account", false)
	return err != nil
}

func saveTokenToCache(token AccessToken) error {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(dirname, ".lyric"), []byte(token), 0600)
}

func loadCachedToken() (AccessToken, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadFile(filepath.Join(dirname, ".lyric"))
	if err != nil {
		return "", err
	}
	return AccessToken(string(data)), nil
}

func Authenticate() (AccessToken, error) {

	token, err := loadCachedToken()
	if err == nil {
		if testToken(token) {
			return token, nil
		}
	}

	fmt.Println("Authentication required. Check your browser.")

	redirectUrl := "http://localhost:62626"

	stateBytes := make([]byte, 32)
	if _, err := rand.Reader.Read(stateBytes); err != nil {
		return token, err
	}
	state := hex.EncodeToString(stateBytes)
	tokenChan := make(chan AccessToken)
	errorChan := make(chan error)

	srv := &http.Server{Addr: ":62626"}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<script>document.location=document.location.href.replace("#","receive?");</script>`))
	})
	http.HandleFunc("/receive", func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.URL.Query().Get("access_token")
		actualState := r.URL.Query().Get("state")
		if actualState != state {
			errorChan <- fmt.Errorf("authentication failed due to invalid state - please try again")
			w.Write([]byte("Authentication error - please go back to your terminal and try again."))
			return
		}
		tokenChan <- AccessToken(accessToken)
		w.Write([]byte("You can now close this tab and check your terminal."))
	})

	defer srv.Close()
	go srv.ListenAndServe()

	url := fmt.Sprintf(
		"https://%s/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s&state=%s&response_type=token",
		apiHost,
		clientID,
		redirectUrl,
		scope,
		state,
	)

	if err := open.Run(url); err != nil {
		return token, err
	}

	select {
	case <-time.After(authTimeoutDuration):
		return token, fmt.Errorf("authentication flow timed out")
	case err := <-errorChan:
		return token, err
	case token = <-tokenChan:
		fmt.Println("Authentication complete!")
		_ = saveTokenToCache(token)
		return token, nil
	}
}
