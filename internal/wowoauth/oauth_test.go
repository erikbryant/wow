package wowoauth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateStateOAuthCookie(t *testing.T) {
	rr := httptest.NewRecorder()
	state := generateStateOAuthCookie(rr)
	if state == "" {
		t.Fatal("empty state")
	}
	cookies := rr.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies=%d", len(cookies))
	}
	c := cookies[0]
	if c.Name != cookieName || c.Value != state || !c.HttpOnly || c.Path != "/auth/blizzard" || c.SameSite != http.SameSiteLaxMode {
		t.Fatalf("cookie=%+v", c)
	}
}

func TestOAuthCallbackMissingCookie(t *testing.T) {
	r := httptest.NewRequest("GET", "/auth/blizzard/profile?state=x&code=y", nil)
	rr := httptest.NewRecorder()
	oAuthBlizzardCallback(rr, r)
	if rr.Code != 307 {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestOAuthCallbackInvalidState(t *testing.T) {
	r := httptest.NewRequest("GET", "/auth/blizzard/profile?state=wrong&code=y", nil)
	r.AddCookie(&http.Cookie{Name: cookieName, Value: "right"})
	rr := httptest.NewRecorder()
	oAuthBlizzardCallback(rr, r)
	if rr.Code != 307 {
		t.Fatalf("status=%d", rr.Code)
	}
}
