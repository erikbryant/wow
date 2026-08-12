package main

import (
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/credentials"
)

func runDelete(args []string) {
	if len(args) != 1 {
		usage()
		os.Exit(1)
	}

	name := args[0]

	err := credentials.Delete(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to delete %s: %s\n", name, err)
		os.Exit(1)
	}
}
