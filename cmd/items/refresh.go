package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

// refreshItem refreshes a single item
func refreshItem(passphrase string, itemID int64) {
	wowapi.Init(passphrase, false)
	rows := []wowitem.Item{}

	iOld, ok := wowitem.Items.Get(itemID)
	if ok {
		rows = append(rows, iOld)
		// Remove the item from the cache
		wowitem.Items.Delete(itemID)
	}

	// Get the latest value from the web API
	result, ok := wowapi.Item(web.ToString(itemID))
	if !ok {
		log.Fatal("Item not found, fix this error message/handling")
	}

	// Write it to the persistence
	iNew := wowitem.NewItem(result)
	wowitem.Items.Set(iNew.ID(), iNew)

	err := wowitem.Items.Save()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to save item cache: ", err)
		os.Exit(2)
	}

	rows = append(rows, iNew)
	output.Table(os.Stdout, rows)
}

// refreshCache refreshes cached items older than a certain age
func refreshCache(passphrase string, maxRefresh int) {
	wowapi.Init(passphrase, false)

	maxAge := 24 * time.Hour * 7 // 1 week
	needsRefresh := 0
	refreshCount := 0

	for _, i := range wowitem.Items.Values() {
		if i.Stale(maxAge) {
			needsRefresh++
			if refreshCount < maxRefresh {
				wowitem.LookupItem(i.ID(), maxAge)
				refreshCount++
			}
		}
	}

	err := wowitem.Items.Save()
	if err != nil {
		log.Fatalln("Failed to save cache: ", err)
	}

	fmt.Printf("Refreshed %d of %d stale items\n", refreshCount, needsRefresh)
}

func runRefresh(args []string) {
	flags := flag.NewFlagSet("refresh", flag.ContinueOnError)

	passphrase := flags.String("passphrase", "", "Passphrase to unlock WoW API client ID/secret")
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
