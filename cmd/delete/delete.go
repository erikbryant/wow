package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/erikbryant/wow/internal/itemcache"
)

var (
	itemId = flag.Int64("id", -1, "Item ID to look up")
)

// usage prints a usage message and terminates the program with an error
func usage() {
	log.Fatal(`Usage:
  delete -id <itemId>   # Delete <itemId> from the cache
`)
}

func main() {
	flag.Parse()

	if *itemId == -1 {
		fmt.Println("You must specify `-id <itemId>`")
		usage()
	}

	fmt.Println("Deleting itemId:", *itemId)
	itemcache.Delete(*itemId)
	itemcache.Save()
}
