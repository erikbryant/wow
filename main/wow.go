package main

// https://develop.battle.net/documentation

import (
	"flag"
	"fmt"
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/auction"
	"github.com/erikbryant/wow/battlePet"
	"github.com/erikbryant/wow/cache"
	"github.com/erikbryant/wow/common"
	"github.com/erikbryant/wow/wowAPI"
	"log"
	"sort"
	"strings"
)

type Bargain struct {
	Quantity    int64
	UnitSavings int64
	Name        string
	ItemLevel   int64
}

var (
	passPhrase  = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	realms      = flag.String("realms", "Commodities,Sisters of Elune,IceCrown,Drak'thul", "WoW realms")
	readThrough = flag.Bool("readThrough", false, "Read live values")
	migrate     = flag.Bool("migrate", false, "Migrate to new item cache data format")
	usefulGoods = map[int64]int64{
		// Generally useful items
		//158212: common.Coins(30, 0, 0), // Crow's Nest Scope
		//59596:  common.Coins(20, 0, 0), // Safety Catch Removal Kit
		//194017: common.Coins(50, 0, 0), // Wildercloth Bag

		// Summoners (versus pet cages) for battle pets I do not have yet
		152878: common.Coins(100, 0, 0), // Enchanted Tiki Mask

		// Enchanting recipes I do not have yet
		//210175: common.Coins(300, 0, 0), // Formula: Enchant Weapon - Dreaming Devotion
	}
)

// jsonToStruct converts a single auction json string into a struct that is much easier to work with
func jsonToStruct(auc map[string]interface{}) auction.Auction {
	var a auction.Auction

	a.Id = web.ToInt64(auc["id"])

	_, ok := auc["item"]
	if !ok {
		fmt.Println("Auction had no item: ", auc)
		return auction.Auction{}
	}
	item := auc["item"].(map[string]interface{})
	a.ItemId = web.ToInt64(item["id"])

	// Is this a Pet Cage?
	if a.ItemId == 82800 {
		// A pet a!
		a.Pet.BreedId = web.ToInt64(item["pet_breed_id"])
		a.Pet.Level = web.ToInt64(item["pet_level"])
		a.Pet.QualityId = web.ToInt64(item["pet_quality_id"])
		a.Pet.SpeciesId = web.ToInt64(item["pet_species_id"])
	}

	if _, ok := auc["buyout"]; ok {
		// Regular a
		a.Buyout = web.ToInt64(auc["buyout"])
	} else {
		if _, ok := auc["unit_price"]; ok {
			// Commodity a
			a.Buyout = web.ToInt64(auc["unit_price"])
		} else {
			fmt.Println("Unknown a type:", auc)
		}
	}

	a.Quantity = web.ToInt64(auc["quantity"])

	return a
}

// unpackAuction converts the []interface{} format we get from the web into structs
func unpackAuctions(auctions []interface{}) map[int64][]auction.Auction {
	a := map[int64][]auction.Auction{}

	for _, auc := range auctions {
		aucStruct := jsonToStruct(auc.(map[string]interface{}))
		if wowAPI.SkipItem(aucStruct.ItemId) {
			continue
		}
		a[aucStruct.ItemId] = append(a[aucStruct.ItemId], aucStruct)
	}

	return a
}

// findBargains returns auctions for which the goods are below our desired prices
func findBargains(goods map[int64]int64, auctions map[int64][]auction.Auction, accessToken string) []Bargain {
	bargains := []Bargain{}

	for itemId, maxPrice := range goods {
		item, ok := wowAPI.LookupItem(itemId, accessToken)
		if !ok {
			continue
		}
		for _, auc := range auctions[itemId] {
			if auc.Buyout <= 0 {
				continue
			}
			if auc.Buyout < maxPrice {
				bargain := Bargain{
					Quantity:    auc.Quantity,
					UnitSavings: maxPrice - auc.Buyout,
					Name:        item.Name,
					ItemLevel:   item.ItemLevel,
				}
				bargains = append(bargains, bargain)
			}
		}
	}

	sort.Slice(bargains, func(i, j int) bool {
		return bargains[i].Name < bargains[j].Name
	})

	return bargains
}

// findArbitrages returns auctions selling for lower than vendor prices
func findArbitrages(auctions map[int64][]auction.Auction, accessToken string) []Bargain {
	bargains := []Bargain{}

	for itemId, itemAuctions := range auctions {
		item, ok := wowAPI.LookupItem(itemId, accessToken)
		if !ok {
			continue
		}
		if item.Equippable {
			// Don't know how to price these
			continue
		}
		for _, auc := range itemAuctions {
			if auc.Buyout <= 0 {
				continue
			}
			if auc.Buyout < item.SellPrice {
				bargain := Bargain{
					Quantity:    auc.Quantity,
					UnitSavings: item.SellPrice - auc.Buyout,
					Name:        item.Name,
					ItemLevel:   item.ItemLevel,
				}
				bargains = append(bargains, bargain)
			}
		}
	}

	sort.Slice(bargains, func(i, j int) bool {
		return bargains[i].Name < bargains[j].Name
	})

	return bargains
}

// printShoppingList prints a list of auctions the user should buy
func printShoppingList(label string, bargains []Bargain) {
	names := []string{}

	for _, bargain := range bargains {
		if bargain.ItemLevel > 0 {
			// I don't know how to price these yet
			//names = append(names, fmt.Sprintf("%-50s iLvl %3d", bargain.Name, bargain.ItemLevel))
			continue
		} else {
			names = append(names, bargain.Name)
		}
	}

	if len(names) > 0 {
		fmt.Printf("--- %s ---\n", label)
		fmt.Println(strings.Join(common.SortUnique(names), "\n"))
		fmt.Println()
	}
}

// getAuctions returns the current auctions and their hash
func getAuctions(realm, accessToken string) (map[int64][]auction.Auction, bool) {
	var ok bool
	var auctions []interface{}

	if strings.ToLower(realm) == "commodities" {
		auctions, ok = wowAPI.Commodities(accessToken)
	} else {
		auctions, ok = wowAPI.Auctions(realm, accessToken)
	}
	if !ok {
		log.Println("ERROR: Unable to obtain auctions for", realm)
		return nil, false
	}

	return unpackAuctions(auctions), true
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
		auctions, ok := getAuctions(realm, accessToken)
		if !ok {
			continue
		}
		fmt.Printf("====== %s (%d unique items) ======\n\n", realm, len(auctions))
		printBargains(auctions, accessToken)
		printPetBargains(auctions)
	}
	fmt.Println()
}

// usage prints a usage message and terminates the program with an error
func usage() {
	log.Fatal("Usage: wow -passPhrase <phrase>")
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

	if *migrate {
		cache.Migrate()
		return
	}

	if *readThrough {
		// Get the latest values
		cache.DisableRead()
	}

	doit(accessToken, *realms)
}
