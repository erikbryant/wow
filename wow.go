package main

// https://develop.battle.net/documentation

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/erikbryant/aes"
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/wowAPI"
	"github.com/erikbryant/wowdb"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"sort"
)

var (
	clientIDCrypt     = "f7FhewxUd0lWQz/zPb27ZcwI/ZqkaMyd5YyuskFyEugQEeiKsfL7dvr11Kx1Y+Mi23qMciOAPe5ksCOy"
	clientSecretCrypt = "CtJH62iU6V3ZeqiHyKItECHahdUYgAFyfHmQ4DRabhWIv6JeK5K4dT7aiybot6MS4JitmDzuWSz1UHHv"
	clientID          string
	clientSecret      string
	passPhrase        = flag.String("passPhrase", "", "Passphrase to unlock WOW API client ID/secret")
	realm             = flag.String("realm", "Sisters of Elune", "WoW realm")
)

// webGetAllItems prefetches item data for every requested item. This is faster than querying for each item and each of its repeats. It also makes the tests simpler.
func webGetAllItems(auctions map[int64]wowAPI.Auction, accessToken string) map[int64]wowAPI.Item {
	var items = map[int64]wowAPI.Item{}

	for _, auction := range auctions {
		if _, ok := items[auction.Item]; ok {
			continue
		}
		items[auction.Item], _ = wowAPI.LookupItem(auction.Item, accessToken)
	}

	return items
}

// realmId returns the ID of the given realm
func realmId(realm, accessToken string) (string, bool) {
	response := wowAPI.Realm(realm, accessToken)
	if response == nil {
		return "", false
	}
	return response["id"].(string), true
}

// bargains returns all auctions for which the given goods are below our desired prices
func bargains(auctions map[int64]wowAPI.Auction, goods map[int64]int64) (toBid []int64, toBuy []int64) {
	for _, auction := range auctions {
		if _, ok := goods[auction.Item]; !ok {
			// We do not need this item.
			continue
		}

		maxPrice := goods[auction.Item]
		if auction.Buyout < maxPrice*auction.Quantity {
			toBuy = append(toBuy, auction.Auc)
		} else {
			if auction.Bid < maxPrice*auction.Quantity {
				toBid = append(toBid, auction.Auc)
			}
		}
	}

	// Keep the results stable across calls.
	sort.Slice(toBid, func(i, j int) bool { return toBid[i] < toBid[j] })
	sort.Slice(toBuy, func(i, j int) bool { return toBuy[i] < toBuy[j] })

	return toBid, toBuy
}

// jsonToStruct converts a single auction json string into a struct that is much easier to work with
func jsonToStruct(auc map[string]interface{}) wowAPI.Auction {
	var auction wowAPI.Auction

	if _, ok := auc["buyout"]; ok {
		auction.Buyout = web.ToInt64(auc["buyout"])
	}
	if _, ok := auc["bid"]; ok {
		auction.Bid = web.ToInt64(auc["bid"])
	}
	auction.Quantity = web.ToInt64(auc["quantity"])
	auction.TimeLeft = web.ToString(auc["timeLeft"])
	auction.Auc = web.ToInt64(auc["id"])

	_, ok := auc["item"]
	if !ok {
		fmt.Println("Auction had no item: ", auc)
		return wowAPI.Auction{}
	}
	item := auc["item"].(map[string]interface{})
	auction.Item = web.ToInt64(item["id"])

	b, _ := json.Marshal(auc)
	auction.JSON = fmt.Sprintf("%s", b)

	return auction
}

// unpackAuction converts the []interface{} format we get from the web into a map of structs
func unpackAuctions(auctions []interface{}) map[int64]wowAPI.Auction {
	aucs := make(map[int64]wowAPI.Auction)

	for _, a := range auctions {
		s := jsonToStruct(a.(map[string]interface{}))
		aucs[s.Auc] = s
	}

	return aucs
}

// arbitrage finds auction prices that are lower than vendor prices
func arbitrage(auctions map[int64]wowAPI.Auction, items map[int64]wowdb.Item) (toBid []int64, toBuy []int64) {
	for _, auction := range auctions {
		item := items[auction.Item]

		var j interface{}
		json.Unmarshal([]byte(item.JSON), &j)
		if j == nil {
			continue
		}
		js := j.(map[string]interface{})
		if js["equippable"].(bool) {
			// I do not understand how to price these.
			continue
		}

		profit := item.SellPrice*auction.Quantity - auction.Buyout
		if auction.Buyout > 0 && profit >= 500 {
			toBuy = append(toBuy, auction.Auc)
			continue
		}

		profit = item.SellPrice*auction.Quantity - auction.Bid
		if profit >= 5000 {
			toBid = append(toBid, auction.Auc)
		}
	}

	return toBid, toBuy
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
func printShoppingList(action string, toGet []int64, auctions map[int64]wowAPI.Auction, items map[int64]wowdb.Item) {
	if len(toGet) == 0 {
		return
	}

	for _, b := range toGet {
		item := items[auctions[b].Item]
		auction := auctions[b]
		profitBid := item.SellPrice*auction.Quantity - auction.Bid
		profitBuy := item.SellPrice*auction.Quantity - auction.Buyout
		fmt.Printf("%s <%s> %s quantity: %d profit: %s/%s\n", action, item.Name, auction.Owner, auction.Quantity, coinsToString(profitBid), coinsToString(profitBuy))
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
		fmt.Println("ERROR: You must specify -passPhrase to unlock the client ID/secret")
		usage()
	}

	clientID, err := aes.Decrypt(clientIDCrypt, *passPhrase)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	clientSecret, err = aes.Decrypt(clientSecretCrypt, *passPhrase)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	wowdb.Open()
	defer wowdb.Close()

	accessToken, ok := wowAPI.AccessToken(clientID, clientSecret)
	if !ok {
		fmt.Println("ERROR: Unable to obtain access token.")
		return
	}

	a, ok := wowAPI.Auctions(*realm, accessToken)
	if !ok {
		fmt.Println("ERROR: Unable to obtain auctions.")
		return
	}
	auctions := unpackAuctions(a)
	fmt.Printf("#Auctions: %d\n\n", len(auctions))

	//c, ok := wowAPI.Commodities(accessToken)
	//if !ok {
	//	fmt.Println("ERROR: Unable to obtain commodity auctions.")
	//	return
	//}
	//auctionsC := unpackAuctions(c)
	//fmt.Printf("#Commodities: %d\n\n", len(auctionsC))

	for _, auction := range auctions {
		item, ok := wowAPI.LookupItem(auction.Item, accessToken)
		if !ok {
			return
		}
		if item.Equippable {
			// I do not understand how to price these
			continue
		}
		if auction.Buyout > 0 && auction.Buyout < item.SellPrice {
			fmt.Printf("Arbitrage! %d < %d\n%v\n%v\n\n", auction.Buyout, item.SellPrice, auction, item)
		}
	}

	//	var goods = map[int64]int64{
	//		// Health
	//		34722: 40000, // Heavy Frostweave Bandage
	//
	//		// Enchanting: Boots
	//		34056: 30000, // Lesser Cosmic Essence
	//
	//		// Enchanting: Runed Copper Rod
	//		10938: 800, // Lesser Magic Essence
	//	}
	//
	//	// Look for bargains on items we need.
	//	toBid, toBuy := bargains(auctions, goods)
	//	printShoppingList("Bid ", toBid, auctions, items)
	//	printShoppingList("Buy!", toBuy, auctions, items)
}
