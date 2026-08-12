package main

import (
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/credentials"
)

func runGet(args []string) {
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}

	for _, name := range args {
		s, err := credentials.Get(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to get %s: %s\n", name, err)
			os.Exit(1)
		}
		fmt.Printf("%s: %s\n", name, s)
	}
}
