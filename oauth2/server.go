package oauth2

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	// blizzardOauthConfig stores the OAUTH config for authenticating with Blizzard
	blizzardOauthConfig = &oauth2.Config{
		ClientID:     "", // Populated at runtime
		ClientSecret: "", // Populated at runtime
		Endpoint:     endpoints.Battlenet,
		RedirectURL:  "http://localhost:8888/auth/blizzard/profile",
		Scopes:       []string{"wow.profile", "sc2.profile"},
	}
	// server is a reference to the webserver
	server = &http.Server{}
)

const (
	// cookieName is the name of the OAUTH cookie
	cookieName = "oauthState"
)

// generateStateOauthCookie stores a unique identifier in a cookie and returns that same identifier
func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(20 * time.Minute)

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: cookieName, Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

// oauthBlizzardLogin creates the auth cookie and redirects to the Blizzard auth server
func oauthBlizzardLogin(w http.ResponseWriter, r *http.Request) {
	// Create oauthState cookie
	oauthState := generateStateOauthCookie(w)

	// AuthCodeURL takes a unique, private state token to protect the user from CSRF attacks.
	// You must always provide a non-empty string and validate it matches the state query
	// parameter on your redirect callback.
	u := blizzardOauthConfig.AuthCodeURL(oauthState)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func tokenToPAT(code string) string {
	client := &http.Client{}
	data := url.Values{
		"redirect_uri": {blizzardOauthConfig.RedirectURL},
		"grant_type":   {"authorization_code"},
		"code":         {code},
	}
	request, err := http.NewRequest("POST", "https://oauth.battle.net/token", strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(blizzardOauthConfig.ClientID, blizzardOauthConfig.ClientSecret)
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var jsonObject map[string]interface{}

	err = json.Unmarshal(contents, &jsonObject)
	if err != nil {
		log.Fatal(err)
	}

	return jsonObject["access_token"].(string)
}

// oauthBlizzardCallback receives the token, converts it to a PAT, and passes that to the webpage requester
func oauthBlizzardCallback(w http.ResponseWriter, r *http.Request) {
	// Read oauthState from Cookie
	oauthState, _ := r.Cookie(cookieName)

	if r == nil {
		log.Println("oauthBlizzardCallback: Empty request")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if r.FormValue("state") != oauthState.Value {
		log.Println("invalid oauth state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// Exchange the token we got for an actual profile access token
	pat := tokenToPAT(r.FormValue("code"))
	w.Write([]byte("success!\n"))
	w.Write([]byte(pat))
}

// handlers registers the OAUTH endpoints
func handlers() http.Handler {
	mux := http.NewServeMux()
	// Root
	mux.Handle("/", http.FileServer(http.Dir("templates/")))

	// OAUTH endpoints
	mux.HandleFunc("/auth/blizzard/login", oauthBlizzardLogin)
	mux.HandleFunc("/auth/blizzard/profile", oauthBlizzardCallback)

	return mux
}

// Start starts the webserver
func Start(clientID, clientSecret string) {
	blizzardOauthConfig.ClientID = clientID
	blizzardOauthConfig.ClientSecret = clientSecret

	server = &http.Server{
		Addr:    fmt.Sprintf(":8888"),
		Handler: handlers(),
	}

	log.Printf("Starting HTTP Server. Listening at %v", server.Addr)
	err := server.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		log.Printf("%v", err)
	}
}

// Shutdown terminates the webserver
func Shutdown() {
	err := server.Shutdown(context.Background())
	if err != nil {
		log.Printf("server shutdown failed: %v\n", err)
	}
}
