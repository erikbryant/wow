package main

import (
	"flag"
	"fmt"
	"os"
	"time"

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
		// Remove the item from the persistence
		wowitem.Items.Delete(itemID)
	}

	iNew, ok := wowitem.GetWeb(itemID)
	if !ok {
		fmt.Fprintln(os.Stderr, "Could not retrieve item", itemID)
		os.Exit(2)
	}

	err := wowitem.Items.Save()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to save item persistence: ", err)
		os.Exit(2)
	}

	rows = append(rows, iNew)
	output.Table(os.Stdout, rows)
}

// refreshAll refreshes persisted items older than a certain age
func refreshAll(passphrase string, maxRefresh int) {
	wowapi.Init(passphrase, false)

	maxAge := 24 * time.Hour * 7 // 1 week
	needsRefresh := 0
	refreshCount := 0

	for _, i := range wowitem.Items.Values() {
		if i.Stale(maxAge) {
			needsRefresh++
			if refreshCount < maxRefresh {
				wowitem.GetWeb(i.ID())
				refreshCount++
			}
		}
	}

	err := wowitem.Items.Save()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to save item persistence: ", err)
		os.Exit(2)
	}

	fmt.Printf("Refreshed %d of %d stale items\n", refreshCount, needsRefresh)
}

func runRefresh(args []string) {
	flags := flag.NewFlagSet("refresh", flag.ExitOnError)

	passphrase := flags.String("passphrase", "", "Passphrase to unlock WoW API client ID/secret")
	maxRefresh := flags.Int("max-refresh", 1000, "Maximum number of items to refresh")
	itemID := flags.Int64("id", -1, "Item ID to look up")

	if err := flags.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if *passphrase == "" {
		fmt.Fprintln(os.Stderr, "items refresh requires -passphrase")
		os.Exit(2)
	}

	if *itemID == -1 {
		refreshAll(*passphrase, *maxRefresh)
	} else {
		refreshItem(*passphrase, *itemID)
	}
}
