package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/erikbryant/wow/internal/itemcache"
	"github.com/erikbryant/wow/internal/wowapi"
)

var (
	passPhrase = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
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
  refresh -passPhrase <phrase>
`)
}

func main() {
	flag.Parse()

	if *passPhrase == "" {
		fmt.Println("You must specify `-passPhrase <phrase>`")
		usage()
	}

	wowapi.Init(*passPhrase, false)
	refreshCache()
	itemcache.Save()
}
