package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/erikbryant/wow/internal/itemcache"
	"github.com/erikbryant/wow/internal/wowapi"
)

var (
	passPhrase  = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	readThrough = flag.Bool("readThrough", false, "Read live values")
	refresh     = flag.Bool("refresh", false, "Refresh cached values")
	delItem     = flag.Bool("delItem", false, "Delete cached value")
	itemId      = flag.Int64("id", 0, "Item ID to look up")
	full        = flag.Bool("full", false, "Display wowitem details")
)

// refreshCache refreshes any cached items older than a certain age
func refreshCache() {
	maxAge := 24 * time.Hour * 7 // 1 week
	needsRefresh := 0
	refreshCount := 0
	maxRefreshCount := 1000

	for _, i := range itemcache.ItemsCopy() {
		if i.Stale(maxAge) {
			needsRefresh++
			if refreshCount < maxRefreshCount {
				itemcache.LookupItem(i.ID(), maxAge)
				refreshCount++
			}
		}
	}

	itemcache.Save()

	fmt.Printf("Refreshed %d of %d stale items\n", refreshCount, needsRefresh)
}

// usage prints a usage message and terminates the program with an error
func usage() {
	log.Fatal(`Usage:
  listitems                                              # Print the entire cache
  listitems -passPhrase <phrase> -id <itemId>            # Print a single wowitem
  listitems -passPhrase <phrase> -refresh                # Refresh items in the cache
  listitems -passPhrase <phrase> -delItem -id <itemId>   # Delete <itemId> from the cache
`)
}

func main() {
	flag.Parse()

	if *itemId == 0 && !*refresh && !*delItem {
		// If no flags, list the whole cache
		itemcache.Print()
		return
	}

	if *passPhrase == "" {
		fmt.Println("You must specify `-passPhrase <phrase>`")
		usage()
	}

	wowapi.Init(*passPhrase, false)

	if *delItem {
		if *itemId == 0 {
			fmt.Println("You must specify `-id <itemId>`")
			usage()
		}
		fmt.Println("Deleting itemId:", *itemId)
		itemcache.Delete(*itemId)
		itemcache.Save()
		return
	}

	if *refresh {
		refreshCache()
		return
	}

	if *readThrough {
		// Get the latest values
		itemcache.DisableRead()
	}

	i, ok := itemcache.LookupItem(*itemId, 0)
	if !ok {
		log.Fatal("Failed to LookupItem: ", *itemId)
	}

	if *full {
		b, _ := json.MarshalIndent(i.XItem, "\t", "\t")
		fmt.Println(string(b))
	} else {
		fmt.Println(i.Format())
	}

	itemcache.Save()
}
