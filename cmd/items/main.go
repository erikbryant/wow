package main

import (
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/path"
)

func usage() {
	fmt.Println(`Usage:
  items <command>

Commands:
  create {appearance|item}        Create a new persistence
  delete -id <id>                 Delete persisted item
  json -id <id>                   Show JSON for an item
  query [options]                 Search for items
  refresh [-max-refresh=1000]     Refresh stale items
  help                            Display this help message

Examples:
  items create appearance
  items create item
  items delete -id 12345
  items json -id 12345
  items query
  items query -rare -in-appearance-set
  items refresh -max-refresh=42
  items refresh -id 12345
  `)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error

	// Use consistent paths across all subcommands
	paths, err := path.New("")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch command {
	case "create":
		err = runCreate(args, paths)
	case "delete":
		err = runDelete(args, paths)
	case "json":
		err = runJSON(args, paths)
	case "query":
		err = runQuery(args, paths)
	case "refresh":
		err = runRefresh(args, paths)
	case "help":
		usage()
	default:
		usage()
		err = fmt.Errorf("unknown command: %s", command)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
