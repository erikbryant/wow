package main

import (
	"fmt"
	"os"

	"golang.org/x/term"

	"github.com/erikbryant/wow/internal/keychain"
)

func readSecret() (string, error) {
	fmt.Print("Secret: ")

	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // ReadPassword doesn't echo the newline.
	if err != nil {
		return "", fmt.Errorf("reading secret: %w", err)
	}

	return string(b), nil
}

func runAdd(args []string) error {
	if len(args) != 1 {
		usage()
		return fmt.Errorf("add requires exactly 1 argument")
	}

	name := args[0]
	value, err := readSecret()
	if err != nil {
		return err
	}

	err = keychain.Add(name, value)
	if err != nil {
		return fmt.Errorf("unable to add %s: %w", name, err)
	}

	return nil
}
