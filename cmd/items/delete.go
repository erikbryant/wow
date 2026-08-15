package main

import (
	"flag"
	"fmt"

	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/wowitem"
)

func deleteItem(itemID int64, paths *path.Paths) error {
	wowItems, err := wowitem.New(paths.Items)
	if err != nil {
		return err
	}

	wowItems.Delete(itemID)
	fmt.Println("Deleted itemID:", itemID)
	err = wowItems.Save()
	if err != nil {
		return err
	}

	return nil
}

func runDelete(args []string, paths *path.Paths) error {
	flags := flag.NewFlagSet("delete", flag.ExitOnError)

	itemID := flags.Int64("id", -1, "Item ID to look up")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *itemID == -1 {
		return fmt.Errorf("delete requires -id")
	}

	err := deleteItem(*itemID, paths)
	if err != nil {
		return err
	}

	return nil
}
