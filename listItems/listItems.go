package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/erikbryant/wow/itemCache"
	"github.com/erikbryant/wow/wowAPI"
	"log"
	"time"
)

var (
	passPhrase  = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	readThrough = flag.Bool("readThrough", false, "Read live values")
	refresh     = flag.Bool("refresh", false, "Refresh cached values")
	delItem     = flag.Bool("delItem", false, "Delete cached value")
	itemId      = flag.Int64("id", 0, "Item ID to look up")
	full        = flag.Bool("full", false, "Display item details")
)

// refreshCache refreshes any cached items older than a certain age
func refreshCache() {
	maxAge := 24 * time.Hour * 7 // 1 week
	needsRefresh := 0
	refreshCount := 0
	maxRefreshCount := 1000

	for _, i := range itemCache.Items() {
		if wowAPI.Stale(i, maxAge) {
			needsRefresh++
		}
	}

	for _, i := range itemCache.Items() {
		if wowAPI.Stale(i, maxAge) {
			wowAPI.LookupItem(i.Id(), maxAge)
			refreshCount++
		}
		if refreshCount >= maxRefreshCount {
			break
		}
	}

	itemCache.Save()

	fmt.Printf("Refreshed %d of %d stale items\n", refreshCount, needsRefresh)
}

// usage prints a usage message and terminates the program with an error
func usage() {
	log.Fatal(`Usage:
  listItems                                              # Print the entire cache
  listItems -passPhrase <phrase> -id <itemId>            # Print a single item
  listItems -passPhrase <phrase> -refresh                # Refresh items in the cache
  listItems -passPhrase <phrase> -delItem -id <itemId>   # Delete <itemId> from the cache
`)
}

func main() {
	flag.Parse()

	if *itemId == 0 && !*refresh && !*delItem {
		// If no flags, list the whole cache
		itemCache.Print()
		return
	}

	if *passPhrase == "" {
		fmt.Println("ERROR: You must specify `-passPhrase <phrase>`")
		usage()
	}

	wowAPI.Init(*passPhrase, false)

	if *delItem {
		if *itemId == 0 {
			fmt.Println("You must specify `-id <itemId>`")
			usage()
		}
		fmt.Println("Deleting itemId:", *itemId)
		itemCache.Delete(*itemId)
		itemCache.Save()
		return
	}

	if *refresh {
		refreshCache()
		return
	}

	if *readThrough {
		// Get the latest values
		itemCache.DisableRead()
	}

	i, ok := wowAPI.LookupItem(*itemId, 0)
	if !ok {
		log.Fatal("Failed to LookupItem: ", *itemId)
	}

	if *full {
		b, _ := json.MarshalIndent(i.XItem, "\t", "\t")
		fmt.Println(string(b))
	} else {
		fmt.Println(i.Format())
	}

	itemCache.Save()
}
