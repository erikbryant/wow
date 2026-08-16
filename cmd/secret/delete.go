package main

import (
	"fmt"

	"github.com/erikbryant/wow/internal/keychain"
)

func runDelete(args []string) error {
	if len(args) != 1 {
		usage()
		return fmt.Errorf("delete requires exactly one argument")
	}

	name := args[0]

	err := keychain.Delete(name)
	if err != nil {
		return fmt.Errorf("unable to delete %s: %w", name, err)
	}

	return nil
}
