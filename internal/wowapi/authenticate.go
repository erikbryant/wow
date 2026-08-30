package wowapi

import (
	"net/http"
	"net/url"

	"github.com/erikbryant/wow/internal/keychain"
	"github.com/erikbryant/wow/internal/wowoauth"
)

// authenticate authenticates the package-level WoW API client.
//
// Applications should normally call this once during startup. The resulting
// authenticated client is then used by the package-level API functions.
func (c *Client) authenticate() error {
	var err error

	data := url.Values{
		"grant_type": {"client_credentials"},
	}

	c.accessToken, err = wowoauth.GetToken(data, c.clientID, c.clientSecret)
	if err != nil {
		return err
	}

	c.profileAccessToken, err = wowoauth.GetPAT(c.clientID, c.clientSecret)
	if err != nil {
		return err
	}

	return nil
}

// getSecretsFromKeychain authenticates the package-level WoW API client
// using credentials stored in the keychain.
func (c *Client) getSecretsFromKeychain(secretPath string) error {
	var err error

	c.clientID, err = keychain.GetSigned(secretPath, "clientID")
	if err != nil {
		return err
	}

	c.clientSecret, err = keychain.GetSigned(secretPath, "clientSecret")
	if err != nil {
		return err
	}

	return nil
}

// isInitialized returns true if the default client has already been initialized
func isInitialized() bool {
	initialized := false

	defaultClientMu.RLock()
	initialized = defaultClient != nil
	defaultClientMu.RUnlock()

	return initialized
}

func Init(secretPath string) error {
	defaultClientMu.RLock()
	defer defaultClientMu.RUnlock()
	if isInitialized() {
		return nil
	}

	var err error

	c := Client{
		apiBase:    defaultAPIBase,
		httpClient: http.DefaultClient,
	}

	err = c.getSecretsFromKeychain(secretPath)
	if err != nil {
		return err
	}

	err = c.authenticate()
	if err != nil {
		return err
	}

	defaultClientMu.Lock()
	defaultClient = &c
	defaultClientMu.Unlock()

	return err
}
