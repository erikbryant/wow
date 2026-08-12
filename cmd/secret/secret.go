package main

import (
	"fmt"
	"os"
)

func usage() {
	fmt.Println(`Usage:
  secret <command>

Commands:
  add <secret>                     Add a secret
  delete <secret>                  Delete a secret
  get <secret> [<secret> ...]      Get one or more secrets
  help                             Display this help message`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "add":
		runAdd(args)

	case "delete":
		runDelete(args)

	case "get":
		runGet(args)

	case "help":
		usage()

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", command)
		usage()
		os.Exit(1)
	}
}
