package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/erikbryant/wow/internal/appearanceset"
	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

// deleteAndGetNew removes the item from the persistence. Even if we later fail to retrieve
// new data, at least we got rid of stale data.
func deleteAndGetNew(wowItems *wowitem.Persistence, itemID int64, client *wowapi.Client) (wowitem.Item, error) {
	wowItems.Delete(itemID)
	return wowItems.GetLive(itemID, client)
}

// invalidateAndRefresh refreshes a single item
func invalidateAndRefresh(itemID int64, paths *path.Paths, client *wowapi.Client) error {
	wowItems, err := wowitem.New(paths.Items)
	if err != nil {
		return err
	}

	iNew, errGetNew := deleteAndGetNew(wowItems, itemID, client)

	// Persist the deletion even if getting the new data fails
	err = wowItems.Save()
	if err != nil {
		return fmt.Errorf("failed to save item persistence: %w", err)
	}

	if errGetNew != nil {
		return fmt.Errorf("could not retrieve itemID %d: %w", itemID, errGetNew)
	}

	as, err := appearanceset.New(paths.Appearances)
	if err != nil {
		return err
	}

	output.Table(os.Stdout, []wowitem.Item{iNew}, as)

	return nil
}

// invalidateAndRefreshStale deletes items older than a certain age and attempts to refresh them
func invalidateAndRefreshStale(maxRefresh int, paths *path.Paths, client *wowapi.Client) error {
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
				_, err := deleteAndGetNew(wowItems, i.ID(), client)
				if err == nil {
					refreshCount++
				} else {
					if i.Source() == "synthetic" {
						fmt.Fprintf(os.Stderr, "deleted synthetic itemID %d; run `wowctl populate` to recreate\n", i.ID())
					} else {
						fmt.Fprintf(os.Stderr, "could not retrieve itemID %d: %s\n", i.ID(), err)
					}
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

	err := wowapi.Init(paths.Secret)
	if err != nil {
		return err
	}

	client, err := wowapi.NewClient()
	if err != nil {
		return err
	}

	if *itemID == -1 {
		err = invalidateAndRefreshStale(*maxRefresh, paths, client)
		if err != nil {
			return err
		}
	} else {
		err = invalidateAndRefresh(*itemID, paths, client)
		if err != nil {
			return err
		}
	}

	return nil
}
