package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/wowitem"
)

func json(itemID int64, wowItem *wowitem.WoWItem) {
	i, err := wowItem.Get(itemID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get itemID %d: %s\n", itemID, err)
		os.Exit(2)
	}

	output.JSON(os.Stdout, []wowitem.Item{i})
}

func runJSON(args []string, wowItem *wowitem.WoWItem) {
	flags := flag.NewFlagSet("json", flag.ExitOnError)

	itemID := flags.Int64("id", -1, "Item ID to look up")

	if err := flags.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if *itemID == -1 {
		fmt.Fprintln(os.Stderr, "json requires -id")
		os.Exit(2)
	}

	json(*itemID, wowItem)
}
