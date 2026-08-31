package wowoauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	cookieName  = "oauthState"
	redirectURL = "http://localhost:8888/auth/blizzard/profile"

	// Blizzard's Battle.net OAuth endpoints.
	authorizeURL = "https://oauth.battle.net/authorize"
	tokenURL     = "https://oauth.battle.net/token"

	// OAuth should never wait indefinitely for the user.
	oauthTimeout = 5 * time.Minute
)

var openBrowser = defaultOpenBrowser

func defaultOpenBrowser(url string) error {
	cmd := exec.Command("open", url)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// oauthFlow contains everything associated with one OAuth attempt.
//
// Nothing related to an individual authentication attempt is stored in
// package-global state.
type oauthFlow struct {
	clientID     string
	clientSecret string

	result chan oauthResult
}

type oauthResult struct {
	token string
	err   error
}

// generateStateOAuthCookie generates a cryptographically random OAuth state
// value, stores it in a cookie, and returns it.
func generateStateOAuthCookie(w http.ResponseWriter) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate OAuth state: %w", err)
	}

	state := base64.RawURLEncoding.EncodeToString(b)

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    state,
		Path:     "/auth/blizzard",
		Expires:  time.Now().Add(oauthTimeout),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // localhost callback uses HTTP
	}

	http.SetCookie(w, cookie)

	return state, nil
}

// login redirects the browser to Blizzard's OAuth authorization endpoint.
func (f *oauthFlow) login(w http.ResponseWriter, r *http.Request) {
	state, err := generateStateOAuthCookie(w)
	if err != nil {
		http.Error(w, "Unable to start authentication.", http.StatusInternalServerError)
		return
	}

	u, err := url.Parse(authorizeURL)
	if err != nil {
		http.Error(w, "Unable to construct authentication URL.", http.StatusInternalServerError)
		return
	}

	q := u.Query()
	q.Set("client_id", f.clientID)
	q.Set("redirect_uri", redirectURL)
	q.Set("response_type", "code")
	q.Set("scope", "wow.profile")
	q.Set("state", state)
	u.RawQuery = q.Encode()

	// Battle.net may regionalize the OAuth hostname based on the user's
	// current location. This application uses a US-homed Battle.net account,
	// so force the OAuth endpoint to the US region.
	u.Path = strings.Replace(u.Path, "/battle.net/", "/us.battle.net/", 1)

	http.Redirect(w, r, u.String(), http.StatusTemporaryRedirect)
}

// callback receives Blizzard's authorization response.
func (f *oauthFlow) callback(w http.ResponseWriter, r *http.Request) {
	if r == nil {
		f.fail(fmt.Errorf("OAuth callback received nil request"))
		return
	}

	stateCookie, err := r.Cookie(cookieName)
	if err != nil {
		f.writeFailure(w, fmt.Errorf("OAuth callback missing state cookie: %w", err))
		return
	}

	state := r.URL.Query().Get("state")
	if state == "" || state != stateCookie.Value {
		f.writeFailure(w, fmt.Errorf("OAuth callback has invalid state"))
		return
	}

	// Blizzard can return an OAuth error instead of an authorization code.
	if oauthError := r.URL.Query().Get("error"); oauthError != "" {
		description := r.URL.Query().Get("error_description")
		if description != "" {
			f.writeFailure(w, fmt.Errorf("blizzard OAuth error: %s: %s", oauthError, description))
		} else {
			f.writeFailure(w, fmt.Errorf("blizzard OAuth error: %s", oauthError))
		}
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		f.writeFailure(w, fmt.Errorf("OAuth callback missing authorization code"))
		return
	}

	token, err := tokenToPAT(code, f.clientID, f.clientSecret)
	if err != nil {
		f.writeFailure(w, fmt.Errorf("get profile access token: %w", err))
		return
	}

	// Clear the OAuth state cookie now that it has been consumed.
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/auth/blizzard",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   false,
	})

	f.succeed(token)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Success! You can close this window."))
}

func (f *oauthFlow) succeed(token string) {
	select {
	case f.result <- oauthResult{token: token}:
	default:
	}
}

func (f *oauthFlow) fail(err error) {
	select {
	case f.result <- oauthResult{err: err}:
	default:
	}
}

func (f *oauthFlow) writeFailure(w http.ResponseWriter, err error) {
	f.fail(err)

	http.Error(
		w,
		fmt.Sprintf("Authentication failed: %s. You can close this window.", err),
		http.StatusBadRequest,
	)
}

func (f *oauthFlow) handlers() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/auth/blizzard/login", f.login)
	mux.HandleFunc("/auth/blizzard/profile", f.callback)

	return mux
}

// GetPAT runs a complete Blizzard OAuth authorization-code flow.
//
// The local server exists only for the duration of this call. All state
// belonging to this authentication attempt is local to the call.
func GetPAT(clientID, clientSecret string) (string, error) {
	flow := &oauthFlow{
		clientID:     clientID,
		clientSecret: clientSecret,
		result:       make(chan oauthResult, 1),
	}

	server := &http.Server{
		Addr:    "localhost:8888",
		Handler: flow.handlers(),
	}

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return "", fmt.Errorf("listen on %s: %w", server.Addr, err)
	}

	serveErr := make(chan error, 1)

	go func() {
		err := server.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
		close(serveErr)
	}()
	defer shutdownServer(server)

	if err := openBrowser("http://localhost:8888/auth/blizzard/login"); err != nil {
		return "", fmt.Errorf("unable to open browser: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), oauthTimeout)
	defer cancel()

	var result oauthResult

	select {
	case result = <-flow.result:
		// OAuth completed.

	case err := <-serveErr:
		if err != nil {
			return "", fmt.Errorf("OAuth server failed: %w", err)
		}
		return "", fmt.Errorf("OAuth server stopped unexpectedly")

	case <-ctx.Done():
		return "", fmt.Errorf("OAuth authentication timed out after %s", oauthTimeout)
	}

	if result.err != nil {
		return "", result.err
	}

	return result.token, nil
}

func shutdownServer(server *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return server.Shutdown(ctx)
}
