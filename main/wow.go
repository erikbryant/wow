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
	readThrough = flag.Bool("readThrough", false, "Read live values")
	realms      = flag.String("realms", "Drak'thul,Eitrigg,IceCrown,Sisters of Elune,Commodities", "WoW realms")
	realmsUS    = flag.Bool("realmsUS", false, "Scan all US realms")

	restOfUS = []string{ // US realms not in the default realm list
		"Aegwynn",
		"Agamaggan",
		"Aggramar",
		"Akama",
		"Alexstrasza",
		"Alleria",
		"Altar of Storms",
		"Alterac Mountains",
		"Andorhal",
		"Anub'arak",
		"Argent Dawn",
		"Azgalor",
		"Azjol-Nerub",
		"Azuremyst",
		"Baelgun",
		"Blackhand",
		"Blackwing Lair",
		"Bloodhoof",
		"Bloodscalp",
		"Bronzebeard",
		"Cairne",
		"Coilfang",
		"Darrowmere",
		"Deathwing",
		"Dentarg",
		"Draenor",
		"Dragonblight",
		"Durotan",
		"Elune",
		"Eredar",
		"Farstriders",
		"Feathermoon",
		"Frostwolf",
		"Ghostlands",
		"Greymane",
		"Kilrogg",
		"Kirin Tor",
		"Kul Tiras",
		"Lightninghoof",
		"Llane",
		"Misha",
		"Nazgrel",
		"Ravencrest",
		"Runetotem",
	}

	usefulGoods = map[int64]int64{
		// Generally useful items
		92741: common.Coins(5000, 0, 0), // Flawless Battle-Stone
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

// scanRealm downloads the available auctions and prints any bargains/arbitrages
func scanRealm(accessToken string, realm string) {
	cache.DisableWrite()
	auctions, ok := auction.GetAuctions(realm, accessToken)
	if !ok {
		return
	}
	fmt.Printf("===========>  %s (%d unique items)  <===========\n\n", realm, len(auctions))
	printBargains(auctions, accessToken)
	printPetBargains(auctions)
	cache.EnableWrite()
	cache.Save()
}

// writeFile writes contents to file
func writeFile(file, contents string) {
	f, err := os.Create(file)
	if err != nil {
		log.Fatal("Failed to create file:", file, err)
	}
	_, err = f.WriteString(contents)
	err = f.Close()
	if err != nil {
		fmt.Println("Failed to close file:", file, err)
	}
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

	profileAccessToken, ok := wowAPI.ProfileAccessToken(*passPhrase)
	if !ok {
		log.Fatal("ERROR: Unable to obtain profile access token.")
	}

	battlePet.Init(accessToken, profileAccessToken)

	if *readThrough {
		// Get the latest values
		cache.DisableRead()
	}

	var realmsToScan []string
	if *realmsUS {
		realmsToScan = restOfUS
	} else {
		realmsToScan = strings.Split(*realms, ",")
	}
	for _, realm := range realmsToScan {
		scanRealm(accessToken, realm)
		fmt.Println()
	}

	generateLua()
}
