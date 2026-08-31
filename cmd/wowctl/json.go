package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

func asJSON(itemID int64, paths *path.Paths) error {
	wowItems, err := wowitem.New(paths.Items)
	if err != nil {
		return err
	}

	client, err := wowapi.NewClient()
	if err != nil {
		return err
	}

	i, err := wowItems.Get(itemID, client)
	if err != nil {
		return fmt.Errorf("failed to get itemID %d: %w", itemID, err)
	}

	output.JSON(os.Stdout, []wowitem.Item{i})

	return nil
}

func runJSON(args []string, paths *path.Paths) error {
	flags := flag.NewFlagSet("json", flag.ExitOnError)

	itemID := flags.Int64("id", -1, "Item ID to look up")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *itemID == -1 {
		return fmt.Errorf("json requires -id")
	}

	err := asJSON(*itemID, paths)
	if err != nil {
		return err
	}

	return nil
}
