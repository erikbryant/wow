package main

// https://develop.battle.net/documentation

import (
	"flag"
	"fmt"
	"github.com/erikbryant/aes"
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/wowAPI"
	"github.com/erikbryant/wowdb"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

var (
	clientIDCrypt     = "f7FhewxUd0lWQz/zPb27ZcwI/ZqkaMyd5YyuskFyEugQEeiKsfL7dvr11Kx1Y+Mi23qMciOAPe5ksCOy"
	clientSecretCrypt = "CtJH62iU6V3ZeqiHyKItECHahdUYgAFyfHmQ4DRabhWIv6JeK5K4dT7aiybot6MS4JitmDzuWSz1UHHv"
	clientID          string
	clientSecret      string
	passPhrase        = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	realm             = flag.String("realm", "Sisters of Elune", "WoW realm")
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

// unpackAuction converts the []interface{} format we get from the web into a map of structs
func unpackAuctions(a1, a2 []interface{}) map[int64]wowAPI.Auction {
	auctions := make(map[int64]wowAPI.Auction)

	for _, a := range a1 {
		auction := jsonToStruct(a.(map[string]interface{}))
		auctions[auction.Id] = auction
	}

	for _, a := range a2 {
		auction := jsonToStruct(a.(map[string]interface{}))
		auctions[auction.Id] = auction
	}

	return auctions
}

// bargains returns auctions for which the goods are at or below our desired prices
func bargains(goods map[int64]int64, auctions map[int64]wowAPI.Auction) (toBuy []wowAPI.Auction) {
	for _, auction := range auctions {
		if _, ok := goods[auction.ItemId]; !ok {
			// We do not need this item
			continue
		}

		maxPrice := goods[auction.ItemId]
		if auction.Buyout <= maxPrice {
			toBuy = append(toBuy, auction)
		}
	}

	return toBuy
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

// printBargains prints a list of auctions the user should consider bidding/buying
func printBargains(action string, auctions []wowAPI.Auction, goods map[int64]int64, accessToken string) {
	for _, auction := range auctions {
		item, ok := wowAPI.LookupItem(auction.ItemId, accessToken)
		if !ok {
			fmt.Println("ERROR: Unable to lookup item for shopping list auction: ", auction)
			continue
		}
		savings := goods[auction.ItemId] - auction.Buyout
		fmt.Printf("%s %25s \t quantity: %d \t savings: %s\n", action, item.Name, auction.Quantity, coinsToString(savings))
	}

	fmt.Println()
}

// usage prints a usage message and terminates the program with an error
func usage() {
	fmt.Println("Usage: wow -passPhrase <phrase>")
	os.Exit(1)
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

	wowdb.Open()
	defer wowdb.Close()

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

	// Look for bargains

	goods := map[int64]int64{
		// Health
		34722: 13800, // Heavy Frostweave Bandage

		// Enchanting
		34057:  8600,   // Abyss Crystal
		7909:   7500,   // Aquamarine
		22445:  12100,  // Arcane Dust
		124440: 29800,  // Arkhana
		124442: 492500, // Chaos Crystal
		109693: 10000,  // Draenic Dust
		//: 740000, // Elixir of Demonslaying
		7082:   257900,  // Essence of Air
		7076:   3500,    // Essence of Earth
		7078:   11400,   // Essence of Fire
		12808:  641000,  // Essence of Undeath
		7080:   458700,  // Essence of Water
		23427:  350000,  // Eternium Ore
		22794:  59900,   // Fel Lotus
		124116: 859700,  // Felhide
		124106: 1580000, // Felwort
		4625:   102400,  // Firebloom
		34056:  1,       // Lesser Cosmic Essence
	}

	toBuy := bargains(goods, auctions)
	printBargains("Buy!", toBuy, goods, accessToken)
}
