package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/wowitem"
)

func json(itemID int64) {
	i, ok := wowitem.Get(itemID)
	if !ok {
		fmt.Fprintln(os.Stderr, "Failed to LookupItem: ", itemID)
		os.Exit(2)
	}

	output.JSON(os.Stdout, []wowitem.Item{i})
}

func runJSON(args []string) {
	flags := flag.NewFlagSet("json", flag.ContinueOnError)

	itemID := flags.Int64("id", -1, "Item ID to look up")

	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}

	if *itemID == -1 {
		fmt.Fprintln(os.Stderr, "json requires -id")
		os.Exit(2)
	}

	json(*itemID)
}
