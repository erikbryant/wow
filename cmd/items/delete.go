package main

import (
	"flag"
	"fmt"

	"github.com/erikbryant/wow/internal/path"
	"github.com/erikbryant/wow/internal/wowitem"
)

func deleteItem(itemID int64) error {
	paths, err := path.New("")
	if err != nil {
		return err
	}
	wowItems := wowitem.New(paths.Items)

	wowItems.Delete(itemID)
	fmt.Println("Deleted itemID:", itemID)
	err = wowItems.Save()

	return err
}

func runDelete(args []string) error {
	flags := flag.NewFlagSet("delete", flag.ExitOnError)

	itemID := flags.Int64("id", -1, "Item ID to look up")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *itemID == -1 {
		return fmt.Errorf("delete requires -id")
	}

	err := deleteItem(*itemID)
	if err != nil {
		return err
	}

	return nil
}
