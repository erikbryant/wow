package main

// https://develop.battle.net/documentation

import (
	"flag"
	"fmt"
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
	ItemLevel   int64
}

var (
	passPhrase  = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	realm       = flag.String("realm", "Sisters of Elune", "WoW realm")
	usefulGoods = map[int64]int64{
		// Health
		211943: 6000, // Scarlet Silk Bandage

		// Enchanting
		34057:  7900,    // Abyss Crystal
		7909:   2000,    // Aquamarine
		22445:  7000,    // Arcane Dust
		124440: 29800,   // Arkhana
		124442: 18700,   // Chaos Crystal
		109693: 5000,    // Draenic Dust
		3819:   97300,   // Dragon's Teeth
		9224:   670000,  // Elixir of Demonslaying
		7082:   170000,  // Essence of Air
		7076:   2000,    // Essence of Earth
		7078:   5000,    // Essence of Fire
		12808:  500000,  // Essence of Undeath
		7080:   458700,  // Essence of Water
		23427:  305000,  // Eternium Ore
		22794:  57500,   // Fel Lotus
		124116: 742400,  // Felhide
		124106: 997000,  // Felwort
		4625:   55000,   // Firebloom
		52719:  42500,   // Greater Celestial Essence
		16203:  89000,   // Greater Eternal Essence
		22446:  9700,    // Greater Planar Essence
		52721:  19900,   // Heavenly Shard
		52555:  42100,   // Hypnotic Dust
		124444: 70000,   // Infernal Brimstone
		34054:  1900,    // Infinite Dust
		14344:  1300,    // Large Brilliant Shard
		34056:  2400,    // Lesser Cosmic Essence
		16202:  8000,    // Lesser Eternal Essence
		10938:  4500,    // Lesser Magic Essence (720 in shop)
		22447:  4300,    // Lesser Planar Essence
		124441: 71500,   // Leylight Shard
		16204:  600,     // Light Illusion Dust
		12803:  233900,  // Living Essence
		111245: 30000,   // Luminous Shard
		52722:  10100,   // Maelstrom Crystal
		22791:  46700,   // Netherbloom
		22792:  39900,   // Nightmare Vine
		22451:  2750000, // Primal Air
		22452:  60000,   // Primal Earth
		21884:  1740000, // Primal Fire
		21886:  699800,  // Primal Life
		22457:  1156300, // Primal Mana
		23571:  6959900, // Primal Might
		22456:  1059800, // Primal Shadow
		21885:  883300,  // Primal Water
		156930: 3500,    // Rich Illusion Dust
		14343:  700,     // Small Brilliant Shard
		22448:  3000,    // Small Prismatic Shard
		10940:  1000,    // Strange Dust (720 in shop)
		113588: 58100,   // Temporal Crystal
		22450:  9700,    // Void Crystal
		52328:  150000,  // Volatile Air
		8153:   3500000, // WildVine

		// Tailoring
		111557: 8800, // Sumptuous Fur

		// Bags
		194017: 500000, // Wildercloth Bag

		// Item pricing research
		33428:  10000000, // Dulled Shiv
		201954: 10000000, // Explorer's Expert Greaves
		15212:  10000000, // Fighter Broadsword
		15248:  10000000, // Gleaming Claymore
		121110: 10000000, // Hagfeather Wristwraps
		2057:   10000000, // Pitted Defias Shortsword
		154778: 10000000, // Ruptured Plate Vambraces
		6563:   10000000, // Shimmering Bracers
		2215:   10000000, // Wooden Shield
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
			if auction.Buyout <= 0 {
				continue
			}
			if auction.Buyout < maxPrice {
				bargain := Bargain{
					Quantity:    auction.Quantity,
					UnitSavings: maxPrice - auction.Buyout,
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
func findArbitrages(auctions map[int64][]common.Auction, accessToken string) []Bargain {
	bargains := []Bargain{}

	for itemId, aucs := range auctions {
		item, ok := wowAPI.LookupItem(itemId, accessToken)
		if !ok {
			continue
		}
		for _, auction := range aucs {
			if auction.Buyout <= 0 {
				continue
			}
			if auction.Buyout < item.SellPrice {
				bargain := Bargain{
					Quantity:    auction.Quantity,
					UnitSavings: item.SellPrice - auction.Buyout,
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
	fmt.Printf("--- %s ---\n", label)

	lastName := ""
	for _, bargain := range bargains {
		if bargain.Name == lastName {
			// Only print an item once
			continue
		}
		if bargain.ItemLevel > 0 {
			// I don't know how to price these yet
			//fmt.Printf("%-50s %d\n", bargain.Name, bargain.ItemLevel)
			continue
		} else {
			fmt.Printf("%s\n", bargain.Name)
		}
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
func printBargains(auctions map[int64][]common.Auction, accessToken string) {
	toBuy := findBargains(usefulGoods, auctions, accessToken)
	printShoppingList(fmt.Sprintf("Bargains (%d)", len(auctions)), toBuy)
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

		fmt.Printf("\n\n\n*** New Auction House Data (hh:%02d) ***\n\n", now.Minute())

		printBargains(c, accessToken)

		a, hash, ok := getAuctions(accessToken)
		if !ok {
			continue
		}
		printBargains(a, accessToken)
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

	accessToken, ok := wowAPI.AccessToken(*passPhrase)
	if !ok {
		log.Fatal("ERROR: Unable to obtain access token.")
	}

	findNewAuctions(accessToken)
}
