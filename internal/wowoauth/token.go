package wowoauth

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func GetToken(data url.Values, clientID, clientSecret string) (string, error) {
	request, err := http.NewRequest("POST", "https://oauth.battle.net/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var jsonObject map[string]any

	err = json.Unmarshal(contents, &jsonObject)
	if err != nil {
		return "", err
	}

	return jsonObject["access_token"].(string), nil
}
