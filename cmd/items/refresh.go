package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/erikbryant/wow/internal/itemcache"
	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

// refreshItem refreshes a single item
func refreshItem(passphrase string, itemID int64) {
	wowapi.Init(passphrase, false)

	iOld, ok := itemcache.LookupItem(itemID, 0)
	if !ok {
		fmt.Fprintln(os.Stderr, "Failed to lookup item: ", itemID)
		os.Exit(2)
	}

	// Get the latest values from the web API
	itemcache.DisableRead()
	defer itemcache.Save()

	iNew, ok := itemcache.LookupItem(itemID, 0)
	if !ok {
		fmt.Fprintln(os.Stderr, "Failed to lookup item: ", itemID)
		os.Exit(2)
	}

	output.Table(os.Stdout, []wowitem.Item{iOld, iNew})
}

// refreshCache refreshes cached items older than a certain age
func refreshCache(passphrase string, maxRefresh int) {
	wowapi.Init(passphrase, false)

	maxAge := 24 * time.Hour * 7 // 1 week
	needsRefresh := 0
	refreshCount := 0

	for _, i := range itemcache.ItemsSlice() {
		if i.Stale(maxAge) {
			needsRefresh++
			if refreshCount < maxRefresh {
				itemcache.LookupItem(i.ID(), maxAge)
				refreshCount++
			}
		}
	}

	err := itemcache.Save()
	if err != nil {
		log.Fatalln("Failed to save cache: ", err)
	}

	fmt.Printf("Refreshed %d of %d stale items\n", refreshCount, needsRefresh)
}

func runRefresh(args []string) {
	flags := flag.NewFlagSet("refresh", flag.ContinueOnError)

	passphrase := flags.String("passphrase", "", "Passphrase to unlock WoW API client Id/secret")
	maxRefresh := flags.Int("max-refresh", 1000, "Maximum number of items to refresh")
	itemID := flags.Int64("id", -1, "Item ID to look up")

	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}

	if *passphrase == "" {
		fmt.Fprintln(os.Stderr, "items refresh requires -passphrase")
		os.Exit(2)
	}

	if *itemID == -1 {
		refreshCache(*passphrase, *maxRefresh)
	} else {
		refreshItem(*passphrase, *itemID)
	}
}
