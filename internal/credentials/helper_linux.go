package credentials

import "fmt"

// kcAdd adds an Item to a Keychain
func kcAdd(service, name, value string) error {
	return fmt.Errorf("kcAdd not implemented")
}

// kcQuery returns a single query result
func kcQuery(service, name string) (queryResult, error) {
	return queryResult{}, fmt.Errorf("kcQuery not implemented")
}

// kcDelete removes an entry from the keychain
func kcDelete(service, name string) error {
	return fmt.Errorf("kcDelete not implemented")
}
