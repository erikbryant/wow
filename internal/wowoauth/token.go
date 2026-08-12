package wowoauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
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

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	return jsonObject["access_token"].(string), nil
}

// GetPAT returns a profile access token (to authenticate user profile API calls)
func GetPAT(clientID, clientSecret string) (string, error) {
	go start(clientID, clientSecret)
	defer shutdown()

	cmd := exec.Command("open", "http://localhost:8888/auth/blizzard/login")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("unable to open browser: %s", err)
	}

	for paToken == "" {
		// Wait for OAuth to complete
		time.Sleep(time.Second)
	}

	return paToken, nil
}
