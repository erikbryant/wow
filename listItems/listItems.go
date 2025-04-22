package main

import (
	"flag"
	"fmt"
	"github.com/erikbryant/wow/cache"
	"github.com/erikbryant/wow/wowAPI"
	"log"
	"time"
)

var (
	passPhrase  = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	readThrough = flag.Bool("readThrough", false, "Read live values")
	refresh     = flag.Bool("refresh", false, "Refresh cached values")
	delete      = flag.Bool("delete", false, "Delete cached value")
	itemId      = flag.Int64("id", 0, "Item ID to look up")
)

// refreshCache refreshes any cached items older than a certain age
func refreshCache() {
	maxAge := 24 * time.Hour * 7 // 1 week
	needsRefresh := 0
	refreshCount := 0
	maxRefreshCount := 1000

	for _, i := range cache.Items() {
		if wowAPI.Stale(i, maxAge) {
			needsRefresh++
		}
	}

	for _, i := range cache.Items() {
		if wowAPI.Stale(i, maxAge) {
			wowAPI.LookupItem(i.Id(), maxAge)
			refreshCount++
		}
		if refreshCount >= maxRefreshCount {
			break
		}
	}

	cache.Save()

	fmt.Printf("Refreshed %d/%d items\n", refreshCount, needsRefresh)
}

// usage prints a usage message and terminates the program with an error
func usage() {
	log.Fatal(`Usage:
  listItems                                             # Print the entire cache
  listItems -passPhrase <phrase> -id <itemId>           # Print a single item
  listItems -passPhrase <phrase> -refresh               # Refresh items in the cache
  listItems -passPhrase <phrase> -delete -id <itemId>   # Delete <itemId> from the cache
`)
}

func main() {
	flag.Parse()

	if *itemId == 0 && !*refresh && !*delete {
		// If no flags, list the whole cache
		cache.Print()
		return
	}

	if *passPhrase == "" {
		fmt.Println("ERROR: You must specify `-passPhrase <phrase>`")
		usage()
	}

	wowAPI.Init(*passPhrase)

	if *delete {
		if *itemId == 0 {
			fmt.Println("You must specify `-id <itemId>`")
			usage()
		}
		fmt.Println("Deleting itemId:", *itemId)
		cache.Delete(*itemId)
		cache.Save()
		return
	}

	if *refresh {
		refreshCache()
		return
	}

	if *readThrough {
		// Get the latest values
		cache.DisableRead()
	}

	i, ok := wowAPI.LookupItem(*itemId, 0)
	if !ok {
		log.Fatal("Failed to LookupItem: ", *itemId)
	}

	fmt.Println(i.Format())
	cache.Save()
}
