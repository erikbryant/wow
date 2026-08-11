package credentials

import (
	"fmt"
)

const service = "github.com/erikbryant/WorldOfWarcraft"

func Get(name string) (string, error) {
	result, err := kcQuery(service, name)
	if err != nil {
		return "", fmt.Errorf("query keychain item: %w", err)
	}
	return string(result.Data), nil
}

func Add(name, value string) error {
	err := kcAdd(service, name, value)
	if err != nil {
		return fmt.Errorf("add keychain item: %w", err)
	}

	return nil
}

func Delete(name string) error {
	err := kcDelete(service, name)
	if err != nil {
		return fmt.Errorf("delete keychain item: %w", err)
	}

	return nil
}
