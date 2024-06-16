package main

// https://develop.battle.net/documentation

import (
	"flag"
	"fmt"
	"github.com/erikbryant/aes"
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/wowAPI"
	"log"
)

type Bargain struct {
	Quantity    int64
	UnitSavings int64
	Name        string
}

var (
	clientIDCrypt     = "f7FhewxUd0lWQz/zPb27ZcwI/ZqkaMyd5YyuskFyEugQEeiKsfL7dvr11Kx1Y+Mi23qMciOAPe5ksCOy"
	clientSecretCrypt = "CtJH62iU6V3ZeqiHyKItECHahdUYgAFyfHmQ4DRabhWIv6JeK5K4dT7aiybot6MS4JitmDzuWSz1UHHv"
	clientID          string
	clientSecret      string
	passPhrase        = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	realm             = flag.String("realm", "Sisters of Elune", "WoW realm")
	usefulGoods       = map[int64]int64{
		// Health
		34722: 75000, // Heavy Frostweave Bandage

		// Enchanting
		34057:  7900,   // Abyss Crystal
		7909:   7500,   // Aquamarine
		22445:  12100,  // Arcane Dust
		124440: 29800,  // Arkhana
		124442: 295000, // Chaos Crystal
		109693: 7400,   // Draenic Dust
		//: 117700, // Dragon's Teeth
		//: 740000, // Elixir of Demonslaying
		7082:   247300,  // Essence of Air
		7076:   3500,    // Essence of Earth
		7078:   9400,    // Essence of Fire
		12808:  639700,  // Essence of Undeath
		7080:   458700,  // Essence of Water
		23427:  350000,  // Eternium Ore
		22794:  59900,   // Fel Lotus
		124116: 836000,  // Felhide
		124106: 1099700, // Felwort
		4625:   55000,   // Firebloom
		34056:  1,       // Lesser Cosmic Essence
	}
)

// jsonToStruct converts a single auction json string into a struct that is much easier to work with
func jsonToStruct(auc map[string]interface{}) wowAPI.Auction {
	var auction wowAPI.Auction

	auction.Id = web.ToInt64(auc["id"])

	_, ok := auc["item"]
	if !ok {
		fmt.Println("Auction had no item: ", auc)
		return wowAPI.Auction{}
	}
	item := auc["item"].(map[string]interface{})
	auction.ItemId = web.ToInt64(item["id"])

	if _, ok := auc["buyout"]; ok {
		// Regular auction
		auction.Buyout = web.ToInt64(auc["buyout"])
	} else {
		if _, ok := auc["unit_price"]; ok {
			// Commodity auction
			auction.Buyout = web.ToInt64(auc["unit_price"])
		}
	}

	auction.Quantity = web.ToInt64(auc["quantity"])

	return auction
}

// unpackAuction converts the []interface{} format we get from the web into structs
func unpackAuctions(a1, a2 []interface{}) map[int64][]wowAPI.Auction {
	auctions := map[int64][]wowAPI.Auction{}

	for _, a := range a1 {
		auction := jsonToStruct(a.(map[string]interface{}))
		auctions[auction.ItemId] = append(auctions[auction.ItemId], auction)
	}

	for _, a := range a2 {
		auction := jsonToStruct(a.(map[string]interface{}))
		auctions[auction.ItemId] = append(auctions[auction.ItemId], auction)
	}

	return auctions
}

// arbitrage returns auctions selling for lower than vendor prices
func arbitrage(auctions map[int64]wowAPI.Auction, accessToken string) (toBuy []wowAPI.Auction) {
	for _, auction := range auctions {
		item, ok := wowAPI.LookupItem(auction.ItemId, accessToken)
		if !ok {
			fmt.Println("ERROR: Unable to lookup item for auction: ", auction)
			continue
		}
		if item.Equippable {
			// I do not understand how to price these
			continue
		}
		if auction.Buyout > 0 && auction.Buyout < item.SellPrice {
			toBuy = append(toBuy, auction)
		}
	}

	return toBuy
}

// coinsToString returns a human-readable, formatted version of the coin amount
func coinsToString(amount int64) string {
	sign := ""
	if amount < 0 {
		sign = "-"
		amount *= -1
	}

	copper := amount % 100
	silver := (amount / 100) % 100
	gold := amount / 10000

	if gold > 0 {
		return fmt.Sprintf("%s%d.%02d.%02d", sign, gold, silver, copper)
	}

	if silver > 0 {
		return fmt.Sprintf("%s%d.%02d", sign, silver, copper)
	}

	return fmt.Sprintf("%s%d", sign, copper)
}

// printShoppingList prints a list of auctions the user should consider bidding/buying
func printShoppingList(action string, auctions []wowAPI.Auction, accessToken string) {
	for _, auction := range auctions {
		item, ok := wowAPI.LookupItem(auction.ItemId, accessToken)
		if !ok {
			fmt.Println("ERROR: Unable to lookup item for shopping list auction: ", auction)
			continue
		}
		profitBuy := item.SellPrice - auction.Buyout
		fmt.Printf("%s %25s \t quantity: %d \t profit: %s\n", action, item.Name, auction.Quantity, coinsToString(profitBuy))
	}

	fmt.Println()
}

// findBargains returns auctions for which the goods are below our desired prices
func findBargains(goods map[int64]int64, auctions map[int64][]wowAPI.Auction, accessToken string) []Bargain {
	bargains := make([]Bargain, 0)

	for itemId, maxPrice := range goods {
		item, ok := wowAPI.LookupItem(itemId, accessToken)
		if !ok {
			fmt.Println("ERROR: Unable to lookup item for bargain: ", itemId)
			continue
		}
		for _, auction := range auctions[itemId] {
			if auction.Buyout < maxPrice {
				bargain := Bargain{
					Quantity:    auction.Quantity,
					UnitSavings: maxPrice - auction.Buyout,
					Name:        item.Name,
				}
				bargains = append(bargains, bargain)
			}
		}
	}

	return bargains
}

// printBargains prints a list of auctions the user should consider bidding/buying
func printBargains(bargains []Bargain) {
	for _, bargain := range bargains {
		fmt.Printf("%20s \t quantity: %5d \t savings: %10s\n", bargain.Name, bargain.Quantity, coinsToString(bargain.UnitSavings))
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

	clientID, err := aes.Decrypt(clientIDCrypt, *passPhrase)
	if err != nil {
		log.Fatal(err)
	}

	clientSecret, err = aes.Decrypt(clientSecretCrypt, *passPhrase)
	if err != nil {
		log.Fatal(err)
	}

	accessToken, ok := wowAPI.AccessToken(clientID, clientSecret)
	if !ok {
		log.Fatal("ERROR: Unable to obtain access token.")
	}

	a, ok := wowAPI.Auctions(*realm, accessToken)
	if !ok {
		log.Fatal("ERROR: Unable to obtain auctions.")
	}
	fmt.Printf("#Auctions: %d\n\n", len(a))

	c, ok := wowAPI.Commodities(accessToken)
	if !ok {
		log.Fatal("ERROR: Unable to obtain commodity auctions.")
	}
	fmt.Printf("#Commodities: %d\n\n", len(c))

	auctions := unpackAuctions(a, c)

	// Look for findBargains

	bargains := findBargains(usefulGoods, auctions, accessToken)
	printBargains(bargains)
}
