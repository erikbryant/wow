package main

// https://develop.battle.net/documentation
// $ go get github.com/go-sql-driver/mysql

import (
	"./database"
	"./web"
	"flag"
	"fmt"
	"time"
)

// Good represents an item that we could go shopping for.
type Good struct {
	item     int64
	maxPrice int64
	name     string
}

var (
	clientID     = flag.String("clientID", "", "WoW API client ID")
	clientSecret = flag.String("clientSecret", "", "WoW API client secret")
	realm        = flag.String("realm", "icecrown", "WoW realm")
)

// getAccessToken() retrieves an access token from battle.net. This token is used to authenticate API calls.
func getAccessToken(id, secret string) string {
	url := "https://us.battle.net/oauth/token?client_id=" + id + "&client_secret=" + secret + "&grant_type=client_credentials"
	response := web.RequestJSON(url)
	return response["access_token"].(string)
}

// getAuctionURL() retrieves the URL for the latest auction house data.
func getAuctionURL(realm, accessToken string) string {
	url := "https://us.api.blizzard.com/wow/auction/data/" + realm + "?locale=en_US&access_token=" + accessToken
	response := web.RequestJSON(url)
	data := response["files"].([]interface{})[0].(map[string]interface{})
	return data["url"].(string)
}

// getAuctions() retrieves the latest auctions from the auction house.
func getAuctions(auctionURL string) []interface{} {
	response := web.RequestJSON(auctionURL)
	auctions := response["auctions"].([]interface{})
	return auctions
}

// letsGoShopping() alerts if it finds bargain prices on a given list of goods.
func letsGoShopping(auctions []interface{}, goods []Good) {
	for _, a := range auctions {
		auction := a.(map[string]interface{})
		id := web.ToInt64(auction["item"])
		quantity := web.ToInt64(auction["quantity"])
		buyout := web.ToInt64(auction["buyout"])
		unitBuyout := buyout / quantity

		for _, good := range goods {
			if id == good.item && unitBuyout < good.maxPrice {
				discount := good.maxPrice - unitBuyout
				fmt.Printf("Shop  '%s' %s quantity: %d save: %d\n", good.name, auction["owner"], quantity, discount)
			}
		}
	}
}

func webGetItem(id, accessToken string) map[string]interface{} {
	url := "https://us.api.blizzard.com/wow/item/" + id + "?locale=en_US&access_token=" + accessToken
	return web.RequestJSON(url)
}

func lookupItem(id int64, accessToken string) (item database.Item) {
	var ok bool
	cache := true

	// Do we have it cached in the database?
	item, ok = database.LookupItem(id)
	if ok {
		return
	}

	i := webGetItem(web.ToString(id), accessToken)
	item.Id = web.ToInt64(i["id"])
	item.JSON = fmt.Sprintf("%v", i)
	_, ok = i["sellPrice"]
	if ok {
		item.SellPrice = web.ToInt64(i["sellPrice"])
	} else {
		fmt.Println("Item had no sellPrice:", i)
		cache = false
	}
	_, ok = i["name"]
	if ok {
		item.Name = i["name"].(string)
	} else {
		fmt.Println("Item had no name:", i)
		cache = false
	}

	// Cache it. Database lookups are much faster than web calls.
	if cache {
		database.SaveItem(item)
	}

	return
}

// getAuctionItems retrieves the item data for every item in the auction house data. This is faster than querying for each item and each of its repeats. It also makes the tests simpler.
func getAuctionItems(auctions []interface{}, accessToken string) map[int64]database.Item {
	var items = map[int64]database.Item{}

	for _, a := range auctions {
		auction := a.(map[string]interface{})
		id := web.ToInt64(auction["item"])
		if _, ok := items[id]; ok {
			continue
		}
		item := lookupItem(id, accessToken)
		items[id] = item
	}

	return items
}

// arbitrage() flags auction prices that are lower than vendor prices.
func arbitrage(auctions []interface{}, items map[int64]database.Item) {
	for _, a := range auctions {
		auction := a.(map[string]interface{})
		item := items[web.ToInt64(auction["item"])]
		bid := web.ToInt64(auction["bid"])
		quantity := web.ToInt64(auction["quantity"])
		buyout := web.ToInt64(auction["buyout"])

		profit := item.SellPrice*quantity - buyout
		if buyout > 0 && profit >= 500 {
			fmt.Printf("Buy!  '%s' %s %d %d\n", item.Name, auction["owner"], quantity, profit)
			continue
		}

		profit = item.SellPrice*quantity - bid
		if profit >= 5000 {
			fmt.Printf("Bid   '%s' %s %d %d\n", item.Name, auction["owner"], quantity, profit)
			continue
		}
	}
}

func main() {
	flag.Parse()

	database.Open()
	defer database.Close()

	lastAuctionURL := ""
	retries := 0
	for {
		accessToken := getAccessToken(*clientID, *clientSecret)

		auctionURL := getAuctionURL(*realm, accessToken)
		if auctionURL == lastAuctionURL {
			fmt.Printf(".")
			time.Sleep(60 * time.Second)
			retries++
			continue
		}
		fmt.Println("retries:", retries)
		fmt.Println()
		retries = 0
		lastAuctionURL = auctionURL

		auctions := getAuctions(auctionURL)
		items := getAuctionItems(auctions, accessToken)

		var goods = []Good{
			// Health
			{33447, 30000, "Runic Healing Potion"},
			{34721, 28000, "Frostweave Bandage"},
			{34722, 40000, "Heavy Frostweave Bandage"},

			// Enchanting
			{34054, 25000, "Infinite Dust"},
			{34056, 30000, "Lesser Cosmic Essence"},
		}

		letsGoShopping(auctions, goods)
		fmt.Println()

		arbitrage(auctions, items)
		fmt.Println()
	}
}
