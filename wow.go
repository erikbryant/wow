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
)

var (
	clientIDCrypt     = "f7FhewxUd0lWQz/zPb27ZcwI/ZqkaMyd5YyuskFyEugQEeiKsfL7dvr11Kx1Y+Mi23qMciOAPe5ksCOy"
	clientSecretCrypt = "CtJH62iU6V3ZeqiHyKItECHahdUYgAFyfHmQ4DRabhWIv6JeK5K4dT7aiybot6MS4JitmDzuWSz1UHHv"
	clientID          string
	clientSecret      string
	passPhrase        = flag.String("passPhrase", "", "Passphrase to unlock WOW API client ID/secret")
	realm             = flag.String("realm", "Sisters of Elune", "WoW realm")
)

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

// bargains returns auctions for which the goods are at or below our desired prices
func bargains(goods map[int64]int64, auctions map[int64]wowAPI.Auction) (toBuy []wowAPI.Auction) {
	for _, auction := range auctions {
		if _, ok := goods[auction.Item]; !ok {
			// We do not need this item
			continue
		}

		maxPrice := goods[auction.Item]
		if auction.Buyout <= maxPrice {
			toBuy = append(toBuy, auction)
		}
	}

	return toBuy
}

// arbitrage returns auctions selling for lower than vendor prices
func arbitrage(auctions map[int64]wowAPI.Auction, accessToken string) (toBuy []wowAPI.Auction) {
	for _, auction := range auctions {
		item, ok := wowAPI.LookupItem(auction.Item, accessToken)
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
		item, ok := wowAPI.LookupItem(auction.Item, accessToken)
		if !ok {
			fmt.Println("ERROR: Unable to lookup item for shopping list auction: ", auction)
			continue
		}
		profitBuy := item.SellPrice*auction.Quantity - auction.Buyout
		fmt.Printf("%s <%s> quantity: %d profit: %s\n", action, item.Name, auction.Quantity, coinsToString(profitBuy))
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

	c, ok := wowAPI.Commodities(accessToken)
	if !ok {
		fmt.Println("ERROR: Unable to obtain commodity auctions.")
		return
	}
	auctionsC := unpackAuctions(c)
	fmt.Printf("#Commodities: %d\n\n", len(auctionsC))

	// Look for bargains

	goods := map[int64]int64{
		// Health
		34722: 1, // Heavy Frostweave Bandage
		// Enchanting
		34057: 1, // Abyss Crystal
		34056: 1, // Lesser Cosmic Essence
	}

	toBuy := bargains(goods, auctions)
	toBuyC := bargains(goods, auctionsC)
	printShoppingList("Buy!", toBuy, accessToken)
	printShoppingList("Buy!", toBuyC, accessToken)

	//toBuy = append(toBuy, arbitrage(auctions, accessToken)...)
}
