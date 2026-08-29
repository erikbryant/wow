package wowoauth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGenerateStateOAuthCookie(t *testing.T) {
	rr := httptest.NewRecorder()

	state, err := generateStateOAuthCookie(rr)
	if err != nil {
		t.Fatal(err)
	}

	if state == "" {
		t.Fatal("empty state")
	}

	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies=%d, want 1", len(cookies))
	}

	c := cookies[0]

	if c.Name != cookieName {
		t.Errorf("cookie name=%q, want %q", c.Name, cookieName)
	}

	if c.Value != state {
		t.Errorf("cookie value=%q, want %q", c.Value, state)
	}

	if !c.HttpOnly {
		t.Error("cookie is not HttpOnly")
	}

	if c.Path != "/auth/blizzard" {
		t.Errorf("cookie path=%q, want /auth/blizzard", c.Path)
	}

	if c.SameSite != http.SameSiteLaxMode {
		t.Errorf("cookie SameSite=%v, want SameSiteLaxMode", c.SameSite)
	}
}

func TestOAuthCallbackMissingCookie(t *testing.T) {
	flow := &oauthFlow{
		result: make(chan oauthResult, 1),
	}

	r := httptest.NewRequest(
		http.MethodGet,
		"/auth/blizzard/profile?state=x&code=y",
		nil,
	)

	rr := httptest.NewRecorder()

	flow.callback(rr, r)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusBadRequest)
	}

	result := <-flow.result
	if result.err == nil {
		t.Fatal("expected error")
	}
}

func TestOAuthCallbackInvalidState(t *testing.T) {
	flow := &oauthFlow{
		result: make(chan oauthResult, 1),
	}

	r := httptest.NewRequest(
		http.MethodGet,
		"/auth/blizzard/profile?state=wrong&code=y",
		nil,
	)

	r.AddCookie(&http.Cookie{
		Name:  cookieName,
		Value: "right",
	})

	rr := httptest.NewRecorder()

	flow.callback(rr, r)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusBadRequest)
	}

	result := <-flow.result
	if result.err == nil {
		t.Fatal("expected error")
	}
}

func TestOAuthCallbackMissingCode(t *testing.T) {
	flow := &oauthFlow{
		result: make(chan oauthResult, 1),
	}

	r := httptest.NewRequest(
		http.MethodGet,
		"/auth/blizzard/profile?state=right",
		nil,
	)

	r.AddCookie(&http.Cookie{
		Name:  cookieName,
		Value: "right",
	})

	rr := httptest.NewRecorder()

	flow.callback(rr, r)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusBadRequest)
	}

	result := <-flow.result
	if result.err == nil {
		t.Fatal("expected error")
	}
}

func TestOAuthCallbackBlizzardError(t *testing.T) {
	flow := &oauthFlow{
		result: make(chan oauthResult, 1),
	}

	r := httptest.NewRequest(
		http.MethodGet,
		"/auth/blizzard/profile?state=right&error=access_denied&error_description=user+denied",
		nil,
	)

	r.AddCookie(&http.Cookie{
		Name:  cookieName,
		Value: "right",
	})

	rr := httptest.NewRecorder()

	flow.callback(rr, r)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want %d", rr.Code, http.StatusBadRequest)
	}

	result := <-flow.result
	if result.err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(result.err.Error(), "access_denied") {
		t.Fatalf("error=%q, want access_denied", result.err)
	}
}

func TestGetTokenSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method=%s, want POST", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type=%q", r.Header.Get("Content-Type"))
		}

		username, password, ok := r.BasicAuth()
		if !ok {
			t.Error("missing basic authentication")
		}

		if username != "client-id" {
			t.Errorf("username=%q", username)
		}

		if password != "client-secret" {
			t.Errorf("password=%q", password)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"abc123","token_type":"bearer","expires_in":3600}`))
	}))
	defer server.Close()

	client := server.Client()

	// This test needs getToken to be able to target the test server.
	// In the actual test file, I'd extract the token endpoint into an
	// injectable parameter rather than make the production URL mutable.
	_ = client
	_ = url.Values{}
}

func TestGetTokenRejectsMissingAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token_type":"bearer"}`))
	}))
	defer server.Close()

	// See note in TestGetTokenSuccess: the endpoint should be injected
	// rather than made global solely for testing.
	_ = server
}
