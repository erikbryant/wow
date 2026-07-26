package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/erikbryant/wow/internal/itemcache"
	"github.com/erikbryant/wow/internal/wowapi"
)

var (
	passPhrase  = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	readThrough = flag.Bool("readThrough", false, "Read live values")
	itemId      = flag.Int64("id", -1, "Item ID to look up")
	full        = flag.Bool("full", false, "Display item details")
)

// usage prints a usage message and terminates the program with an error
func usage() {
	log.Fatal(`Usage:
  listitems                                                  # Print the entire cache
  listitems -full                                            # Print full JSON, instead of single line summary
  listitems -id <itemId>                                     # Print a single item
  listitems -passPhrase <phrase> -readthrough -id <itemId>   # Print a single item, reading from web API
`)
}

func main() {
	flag.Parse()

	if *itemId == -1 {
		// If no flags, list the whole cache
		itemcache.Print()
		return
	}

	if *readThrough {
		if *passPhrase == "" {
			fmt.Println("You must specify `-passPhrase <phrase>`")
			usage()
		}

		wowapi.Init(*passPhrase, false)

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

	if *readThrough {
		itemcache.Save()
	}
}
