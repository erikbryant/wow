package main

import (
	"fmt"

	"github.com/erikbryant/wow/internal/credentials"
)

func runAdd(args []string) error {
	if len(args) != 2 {
		usage()
		return fmt.Errorf("add requires 2 arguments")
	}

	name := args[0]
	value := args[1]

	err := credentials.Add(name, value)
	if err != nil {
		return fmt.Errorf("unable to add %s: %w", name, err)
	}

	return nil
}
