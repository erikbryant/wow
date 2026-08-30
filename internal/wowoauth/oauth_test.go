package wowoauth

import (
	"net/http"
	"net/http/httptest"
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
