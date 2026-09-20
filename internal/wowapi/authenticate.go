package wowapi

import (
	"net/url"

	"github.com/erikbryant/wow/internal/keychain"
	"github.com/erikbryant/wow/internal/wowoauth"
)

// authenticate authenticates the package-level WoW API client.
//
// Applications should normally call this once during startup. The resulting
// authenticated client is then used by the package-level API functions.
func authenticate(clientID, clientSecret string) (string, string, error) {
	var err error

	data := url.Values{
		"grant_type": {"client_credentials"},
	}

	accessToken, err := wowoauth.GetToken(data, clientID, clientSecret)
	if err != nil {
		return "", "", err
	}

	profileAccessToken, err := wowoauth.GetPAT(clientID, clientSecret)
	if err != nil {
		return "", "", err
	}

	return accessToken, profileAccessToken, nil
}

// getSecretsFromKeychain authenticates the package-level WoW API client
// using credentials stored in the keychain.
func getSecretsFromKeychain(secretPath string) (string, string, error) {
	var err error

	clientID, err := keychain.GetSigned(secretPath, "clientID")
	if err != nil {
		return "", "", err
	}

	clientSecret, err := keychain.GetSigned(secretPath, "clientSecret")
	if err != nil {
		return "", "", err
	}

	return clientID, clientSecret, nil
}
