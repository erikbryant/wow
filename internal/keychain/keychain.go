package keychain

import (
	"fmt"
	"os/exec"
	"strings"
)

const service = "github.com/erikbryant/WorldOfWarcraft"

// GetSigned uses a signed external app to read from the keychain without triggering the "unknown app" dialogs
func GetSigned(secret string, key string) (string, error) {
	out, err := exec.Command(secret, "get", key).Output()
	if err != nil {
		return "", fmt.Errorf("unable to get %s: %w", key, err)
	}

	parts := strings.SplitN(string(out), ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid secret helper output")
	}

	return strings.TrimSpace(parts[1]), nil
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
