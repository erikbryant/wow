package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/erikbryant/wow/internal/credentials"
)

func runCredentialsDelete(args []string) {
	if len(args) != 1 {
		usage()
		os.Exit(1)
	}

	name := args[0]
	c := credentials.New()
	err := c.Delete(name)
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
	c := credentials.New()
	s, err := c.Get(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to get %s: %s\n", name, err)
		os.Exit(1)
	}
	fmt.Printf("%s: %s\n", name, s)
}

func runCredentialsSet(args []string) {
	fmt.Println("credentials set", strings.Join(args, " "))

	if len(args) != 2 {
		usage()
		os.Exit(1)
	}

	name := args[0]
	value := args[1]
	c := credentials.New()
	err := c.Set(name, value)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to set %s: %s\n", name, err)
		os.Exit(1)
	}
}

func runCredentials(args []string) {
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}

	command := args[0]
	args = args[1:]

	switch command {
	case "delete":
		runCredentialsDelete(args)

	case "get":
		runCredentialsGet(args)

	case "set":
		runCredentialsSet(args)

	case "help":
		usage()

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", command)
		usage()
		os.Exit(1)
	}
}
