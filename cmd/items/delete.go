package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/wowitem"
)

func deleteItem(itemID int64, wowItem *wowitem.WoWItem) error {
	wowItem.Items.Delete(itemID)
	fmt.Println("Deleted itemID:", itemID)
	err := wowItem.Items.Save()
	return err
}

func runDelete(args []string, wowItem *wowitem.WoWItem) {
	flags := flag.NewFlagSet("delete", flag.ExitOnError)

	itemID := flag.Int64("id", -1, "Item ID to look up")

	if err := flags.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if *itemID == -1 {
		fmt.Fprintln(os.Stderr, "delete requires -id")
		os.Exit(2)
	}

	err := deleteItem(*itemID, wowItem)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
