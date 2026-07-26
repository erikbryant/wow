package main

import (
	"fmt"
	"os"
)

func usage() {
	fmt.Println(`Usage:
  items <command>

Commands:
  delete -id <id>                                    Delete cached item
  list                                               List all items
  query [options]                                    Search for items
  refresh -passphrase <pass> [-max-refresh=1000]     Refresh stale items
  show -id <id> [-format={summary|json}]             Show a single item's details

Examples:
  items delete -id 12345
  items list
  items query -rare -in-appearance-set
  items refresh -passphrase foobar -max-refresh=42
  items refresh -id 12345 -passphrase foobar
  items show -id 19019 -format=json
  items show -id 12345`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "delete":
		runDelete(args)

	case "list":
		runList(args)

	case "query":
		runQuery(args)

	case "refresh":
		runRefresh(args)

	case "show":
		runShow(args)

	case "help":
		usage()

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", command)
		usage()
		os.Exit(1)
	}
}
