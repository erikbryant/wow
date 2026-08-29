package wowoauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const tokenHTTPTimeout = 30 * time.Second

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type tokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func GetToken(data url.Values, clientID, clientSecret string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), tokenHTTPTimeout)
	defer cancel()

	return getToken(ctx, http.DefaultClient, tokenURL, data, clientID, clientSecret)
}

func getToken(
	ctx context.Context,
	client *http.Client,
	endpoint string,
	data url.Values,
	clientID string,
	clientSecret string,
) (string, error) {
	if client == nil {
		client = http.DefaultClient
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(clientID, clientSecret)

	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("send token request: %w", err)
	}
	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var tokenError tokenErrorResponse

		if err := json.Unmarshal(contents, &tokenError); err == nil &&
			tokenError.Error != "" {
			if tokenError.ErrorDescription != "" {
				return "", fmt.Errorf(
					"token request failed: %s: %s",
					tokenError.Error,
					tokenError.ErrorDescription,
				)
			}

			return "", fmt.Errorf("token request failed: %s", tokenError.Error)
		}

		return "", fmt.Errorf(
			"token request failed with HTTP status %s",
			response.Status,
		)
	}

	var token tokenResponse
	if err := json.Unmarshal(contents, &token); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	if token.AccessToken == "" {
		return "", fmt.Errorf("token response did not contain access_token")
	}

	return token.AccessToken, nil
}

func tokenToPAT(code, clientID, clientSecret string) (string, error) {
	data := url.Values{
		"redirect_uri": {redirectURL},
		"grant_type":   {"authorization_code"},
		"code":         {code},
	}

	return GetToken(data, clientID, clientSecret)
}
