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
	"sort"
	"time"
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
		7909:   2000,    // Aquamarine
		22445:  12100,   // Arcane Dust
		124440: 29800,   // Arkhana
		124442: 18700,   // Chaos Crystal
		109693: 5700,    // Draenic Dust
		3819:   97300,   // Dragon's Teeth
		9224:   700000,  // Elixir of Demonslaying
		7082:   170000,  // Essence of Air
		7076:   2000,    // Essence of Earth
		7078:   5000,    // Essence of Fire
		12808:  558000,  // Essence of Undeath
		7080:   458700,  // Essence of Water
		23427:  305000,  // Eternium Ore
		22794:  57500,   // Fel Lotus
		124116: 742400,  // Felhide
		124106: 999800,  // Felwort
		4625:   55000,   // Firebloom
		34056:  2400,    // Lesser Cosmic Essence
		16202:  35100,   // Lesser Eternal Essence
		10938:  4500,    // Lesser Magic Essence (720 in shop)
		22447:  5000,    // Lesser Planar Essence
		124441: 71500,   // Leylight Shard
		16204:  600,     // Light Illusion Dust
		12803:  233900,  // Living Essence
		111245: 31000,   // Luminous Shard
		52722:  10100,   // Maelstrom Crystal
		22791:  55000,   // Netherbloom
		22792:  39900,   // Nightmare Vine
		22451:  2750000, // Primal Air
		22452:  60000,   // Primal Earth
		21884:  1740000, // Primal Fire
		21886:  902400,  // Primal Life
		22457:  1480000, // Primal Mana
		23571:  7110000, // Primal Might
		22456:  1059800, // Primal Shadow
		21885:  906500,  // Primal Water
		156930: 3500,    // Rich Illusion Dust
		14343:  700,     // Small Brilliant Shard
		22448:  4500,    // Small Prismatic Shard
		10940:  1000,    // Strange Dust (720 in shop)
		113588: 65500,   // Temporal Crystal
		22450:  9700,    // Void Crystal
		52328:  150000,  // Volatile Air
		8153:   3500000, // WildVine

		// Tailoring
		111557: 8800, // Sumptuous Fur

		// Bags
		194017: 500000, // Wildercloth Bag

		// Item pricing research
		33428:  1000000, // Dulled Shiv
		2057:   1000000, // Pitted Defias Shortsword
		6563:   1000000, // Shimmering Bracers
		15248:  1000000, // Gleaming Claymore
		15212:  1000000, // Fighter Broadsword
		121110: 1000000, // Hagfeather Wristwraps
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

	sort.Slice(bargains, func(i, j int) bool {
		return bargains[i].Name < bargains[j].Name
	})

	return bargains
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

	sort.Slice(bargains, func(i, j int) bool {
		return bargains[i].Name < bargains[j].Name
	})

	return bargains
}

// printShoppingList prints a list of auctions the user should consider bidding/buying
func printShoppingList(label string, bargains []Bargain) {
	fmt.Printf("--- %s ---\n", label)

	lastName := ""
	for _, bargain := range bargains {
		if bargain.Name == lastName {
			// Only print an item once
			continue
		}
		fmt.Printf("%s\n", bargain.Name)
		lastName = bargain.Name
	}

	fmt.Println()
}

// hash returns a checksum hash of the given data
func hash(blob []interface{}) int64 {
	return int64(len(blob))
}

// getCommodities returns the current auctions and their hash
func getCommodities(accessToken string) (map[int64][]common.Auction, int64, bool) {
	auctions, ok := wowAPI.Commodities(accessToken)
	if !ok {
		log.Fatal("ERROR: Unable to obtain commodity auctions.")
	}
	return unpackAuctions(auctions), hash(auctions), true
}

// getAuctions returns the current auctions and their hash
func getAuctions(accessToken string) (map[int64][]common.Auction, int64, bool) {
	auctions, ok := wowAPI.Auctions(*realm, accessToken)
	if !ok {
		log.Fatal("ERROR: Unable to obtain auctions.")
	}
	return unpackAuctions(auctions), hash(auctions), true
}

// printBargains prints the bargains found in the auction house
func printBargains(auctions map[int64][]common.Auction, accessToken string, includeArbitrage bool) {
	toBuy := findBargains(usefulGoods, auctions, accessToken)
	printShoppingList(fmt.Sprintf("Bargains (%d)", len(auctions)), toBuy)
	if !includeArbitrage {
		// Non-commodity auctions are strange. They have items for sale.
		// These items have sell prices. But, the sell prices in the
		// item's data are not the sell prices that the vendors offer.
		// How do we know what the actual vendor offer will be?
		//
		// Is this a reputation issue? Do we scale the sell price down
		// based on how little reputation the seller has with this faction?
		return
	}
	toBuy = findArbitrages(auctions, accessToken)
	printShoppingList("Arbitrages", toBuy)
}

// findNewAuctions loops forever, printing any new auction data it finds
func findNewAuctions(accessToken string) {
	lastHash := int64(-1)

	for {
		now := time.Now()
		c, hash, ok := getCommodities(accessToken)
		if !ok {
			continue
		}
		if hash == lastHash {
			// We have already seen this auction data
			time.Sleep(1 * time.Minute)
			continue
		}
		lastHash = hash

		fmt.Printf("\n\n\n*** NEW AUCTION DATA (hh:%d) ***\n\n", now.Minute())

		printBargains(c, accessToken, true)

		a, hash, ok := getAuctions(accessToken)
		if !ok {
			continue
		}
		printBargains(a, accessToken, false)
	}
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

	findNewAuctions(accessToken)
}
