package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/wowitem"
)

func deleteItem(itemID int64) error {
	fmt.Println("Deleting itemID:", itemID)
	wowitem.Items.Delete(itemID)
	err := wowitem.Items.Save()
	return err
}

func runDelete(args []string) {
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

	err := deleteItem(*itemID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
