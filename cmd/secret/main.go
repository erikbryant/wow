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

	var err error

	switch command {
	case "add":
		err = runAdd(args)
	case "delete":
		err = runDelete(args)
	case "get":
		err = runGet(args)
	case "help":
		usage()
	default:
		usage()
		err = fmt.Errorf("unknown command: %s", command)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
