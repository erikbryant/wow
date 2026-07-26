package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/itemcache"
)

func show(itemID int64, format string) {
	i, ok := itemcache.LookupItem(itemID, 0)
	if !ok {
		fmt.Fprintln(os.Stderr, "Failed to LookupItem: ", itemID)
		os.Exit(2)
	}

	switch format {
	case "json":
		b, _ := json.MarshalIndent(i.XItem, "\t", "\t")
		fmt.Println(string(b))
	case "summary":
		fmt.Println(i.Format())
	default:
		fmt.Fprintln(os.Stderr, "Unknown format: ", format)
		os.Exit(2)
	}
}

func runShow(args []string) {
	flags := flag.NewFlagSet("refresh", flag.ContinueOnError)

	itemID := flags.Int64("id", -1, "Item ID to look up")
	format := flags.String("format", "summary", "Output format: json or summary")

	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}

	if *itemID == -1 {
		fmt.Fprintln(os.Stderr, "show requires --id")
		os.Exit(2)
	}

	show(*itemID, *format)
}
