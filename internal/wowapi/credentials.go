package wowapi

import (
	"fmt"
	"net/url"

	"github.com/erikbryant/aes"
	"github.com/erikbryant/wow/internal/wowoauth"
)

var (
	clientIDCrypt      = "f7FhewxUd0lWQz/zPb27ZcwI/ZqkaMyd5YyuskFyEugQEeiKsfL7dvr11Kx1Y+Mi23qMciOAPe5ksCOy"
	clientSecretCrypt  = "CtJH62iU6V3ZeqiHyKItECHahdUYgAFyfHmQ4DRabhWIv6JeK5K4dT7aiybot6MS4JitmDzuWSz1UHHv"
	clientID           string
	clientSecret       string
	accessToken        string
	profileAccessToken string
)

// wowAccessToken retrieves an access token (to authenticate API calls)
func wowAccessToken() (string, error) {
	data := url.Values{
		"grant_type": {"client_credentials"},
	}

	return wowoauth.GetToken(data, clientID, clientSecret)
}

// wowProfileAccessToken returns a profile access token (to authenticate user profile API calls)
func wowProfileAccessToken() (string, error) {
	return wowoauth.GetPAT(clientID, clientSecret)
}

func Init(passphrase string) error {
	var err error

	clientID, err = aes.Decrypt(clientIDCrypt, passphrase)
	if err != nil {
		return fmt.Errorf("unable to decrypt clientID: %s", err)
	}

	clientSecret, err = aes.Decrypt(clientSecretCrypt, passphrase)
	if err != nil {
		return fmt.Errorf("unable to decrypt clientSecret: %s", err)
	}

	accessToken, err = wowAccessToken()
	if err != nil {
		return fmt.Errorf("unable to get access token: %s", err)
	}

	profileAccessToken, err = wowProfileAccessToken()
	if err != nil {
		return fmt.Errorf("unable to get profile access token: %s", err)
	}

	return nil
}
