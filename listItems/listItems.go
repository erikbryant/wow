package main

import (
	"flag"
	"fmt"
	"github.com/erikbryant/wow/cache"
	"github.com/erikbryant/wow/wowAPI"
	"log"
)

var (
	passPhrase = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	itemId     = flag.Int64("id", 0, "Item ID to look up")
)

// usage prints a usage message and terminates the program with an error
func usage() {
	log.Fatal("Usage: wow -passPhrase <phrase>")
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

	accessToken, ok := wowAPI.AccessToken(*passPhrase)
	if !ok {
		log.Fatal("ERROR: Unable to obtain access token.")
	}

	wowAPI.LookupItem(*itemId, accessToken)
}
