package main

import (
	"flag"
	"fmt"
	"github.com/erikbryant/wow/cache"
	"github.com/erikbryant/wow/item"
	"github.com/erikbryant/wow/wowAPI"
	"log"
	"time"
)

var (
	passPhrase  = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	readThrough = flag.Bool("readThrough", false, "Read live values")
	refresh     = flag.Bool("refresh", false, "Refresh cached values")
	itemId      = flag.Int64("id", 0, "Item ID to look up")
)

// usage prints a usage message and terminates the program with an error
func usage() {
	log.Fatal("Usage: listItems -passPhrase <phrase>")
}

// stale returns whether the item is older than an arbitrary time
func stale(i item.Item) bool {
	return time.Now().Sub(i.Updated()) > 1*24*time.Hour
}

// refreshCache refreshes any cached items older than 1 day
func refreshCache() {
	refreshCount := 0
	maxRefreshCount := 1000

	cache.DisableWrite()
	for _, i := range cache.Items() {
		if stale(i) {
			fmt.Println("Refreshing stale item:", i.Format())
			cache.DisableRead()
			wowAPI.LookupItem(i.Id())
			cache.EnableRead()
			refreshCount++
		}
		if refreshCount >= maxRefreshCount {
			break
		}
	}
	cache.EnableWrite()

	cache.Save()

	fmt.Printf("Refreshed %d/%d items\n", refreshCount, maxRefreshCount)
}

func main() {
	flag.Parse()

	if *itemId == 0 && !*refresh {
		cache.Print()
		return
	}

	if *passPhrase == "" {
		fmt.Println("ERROR: You must specify -passPhrase to unlock the client Id/secret")
		usage()
	}

	wowAPI.Init(*passPhrase)

	if *refresh {
		refreshCache()
		return
	}

	if *readThrough {
		// Get the latest values
		cache.DisableRead()
	}

	i, ok := wowAPI.LookupItem(*itemId)
	if !ok {
		return
	}
	fmt.Println(i.Format())
}
