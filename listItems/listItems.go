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
	lua        = flag.Bool("lua", false, "Generate lua table of all items")
)

// usage prints a usage message and terminates the program with an error
func usage() {
	log.Fatal("Usage: wow -passPhrase <phrase>")
}

func main() {
	flag.Parse()

	if *lua {
		cache.PrintLuaVendorPrice()
		fmt.Println(`
ItemCache = {
  VendorSellPrice = VendorSellPrice,
}`)
		return
	}

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

	item, ok := wowAPI.LookupItem(*itemId, accessToken)
	if !ok {
		return
	}
	fmt.Println(item.Format())
}
