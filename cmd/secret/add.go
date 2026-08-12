package main

import (
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/credentials"
)

func runAdd(args []string) {
	if len(args) != 2 {
		usage()
		os.Exit(1)
	}

	name := args[0]
	value := args[1]

	err := credentials.Add(name, value)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to add %s: %s\n", name, err)
		os.Exit(1)
	}
}
