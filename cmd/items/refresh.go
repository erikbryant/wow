package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/erikbryant/wow/internal/appearanceset"
	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/wowitem"
)

// refreshItem refreshes a single item
func refreshItem(itemID int64, paths *path.Paths) error {
	wowItems, err := wowitem.New(paths.Items)
	if err != nil {
		return err
	}

	// Remove the item from the persistence. Even if we later fail to retrieve
	// new data, at least we got rid of stale data.
	wowItems.Delete(itemID)
	iNew, errGetLive := wowItems.GetLive(itemID)
	// We need to persist the deletion even if GetLive fails
	err = wowItems.Save()
	if err != nil {
		return fmt.Errorf("failed to save item persistence: %w", err)
	}

	if errGetLive != nil {
		return fmt.Errorf("could not retrieve itemID %d: %w", itemID, errGetLive)
	}

	as, err := appearanceset.New(paths.Appearances)
	if err != nil {
		return err
	}

	output.Table(os.Stdout, []wowitem.Item{iNew}, as)

	return nil
}

// refreshAll refreshes persisted items older than a certain age
func refreshAll(maxRefresh int, paths *path.Paths) error {
	maxAge := 24 * time.Hour * 7 // 1 week
	needsRefresh := 0
	refreshCount := 0

	wowItems, err := wowitem.New(paths.Items)
	if err != nil {
		return err
	}

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

func runRefresh(args []string, paths *path.Paths) error {
	flags := flag.NewFlagSet("refresh", flag.ExitOnError)

	maxRefresh := flags.Int("max-refresh", 1000, "Maximum number of items to refresh")
	itemID := flags.Int64("id", -1, "Item ID to look up")

	if err := flags.Parse(args); err != nil {
		return err
	}

	err := authenticate(paths)
	if err != nil {
		return err
	}

	if *itemID == -1 {
		err = refreshAll(*maxRefresh, paths)
		if err != nil {
			return err
		}
	} else {
		err = refreshItem(*itemID, paths)
		if err != nil {
			return err
		}
	}

	return nil
}
