package main

import (
	"fmt"
	"os"
)

func usage() {
	fmt.Println(`Usage:
  items <command>

Commands:
  delete -id <id>                                    Delete persisted item
  json -id <id>                                      Show JSON for an item
  query [options]                                    Search for items
  refresh -passphrase <pass> [-max-refresh=1000]     Refresh stale items
  help                                               Display this help message

Examples:
  items delete -id 12345
  items json -id 12345
  items query
  items query -rare -in-appearance-set
  items refresh -max-refresh=42
  items refresh -id 12345`)
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
