package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/erikbryant/wow/internal/itemcache"
	"github.com/erikbryant/wow/internal/wowapi"
)

// refreshItem refreshes a single item
func refreshItem(passphrase string, itemID int64) {
	wowapi.Init(passphrase, false)

	show(itemID, "summary")

	// Get the latest values
	itemcache.DisableRead()
	defer itemcache.Save()

	show(itemID, "summary")
}

// refreshCache refreshes cached items older than a certain age
func refreshCache(passphrase string, maxRefresh int) {
	wowapi.Init(passphrase, false)

	maxAge := 24 * time.Hour * 7 // 1 week
	needsRefresh := 0
	refreshCount := 0

	for _, i := range itemcache.ItemsCopy() {
		if i.Stale(maxAge) {
			needsRefresh++
			if refreshCount < maxRefresh {
				itemcache.LookupItem(i.ID(), maxAge)
				refreshCount++
			}
		}
	}

	itemcache.Save()

	fmt.Printf("Refreshed %d of %d stale items\n", refreshCount, needsRefresh)
}

func runRefresh(args []string) {
	flags := flag.NewFlagSet("refresh", flag.ContinueOnError)

	passphrase := flags.String("passphrase", "", "Passphrase to unlock WOW API client Id/secret")
	maxRefresh := flags.Int("max-refresh", 1000, "Maximum number of items to refresh")
	itemID := flags.Int64("id", -1, "Item ID to look up")

	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}

	if *passphrase == "" {
		fmt.Fprintln(os.Stderr, "refresh requires --passphrase")
		os.Exit(2)
	}

	if *itemID == -1 {
		refreshCache(*passphrase, *maxRefresh)
	} else {
		refreshItem(*passphrase, *itemID)
	}
}
