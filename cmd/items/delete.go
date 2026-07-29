package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/erikbryant/wow/internal/itemcache"
)

func deleteItem(itemID int64) {
	fmt.Println("Deleting itemId:", itemID)
	itemcache.Delete(itemID)
	err := itemcache.Save()
	if err != nil {
		log.Fatal(err)
	}
}

func runDelete(args []string) {
	flags := flag.NewFlagSet("delete", flag.ContinueOnError)

	itemID := flag.Int64("id", -1, "Item ID to look up")

	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}

	if *itemID == -1 {
		fmt.Fprintln(os.Stderr, "delete requires -id")
		os.Exit(2)
	}

	deleteItem(*itemID)
}
