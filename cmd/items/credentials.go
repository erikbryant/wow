package main

import (
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/credentials"
)

func runCredentialsAdd(args []string) {
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

func runCredentialsDelete(args []string) {
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

func runCredentialsGet(args []string) {
	if len(args) != 1 {
		usage()
		os.Exit(1)
	}

	name := args[0]

	s, err := credentials.Get(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to get %s: %s\n", name, err)
		os.Exit(1)
	}
	fmt.Printf("%s: %s\n", name, s)
}

func runCredentials(args []string) {
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}

	command := args[0]
	args = args[1:]

	switch command {
	case "add":
		runCredentialsAdd(args)

	case "delete":
		runCredentialsDelete(args)

	case "get":
		runCredentialsGet(args)

	case "help":
		usage()

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", command)
		usage()
		os.Exit(1)
	}
}
