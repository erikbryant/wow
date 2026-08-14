package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/output"
	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/wowitem"
)

func json(itemID int64) error {
	paths, err := path.New("")
	if err != nil {
		return err
	}
	wowItems := wowitem.New(paths.Items)

	i, err := wowItems.Get(itemID)
	if err != nil {
		return fmt.Errorf("failed to get itemID %d: %w", itemID, err)
	}

	output.JSON(os.Stdout, []wowitem.Item{i})

	return nil
}

func runJSON(args []string) error {
	flags := flag.NewFlagSet("json", flag.ExitOnError)

	itemID := flags.Int64("id", -1, "Item ID to look up")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *itemID == -1 {
		return fmt.Errorf("json requires -id")
	}

	err := json(*itemID)
	if err != nil {
		return err
	}

	return nil
}
