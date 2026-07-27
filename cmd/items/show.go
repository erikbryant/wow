package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/itemcache"
	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/wowitem"
)

func show(itemID int64, format string) {
	i, ok := itemcache.LookupItem(itemID, 0)
	if !ok {
		fmt.Fprintln(os.Stderr, "Failed to LookupItem: ", itemID)
		os.Exit(2)
	}

	switch format {
	case "json":
		output.JSON(os.Stdout, []wowitem.Item{i})
	case "summary":
		s := output.Summary(i)
		fmt.Println(s)
	default:
		fmt.Fprintln(os.Stderr, "Unknown format: ", format)
		os.Exit(2)
	}
}

func runShow(args []string) {
	flags := flag.NewFlagSet("show", flag.ContinueOnError)

	itemID := flags.Int64("id", -1, "Item ID to look up")
	format := flags.String("format", "summary", "Output format: json or summary")

	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}

	if *itemID == -1 {
		fmt.Fprintln(os.Stderr, "show requires -id")
		os.Exit(2)
	}

	show(*itemID, *format)
}
