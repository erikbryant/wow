package main

// https://develop.battle.net/documentation

import (
	"flag"
	"fmt"
	"github.com/erikbryant/wow/auction"
	"github.com/erikbryant/wow/battlePet"
	"github.com/erikbryant/wow/cache"
	"github.com/erikbryant/wow/common"
	"github.com/erikbryant/wow/wowAPI"
	"log"
	"os"
	"strings"
)

var (
	passPhrase  = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	realms      = flag.String("realms", "Commodities,Sisters of Elune,IceCrown,Drak'thul,Eitrigg", "WoW realms")
	readThrough = flag.Bool("readThrough", false, "Read live values")
	usefulGoods = map[int64]int64{
		// Generally useful items
		92741: common.Coins(5000, 0, 0), // Flawless Battle-Stone

		// Summoners (versus pet cages) for battle pets I do not have yet
		152878: common.Coins(100, 0, 0), // Enchanted Tiki Mask

		// Enchanting recipes I do not have yet
		//210175: common.Coins(300, 0, 0), // Formula: Enchant Weapon - Dreaming Devotion
	}
)

// findBargains returns auctions for which the goods are below our desired prices
func findBargains(goods map[int64]int64, auctions map[int64][]auction.Auction, accessToken string) []string {
	bargains := []string{}

	for itemId, maxPrice := range goods {
		item, ok := wowAPI.LookupItem(itemId, accessToken)
		if !ok {
			continue
		}
		if item.Equippable() {
			// Don't know how to price these
			continue
		}
		for _, auc := range auctions[itemId] {
			if auc.Buyout <= 0 {
				continue
			}
			if auc.Buyout < maxPrice {
				bargains = append(bargains, item.Name())
			}
		}
	}

	return bargains
}

// findArbitrages returns auctions selling for lower than vendor prices
func findArbitrages(auctions map[int64][]auction.Auction, accessToken string) []string {
	bargains := []string{}

	for itemId, itemAuctions := range auctions {
		item, ok := wowAPI.LookupItem(itemId, accessToken)
		if !ok {
			continue
		}
		for _, auc := range itemAuctions {
			if auc.Buyout <= 0 {
				continue
			}
			if auc.Buyout < item.SellPrice() {
				bargains = append(bargains, item.Name())
			}
		}
	}

	return bargains
}

// printShoppingList prints a list of auctions the user should buy
func printShoppingList(label string, names []string) {
	if len(names) > 0 {
		fmt.Printf("--- %s ---\n", label)
		fmt.Println(strings.Join(common.SortUnique(names), "\n"))
		fmt.Println()
	}
}

// printPetBargains prints a list of pets the user should buy
func printPetBargains(auctions map[int64][]auction.Auction) {
	bargains := []string{}

	for _, petAuction := range auctions[battlePet.PetCageItemId] {
		if battlePet.Own(petAuction.Pet.SpeciesId) {
			continue
		}
		if petAuction.Buyout > 1000000 {
			continue
		}
		if petAuction.Pet.QualityId < common.QualityId("Rare") {
			continue
		}
		bargains = append(bargains, battlePet.Name(petAuction.Pet.SpeciesId))
	}

	if len(bargains) > 0 {
		fmt.Println("--- Pet auction bargains ---")
		fmt.Println(strings.Join(common.SortUnique(bargains), "\n"))
		fmt.Println()
	}
}

// printBargains prints the bargains found in the auction house
func printBargains(auctions map[int64][]auction.Auction, accessToken string) {
	toBuy := findBargains(usefulGoods, auctions, accessToken)
	printShoppingList("Bargains", toBuy)
	toBuy = findArbitrages(auctions, accessToken)
	printShoppingList("Arbitrages", toBuy)
}

// doit downloads the available auctions and prints any bargains/arbitrages
func doit(accessToken string, realmList string) {
	battlePet.Init(accessToken)

	for _, realm := range strings.Split(realmList, ",") {
		cache.DisableWrite()

		auctions, ok := auction.GetAuctions(realm, accessToken)
		if !ok {
			continue
		}
		fmt.Printf("====== %s (%d unique items) ======\n\n", realm, len(auctions))
		printBargains(auctions, accessToken)
		printPetBargains(auctions)

		cache.EnableWrite()
		cache.Save()
	}
	fmt.Println()
}

// writeFile writes contents to file
func writeFile(file, contents string) {
	f, err := os.Create(file)
	if err != nil {
		log.Fatal("Failed to create file:", file, err)
	}
	_, err = f.WriteString(contents)
	f.Close()
}

// generateLua writes the WoW addon lua files
func generateLua() {
	writeFile("./generated/PriceCache.lua", cache.LuaVendorPrice())
	writeFile("./generated/PetCache.lua", battlePet.LuaPetId())
}

// usage prints a usage message and terminates the program with an error
func usage() {
	log.Fatal(`Usage:
  wow -passPhrase <phrase> [-realms <realms>] [-readThrough]
`)
}

func main() {
	flag.Parse()

	if *passPhrase == "" {
		fmt.Println("ERROR: You must specify -passPhrase to unlock the client Id/secret")
		usage()
	}

	accessToken, ok := wowAPI.AccessToken(*passPhrase)
	if !ok {
		log.Fatal("ERROR: Unable to obtain access token.")
	}

	if *readThrough {
		// Get the latest values
		cache.DisableRead()
	}

	doit(accessToken, *realms)

	generateLua()
}
