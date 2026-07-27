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
  json -id <id>                                      Show JSON for an item
  list                                               List all items
  query [options]                                    Search for items
  refresh -passphrase <pass> [-max-refresh=1000]     Refresh stale items

Examples:
  items delete -id 12345
  items json -id 12345
  items list
  items query -rare -in-appearance-set
  items refresh -passphrase foobar -max-refresh=42
  items refresh -passphrase foobar -id 12345`)
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

	case "json":
		runJSON(args)

	case "list":
		runList(args)

	case "query":
		runQuery(args)

	case "refresh":
		runRefresh(args)

	case "help":
		usage()

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", command)
		usage()
		os.Exit(1)
	}
}
