package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/erikbryant/wow/internal/credentials"
	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

// refreshItem refreshes a single item
func refreshItem(itemID int64) error {
	paths, err := path.New("")
	if err != nil {
		return err
	}
	wowItems := wowitem.New(paths.Items)

	// Remove the item from the persistence. Even if we later fail to retrieve
	// new data, at least we got rid of stale data.
	wowItems.Delete(itemID)

	iNew, err := wowItems.GetLive(itemID)
	if err != nil {
		return fmt.Errorf("could not retrieve itemID %d: %w", itemID, err)
	}

	err = wowItems.Save()
	if err != nil {
		return fmt.Errorf("failed to save item persistence: %w", err)
	}

	output.Table(os.Stdout, []wowitem.Item{iNew})

	return nil
}

// refreshAll refreshes persisted items older than a certain age
func refreshAll(maxRefresh int) error {
	maxAge := 24 * time.Hour * 7 // 1 week
	needsRefresh := 0
	refreshCount := 0

	paths, err := path.New("")
	if err != nil {
		return err
	}
	wowItems := wowitem.New(paths.Items)

	for _, i := range wowItems.Values() {
		if i.Stale(maxAge) {
			needsRefresh++
			if refreshCount < maxRefresh {
				_, err := wowItems.GetLive(i.ID())
				if err == nil {
					refreshCount++
				} else {
					fmt.Fprintf(os.Stderr, "could not retrieve itemID %d: %s\n", i.ID(), err)
					continue
				}
			}
		}
	}

	err = wowItems.Save()
	if err != nil {
		return fmt.Errorf("failed to save item persistence: %w", err)
	}

	fmt.Printf("Refreshed %d of %d stale items\n", refreshCount, needsRefresh)

	return nil
}

func runRefresh(args []string) error {
	flags := flag.NewFlagSet("refresh", flag.ExitOnError)

	maxRefresh := flags.Int("max-refresh", 1000, "Maximum number of items to refresh")
	itemID := flags.Int64("id", -1, "Item ID to look up")

	if err := flags.Parse(args); err != nil {
		return err
	}

	paths, err := path.New("")
	if err != nil {
		return err
	}

	clientID, err := credentials.ReadFromKeychain(paths.Secret, "clientID")
	if err != nil {
		return err
	}

	clientSecret, err := credentials.ReadFromKeychain(paths.Secret, "clientSecret")
	if err != nil {
		return err
	}

	err = wowapi.Authenticate(clientID, clientSecret)
	if err != nil {
		return err
	}

	if *itemID == -1 {
		err = refreshAll(*maxRefresh)
		if err != nil {
			return err
		}
	} else {
		err = refreshItem(*itemID)
		if err != nil {
			return err
		}
	}

	return nil
}
