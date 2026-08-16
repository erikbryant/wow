package main

import (
	"fmt"

	"github.com/erikbryant/wow/internal/keychain"
)

func runGet(args []string) error {
	if len(args) < 1 {
		usage()
		return fmt.Errorf("get requires at least one argument")
	}

	for _, name := range args {
		s, err := keychain.Get(name)
		if err != nil {
			return fmt.Errorf("unable to get %s: %w", name, err)
		}
		fmt.Printf("%s: %s\n", name, s)
	}

	return nil
}
