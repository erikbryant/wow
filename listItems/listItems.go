package main

import (
	"flag"
	"fmt"
	"github.com/erikbryant/wow/cache"
	"github.com/erikbryant/wow/wowAPI"
	"log"
)

var (
	passPhrase  = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	readThrough = flag.Bool("readThrough", false, "Read live values")
	itemId      = flag.Int64("id", 0, "Item ID to look up")
)

// usage prints a usage message and terminates the program with an error
func usage() {
	log.Fatal("Usage: listItems -passPhrase <phrase>")
}

func main() {
	flag.Parse()

	if *itemId == 0 {
		cache.Print()
		return
	}

	if *passPhrase == "" {
		fmt.Println("ERROR: You must specify -passPhrase to unlock the client Id/secret")
		usage()
	}

	wowAPI.Init(*passPhrase)

	accessToken, ok := wowAPI.AccessToken()
	if !ok {
		log.Fatal("ERROR: Unable to obtain access token.")
	}

	if *readThrough {
		// Get the latest values
		cache.DisableRead()
	}

	i, ok := wowAPI.LookupItem(*itemId, accessToken)
	if !ok {
		return
	}
	fmt.Println(i.Format())
}
