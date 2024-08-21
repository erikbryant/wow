package oauth2

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
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
	blizzardOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8888/auth/blizzard/profile",
		ClientID:     "",
		ClientSecret: "",
		Scopes:       []string{"wow.profile", "sc2.profile"},
		Endpoint:     endpoints.Battlenet,
	}
	server = &http.Server{}
)

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(20 * time.Minute)

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

func oauthBlizzardLogin(w http.ResponseWriter, r *http.Request) {
	// Create oauthState cookie
	oauthState := generateStateOauthCookie(w)

	// AuthCodeURL takes a unique, private state token to protect the user from CSRF attacks.
	// You must always provide a non-empty string and validate it matches the state query
	// parameter on your redirect callback.
	u := blizzardOauthConfig.AuthCodeURL(oauthState)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func oauthBlizzardCallback(w http.ResponseWriter, r *http.Request) {
	// Read oauthState from Cookie
	oauthState, _ := r.Cookie("oauthstate")

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

	client := &http.Client{}
	data := url.Values{
		"redirect_uri": {blizzardOauthConfig.RedirectURL},
		"grant_type":   {"authorization_code"},
		"code":         {r.FormValue("code")},
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

	profileAccessToken := jsonObject["access_token"].(string)
	w.Write([]byte("success!\n"))
	w.Write([]byte(profileAccessToken))
}

func handlers() http.Handler {
	mux := http.NewServeMux()
	// Root
	mux.Handle("/", http.FileServer(http.Dir("templates/")))

	// OAUTH endpoints
	mux.HandleFunc("/auth/blizzard/login", oauthBlizzardLogin)
	mux.HandleFunc("/auth/blizzard/profile", oauthBlizzardCallback)

	return mux
}

func Start(clientID, clientSecret string) {
	blizzardOauthConfig.ClientID = clientID
	blizzardOauthConfig.ClientSecret = clientSecret

	server = &http.Server{
		Addr:    fmt.Sprintf(":8888"),
		Handler: handlers(),
	}

	log.Printf("Starting HTTP Server. Listening at %v", server.Addr)
	err := server.ListenAndServe()
	log.Printf("%v", err)
}

func Shutdown() {
	err := server.Shutdown(context.Background())
	if err != nil {
		log.Printf("server shutdown failed: %v\n", err)
	}
}
