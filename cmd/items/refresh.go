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
func refreshItem(itemID int64, wowItem *wowitem.WoWItem) {
	rows := []wowitem.Item{}

	iOld, ok := wowItem.Items.Get(itemID)
	if ok {
		rows = append(rows, iOld)
		// Remove the item from the persistence
		wowItem.Items.Delete(itemID)
	}

	iNew, ok := wowItem.GetWeb(itemID)
	if !ok {
		fmt.Fprintln(os.Stderr, "Could not retrieve item", itemID)
		os.Exit(2)
	}

	err := wowItem.Items.Save()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to save item persistence: ", err)
		os.Exit(2)
	}

	rows = append(rows, iNew)
	output.Table(os.Stdout, rows)
}

// refreshAll refreshes persisted items older than a certain age
func refreshAll(maxRefresh int, wowItem *wowitem.WoWItem) {
	maxAge := 24 * time.Hour * 7 // 1 week
	needsRefresh := 0
	refreshCount := 0

	for _, i := range wowItem.Items.Values() {
		if i.Stale(maxAge) {
			needsRefresh++
			if refreshCount < maxRefresh {
				wowItem.GetWeb(i.ID())
				refreshCount++
			}
		}
	}

	err := wowItem.Items.Save()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to save item persistence: ", err)
		os.Exit(2)
	}

	fmt.Printf("Refreshed %d of %d stale items\n", refreshCount, needsRefresh)
}

func runRefresh(args []string, wowItem *wowitem.WoWItem) {
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

	err := wowapi.Init(*passphrase)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if *itemID == -1 {
		refreshAll(*maxRefresh, wowItem)
	} else {
		refreshItem(*itemID, wowItem)
	}
}
