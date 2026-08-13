package credentials

import (
	"fmt"
	"os/exec"
	"strings"
)

const service = "github.com/erikbryant/WorldOfWarcraft"

// ReadFromKeychain reads from the keychain without triggering the "unknown app" dialogs
func ReadFromKeychain(key string) (string, error) {
	out, err := exec.Command("./bin/secret", "get", key).Output()
	if err != nil {
		return "", fmt.Errorf("unable to get clientID: %s", err)
	}

	secret := string(out)
	secret = strings.TrimSpace(secret)
	parts := strings.Split(secret, " ")

	return parts[1], nil
}

// Get returns a value from the keychain
func Get(name string) (string, error) {
	result, err := kcQuery(service, name)
	if err != nil {
		return "", fmt.Errorf("query keychain item: %w", err)
	}
	return string(result.Data), nil
}

// Add stores a new value in the keychain
func Add(name, value string) error {
	err := kcAdd(service, name, value)
	if err != nil {
		return fmt.Errorf("add keychain item: %w", err)
	}

	return nil
}

// Delete deletes a value from the keychain
func Delete(name string) error {
	err := kcDelete(service, name)
	if err != nil {
		return fmt.Errorf("delete keychain item: %w", err)
	}

	return nil
}
