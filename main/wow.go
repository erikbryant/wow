package main

// https://develop.battle.net/documentation

import (
	"flag"
	"fmt"
	"github.com/erikbryant/aes"
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/common"
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
		211943: 6000, // Scarlet Silk Bandage

		// Enchanting
		34057:  7900,    // Abyss Crystal
		7909:   5000,    // Aquamarine
		22445:  12100,   // Arcane Dust
		124440: 29800,   // Arkhana
		124442: 240000,  // Chaos Crystal
		109693: 7400,    // Draenic Dust
		3819:   117700,  // Dragon's Teeth
		9224:   730000,  // Elixir of Demonslaying
		7082:   200000,  // Essence of Air
		7076:   2000,    // Essence of Earth
		7078:   5000,    // Essence of Fire
		12808:  558000,  // Essence of Undeath
		7080:   458700,  // Essence of Water
		23427:  328800,  // Eternium Ore
		22794:  57500,   // Fel Lotus
		124116: 777500,  // Felhide
		124106: 1047700, // Felwort
		4625:   55000,   // Firebloom
		34056:  2400,    // Lesser Cosmic Essence
		16202:  39200,   // Lesser Eternal Essence
		10938:  4500,    // Lesser Magic Essence
		22447:  5000,    // Lesser Planar Essence
		124441: 89800,   // Leylight Shard
		16204:  600,     // Light Illusion Dust
		12803:  233900,  // Living Essence
		111245: 75700,   // Luminous Shard
		52722:  10100,   // Maelstrom Crystal
		22791:  80000,   // Netherbloom
		22792:  52500,   // Nightmare Vine
		22451:  4009900, // Primal Air
		22452:  74500,   // Primal Earth
		21884:  1745000, // Primal Fire
		21886:  958900,  // Primal Life
		22457:  1480000, // Primal Mana
		23571:  7235000, // Primal Might
		22456:  1059800, // Primal Shadow
		21885:  1064000, // Primal Water
		156930: 4600,    // Rich Illusion Dust
		14343:  700,     // Small Brilliant Shard
		22448:  4500,    // Small Prismatic Shard
		10940:  4400,    // Strange Dust
		113588: 72000,   // Temporal Crystal
		22450:  11300,   // Void Crystal
		52328:  159900,  // Volatile Air
		8153:   3500000, // WildVine

		// Tailoring
		111557: 8800, // Sumptuous Fur
	}
)

// jsonToStruct converts a single auction json string into a struct that is much easier to work with
func jsonToStruct(auc map[string]interface{}) common.Auction {
	var auction common.Auction

	auction.Id = web.ToInt64(auc["id"])

	_, ok := auc["item"]
	if !ok {
		fmt.Println("Auction had no item: ", auc)
		return common.Auction{}
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
func unpackAuctions(a1 []interface{}) map[int64][]common.Auction {
	auctions := map[int64][]common.Auction{}

	for _, a := range a1 {
		auction := jsonToStruct(a.(map[string]interface{}))
		if wowAPI.SkipItem(auction.ItemId) {
			continue
		}
		auctions[auction.ItemId] = append(auctions[auction.ItemId], auction)
	}

	return auctions
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

// findBargains returns auctions for which the goods are below our desired prices
func findBargains(goods map[int64]int64, auctions map[int64][]common.Auction, accessToken string) []Bargain {
	bargains := []Bargain{}

	for itemId, maxPrice := range goods {
		item, ok := wowAPI.LookupItem(itemId, accessToken)
		if !ok {
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

// printShoppingList prints a list of auctions the user should consider bidding/buying
func printShoppingList(bargains []Bargain) {
	for _, bargain := range bargains {
		fmt.Printf("%50s \t quantity: %5d \t savings: %10s\n", bargain.Name, bargain.Quantity, coinsToString(bargain.UnitSavings))
	}
	fmt.Println()
}

// findArbitrages returns auctions selling for lower than vendor prices
func findArbitrages(auctions map[int64][]common.Auction, accessToken string) []Bargain {
	bargains := []Bargain{}

	for itemId, aucs := range auctions {
		item, ok := wowAPI.LookupItem(itemId, accessToken)
		if !ok {
			continue
		}
		for _, auction := range aucs {
			if item.Equippable {
				// I do not understand how to price these
				continue
			}
			if auction.Buyout < item.SellPrice {
				bargain := Bargain{
					Quantity:    auction.Quantity,
					UnitSavings: item.SellPrice - auction.Buyout,
					Name:        item.Name,
				}
				bargains = append(bargains, bargain)
			}
		}
	}

	return bargains
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

	auctions := unpackAuctions(a)
	commodities := unpackAuctions(c)

	// Look for things to buy
	fmt.Println("Bargains:")
	toBuy := findBargains(usefulGoods, auctions, accessToken)
	printShoppingList(toBuy)
	toBuy = findBargains(usefulGoods, commodities, accessToken)
	printShoppingList(toBuy)
	fmt.Println("Arbitrages:")
	toBuy = findArbitrages(commodities, accessToken)
	printShoppingList(toBuy)
}
