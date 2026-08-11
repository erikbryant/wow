package credentials

import (
	"errors"
	"fmt"

	"github.com/keybase/go-keychain"
)

type KeychainStore struct{}

const service = "github.com/erikbryant/WorldOfWarcraft"

var (
	ErrNotFound = errors.New("credential not found")
	ErrExists   = errors.New("credential already exists")
)

func New() KeychainStore {
	return KeychainStore{}
}

func (KeychainStore) Get(name string) (string, error) {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService(service)
	item.SetAccount(name)
	item.SetMatchLimit(keychain.MatchLimitOne)
	item.SetReturnData(true)

	results, err := keychain.QueryItem(item)
	if err != nil {
		return "", fmt.Errorf("query keychain: %w", err)
	}

	if len(results) == 0 {
		return "", ErrNotFound
	}

	if len(results) != 1 {
		return "", fmt.Errorf("expected one keychain item, got %d", len(results))
	}

	return string(results[0].Data), nil
}

func (KeychainStore) Set(name, value string) error {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService(service)
	item.SetAccount(name)
	item.SetData([]byte(value))
	item.SetSynchronizable(keychain.SynchronizableNo)

	err := keychain.AddItem(item)
	if err == keychain.ErrorDuplicateItem {
		return ErrExists
	}
	if err != nil {
		return fmt.Errorf("add keychain item: %w", err)
	}

	return nil
}

func (KeychainStore) Delete(name string) error {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService(service)
	item.SetAccount(name)

	err := keychain.DeleteItem(item)
	if err == keychain.ErrorItemNotFound {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("delete keychain item: %w", err)
	}

	return nil
}
