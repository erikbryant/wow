package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/erikbryant/wow/internal/credentials"
	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

// refreshItem refreshes a single item
func refreshItem(itemID int64) {
	wowItems := wowitem.New()

	rows := []wowitem.Item{}

	iOld, ok := wowItems.Items.Get(itemID)
	if ok {
		rows = append(rows, iOld)
		// Remove the item from the persistence
		wowItems.Items.Delete(itemID)
	}

	iNew, err := wowItems.GetWeb(itemID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not retrieve itemID %d: %s\n", itemID, err)
		os.Exit(2)
	}

	err = wowItems.Items.Save()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to save item persistence: ", err)
		os.Exit(2)
	}

	rows = append(rows, iNew)
	output.Table(os.Stdout, rows)
}

// refreshAll refreshes persisted items older than a certain age
func refreshAll(maxRefresh int) {
	wowItems := wowitem.New()
	maxAge := 24 * time.Hour * 7 // 1 week
	needsRefresh := 0
	refreshCount := 0

	for _, i := range wowItems.Items.Values() {
		if i.Stale(maxAge) {
			needsRefresh++
			if refreshCount < maxRefresh {
				wowItems.GetWeb(i.ID())
				refreshCount++
			}
		}
	}

	err := wowItems.Items.Save()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to save item persistence: ", err)
		os.Exit(2)
	}

	fmt.Printf("Refreshed %d of %d stale items\n", refreshCount, needsRefresh)
}

func runRefresh(args []string) {
	flags := flag.NewFlagSet("refresh", flag.ExitOnError)

	maxRefresh := flags.Int("max-refresh", 1000, "Maximum number of items to refresh")
	itemID := flags.Int64("id", -1, "Item ID to look up")

	if err := flags.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	clientID, err := credentials.ReadFromKeychain("clientID")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	clientSecret, err := credentials.ReadFromKeychain("clientSecret")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = wowapi.Authenticate(clientID, clientSecret)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if *itemID == -1 {
		refreshAll(*maxRefresh)
	} else {
		refreshItem(*itemID)
	}
}
