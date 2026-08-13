package wowapi

import (
	"fmt"
	"net/url"

	"github.com/erikbryant/wow/internal/wowoauth"
)

var (
	accessToken        string
	profileAccessToken string
)

// wowAccessToken retrieves an access token (to authenticate API calls)
func wowAccessToken(clientID, clientSecret string) (string, error) {
	data := url.Values{
		"grant_type": {"client_credentials"},
	}

	return wowoauth.GetToken(data, clientID, clientSecret)
}

// wowProfileAccessToken returns a profile access token (to authenticate user profile API calls)
func wowProfileAccessToken(clientID, clientSecret string) (string, error) {
	return wowoauth.GetPAT(clientID, clientSecret)
}

func Authenticate(clientID, clientSecret string) error {
	var err error

	accessToken, err = wowAccessToken(clientID, clientSecret)
	if err != nil {
		return fmt.Errorf("unable to get access token: %w", err)
	}

	profileAccessToken, err = wowProfileAccessToken(clientID, clientSecret)
	if err != nil {
		return fmt.Errorf("unable to get profile access token: %w", err)
	}

	return nil
}
