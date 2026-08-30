package wowoauth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGetTokenSuccess(t *testing.T) {
	var gotMethod string
	var gotContentType string
	var gotUsername string
	var gotPassword string
	var gotForm url.Values

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")

		var ok bool
		gotUsername, gotPassword, ok = r.BasicAuth()
		if !ok {
			t.Error("missing basic authentication")
		}

		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
			return
		}
		gotForm = r.PostForm

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(
			`{"access_token":"abc123","token_type":"bearer","expires_in":3600}`,
		))
	}))
	defer server.Close()

	client := server.Client()

	data := url.Values{
		"grant_type": {"authorization_code"},
		"code":       {"authorization-code"},
		"redirect_uri": {
			"http://localhost:8888/auth/blizzard/profile",
		},
	}

	token, err := getToken(
		context.Background(),
		client,
		server.URL,
		data,
		"client-id",
		"client-secret",
	)
	if err != nil {
		t.Fatalf("getToken: %v", err)
	}

	if token != "abc123" {
		t.Fatalf("token=%q, want %q", token, "abc123")
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method=%q, want %q", gotMethod, http.MethodPost)
	}

	if gotContentType != "application/x-www-form-urlencoded" {
		t.Errorf(
			"Content-Type=%q, want %q",
			gotContentType,
			"application/x-www-form-urlencoded",
		)
	}

	if gotUsername != "client-id" {
		t.Errorf("username=%q, want %q", gotUsername, "client-id")
	}

	if gotPassword != "client-secret" {
		t.Errorf("password=%q, want %q", gotPassword, "client-secret")
	}

	if got := gotForm.Get("grant_type"); got != "authorization_code" {
		t.Errorf("grant_type=%q, want %q", got, "authorization_code")
	}

	if got := gotForm.Get("code"); got != "authorization-code" {
		t.Errorf("code=%q, want %q", got, "authorization-code")
	}

	if got := gotForm.Get("redirect_uri"); got != redirectURL {
		t.Errorf("redirect_uri=%q, want %q", got, redirectURL)
	}
}

func TestGetTokenRejectsMissingAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token_type":"bearer","expires_in":3600}`))
	}))
	defer server.Close()

	token, err := getToken(
		context.Background(),
		server.Client(),
		server.URL,
		url.Values{"grant_type": {"authorization_code"}},
		"client-id",
		"client-secret",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if token != "" {
		t.Errorf("token=%q, want empty", token)
	}

	if !strings.Contains(err.Error(), "did not contain access_token") {
		t.Errorf("error=%q, want missing access_token error", err)
	}
}

func TestGetTokenOAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(
			w,
			`{"error":"invalid_grant","error_description":"authorization code expired"}`,
			http.StatusBadRequest,
		)
	}))
	defer server.Close()

	token, err := getToken(
		context.Background(),
		server.Client(),
		server.URL,
		url.Values{"grant_type": {"authorization_code"}},
		"client-id",
		"client-secret",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if token != "" {
		t.Errorf("token=%q, want empty", token)
	}

	want := "token request failed: invalid_grant: authorization code expired"
	if err.Error() != want {
		t.Errorf("error=%q, want %q", err, want)
	}
}

func TestGetTokenHTTPErrorWithoutOAuthError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
	}))
	defer server.Close()

	token, err := getToken(
		context.Background(),
		server.Client(),
		server.URL,
		url.Values{"grant_type": {"authorization_code"}},
		"client-id",
		"client-secret",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if token != "" {
		t.Errorf("token=%q, want empty", token)
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error=%q, want HTTP 500", err)
	}
}

func TestGetTokenMalformedSuccessResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":`))
	}))
	defer server.Close()

	token, err := getToken(
		context.Background(),
		server.Client(),
		server.URL,
		url.Values{"grant_type": {"authorization_code"}},
		"client-id",
		"client-secret",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if token != "" {
		t.Errorf("token=%q, want empty", token)
	}

	if !strings.Contains(err.Error(), "decode token response") {
		t.Errorf("error=%q, want decode error", err)
	}
}

func TestGetTokenOAuthErrorWithoutDescription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_client"}`))
	}))
	defer server.Close()

	token, err := getToken(
		context.Background(),
		server.Client(),
		server.URL,
		url.Values{"grant_type": {"authorization_code"}},
		"client-id",
		"client-secret",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if token != "" {
		t.Errorf("token=%q, want empty", token)
	}

	want := "token request failed: invalid_client"
	if err.Error() != want {
		t.Errorf("error=%q, want %q", err, want)
	}
}

func TestGetTokenContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	token, err := getToken(
		ctx,
		server.Client(),
		server.URL,
		url.Values{"grant_type": {"authorization_code"}},
		"client-id",
		"client-secret",
	)
	if err == nil {
		t.Fatal("expected error")
	}

	if token != "" {
		t.Errorf("token=%q, want empty", token)
	}

	if !strings.Contains(err.Error(), "send token request") {
		t.Errorf("error=%q, want send token request error", err)
	}
}
