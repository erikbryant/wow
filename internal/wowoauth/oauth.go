package wowoauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

var (
	// blizzardOAuthConfig stores the OAuth user config for authenticating with Blizzard
	blizzardOAuthConfig = &oauth2.Config{
		ClientID:     "", // Populated at runtime
		ClientSecret: "", // Populated at runtime
		Endpoint:     endpoints.Battlenet,
		RedirectURL:  "http://localhost:8888/auth/blizzard/profile",
		Scopes:       []string{"wow.profile", "sc2.profile"},
	}
	// server is a reference to the webserver
	server = &http.Server{}
	// paToken stores the last-known profile access token
	paToken = ""
)

const (
	// cookieName is the name of the OAuth cookie
	cookieName = "oauthState"
)

// generateStateOAuthCookie stores a unique identifier in a cookie and returns that same identifier
func generateStateOAuthCookie(w http.ResponseWriter) (string, error) {
	var expiration = time.Now().Add(20 * time.Minute)

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: cookieName, Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state, nil
}

// oAuthBlizzardLogin creates the auth cookie and redirects to the Blizzard auth server
func oAuthBlizzardLogin(w http.ResponseWriter, r *http.Request) {
	// Create oAuthState cookie
	oAuthState, err := generateStateOAuthCookie(w)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to create OAuth cookie:", err)
		os.Exit(1)
	}

	// AuthCodeURL takes a unique, private state token to protect the user from CSRF attacks.
	// You must always provide a non-empty string and validate it matches the state query
	// parameter on your redirect callback.
	u := blizzardOAuthConfig.AuthCodeURL(oAuthState)
	// My account is homed in the US. battle.net resolves to whatever local country. Force it to use 'us'.
	u = strings.Replace(u, "/battle.net/", "/us.battle.net/", 1)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func tokenToPAT(code string) (string, error) {
	data := url.Values{
		"redirect_uri": {blizzardOAuthConfig.RedirectURL},
		"grant_type":   {"authorization_code"},
		"code":         {code},
	}

	return GetToken(data, blizzardOAuthConfig.ClientID, blizzardOAuthConfig.ClientSecret)
}

// oAuthBlizzardCallback receives the token, converts it to a PAT, and passes that to the webpage requester
func oAuthBlizzardCallback(w http.ResponseWriter, r *http.Request) {
	if r == nil {
		log.Println("oAuthBlizzardCallback: empty request")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Read OAuth state from Cookie
	oAuthState, err := r.Cookie(cookieName)
	if err != nil {
		log.Println("oAuthBlizzardCallback: cookie error:", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if r.FormValue("state") != oAuthState.Value {
		log.Println("oAuthBlizzardCallback: invalid OAuth state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Exchange the token we got for an actual profile access token
	msg := "Success!"
	paToken, err = tokenToPAT(r.FormValue("code"))
	if err != nil {
		msg = fmt.Sprintf("Could not get a profile access token: %s", err)
	}
	msg += " You can close this window."
	w.Write([]byte(msg))
}

// handlers registers the OAuth endpoints
func handlers() http.Handler {
	mux := http.NewServeMux()
	// Root
	mux.Handle("/", http.FileServer(http.Dir("templates/")))

	// OAuth endpoints
	mux.HandleFunc("/auth/blizzard/login", oAuthBlizzardLogin)
	mux.HandleFunc("/auth/blizzard/profile", oAuthBlizzardCallback)

	return mux
}

// start starts the webserver
func start(clientID, clientSecret string) {
	blizzardOAuthConfig.ClientID = clientID
	blizzardOAuthConfig.ClientSecret = clientSecret

	server = &http.Server{
		Addr:    fmt.Sprintf(":8888"),
		Handler: handlers(),
	}

	//log.Printf("Starting HTTP Server. Listening at %v", server.Addr)
	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Printf("%v", err)
	}
}

// shutdown terminates the webserver
func shutdown() {
	err := server.Shutdown(context.Background())
	if err != nil {
		log.Printf("server shutdown failed: %v\n", err)
	}
}
