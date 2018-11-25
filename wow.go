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

var (
	clientID     = flag.String("clientID", "", "WoW API client ID")
	clientSecret = flag.String("clientSecret", "", "WoW API client secret")
	realm        = flag.String("realm", "icecrown", "WoW realm")
)

// accessToken() retrieves an access token from battle.net. This token is used to authenticate API calls.
func accessToken(id, secret string) string {
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

// letsGoShopping() alerts if it finds bargain prices on goods I need.
func letsGoShopping(auctions []interface{}) {
	var goods = []struct {
		item     int64
		maxPrice int64
		name     string
	}{
		{33447, 28000, "Runic Healing Potion"},
		{34721, 28000, "Frostweave Bandage"},
		{34722, 40000, "Heavy Frostweave Bandage"},

		// Enchanting
		{34054, 20000, "Infinite Dust"},
		{34056, 25000, "Lesser Cosmic Essence"},
	}

	for _, a := range auctions {
		auction := a.(map[string]interface{})
		id := web.ToInt64(auction["item"])
		quantity := web.ToInt64(auction["quantity"])
		buyout := web.ToInt64(auction["buyout"])
		unitBuyout := buyout / quantity

		for _, good := range goods {
			if id == good.item && unitBuyout < good.maxPrice {
				discount := good.maxPrice - unitBuyout
				fmt.Printf("Shop  '%s' %s %d %d\n", good.name, auction["owner"], quantity, discount)
			}
		}
	}
}

func webGetItem(id, accessToken string) map[string]interface{} {
	url := "https://us.api.blizzard.com/wow/item/" + id + "?locale=en_US&access_token=" + accessToken
	return web.RequestJSON(url)
}

func lookupItem(id int64, accessToken string) database.Item {
	var item database.Item
	var ok bool

	// Do we have it cached in the database?
	item, ok = database.LookupItem(id)
	if ok {
		return item
	}

	i := webGetItem(web.ToString(id), accessToken)
	_, ok = i["sellPrice"]
	if !ok {
		fmt.Println("Item had no sellPrice")
		return item
	}
	item.Id = web.ToInt64(i["id"])
	item.Name = i["name"].(string)
	item.SellPrice = web.ToInt64(i["sellPrice"])
	item.JSON = fmt.Sprintf("%v", i)

	database.SaveItem(item)

	return item
}

// arbitrage() alerts if it finds auction prices that are lower than vendor prices.
func arbitrage(auctions []interface{}, accessToken string) {
	blacklist := map[int64]bool{
		71359:  true, // Chelley's Sterilized Scalpel
		71366:  true, // Lava Bolt Crossbow
		111475: true, // Beater's Beat Stick
	}

	for _, a := range auctions {
		auction := a.(map[string]interface{})
		id := web.ToInt64(auction["item"])
		if blacklist[id] {
			continue
		}
		item := lookupItem(id, accessToken)
		sellPrice := item.SellPrice
		bid := web.ToInt64(auction["bid"])
		quantity := web.ToInt64(auction["quantity"])
		buyout := web.ToInt64(auction["buyout"])
		unitBid := bid / quantity
		unitBuyout := buyout / quantity
		if unitBuyout > 0 && unitBuyout < sellPrice {
			profit := sellPrice*quantity - buyout
			if profit < 100 {
				continue
			}
			fmt.Printf("Buy!  '%s' %s %d %d\n", item.Name, auction["owner"], quantity, profit)
		} else {
			if unitBid < sellPrice {
				profit := sellPrice*quantity - bid
				if profit < 500 {
					continue
				}
				fmt.Printf("Bid   '%s' %s %d %d\n", item.Name, auction["owner"], quantity, profit)
			}
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
		token := accessToken(*clientID, *clientSecret)

		auctionURL := getAuctionURL(*realm, token)
		if auctionURL == lastAuctionURL {
			fmt.Println("No new AH data ...")
			time.Sleep(60 * time.Second)
			retries++
			continue
		}
		fmt.Println("retries:", retries)
		retries = 0
		lastAuctionURL = auctionURL

		auctions := getAuctions(auctionURL)

		fmt.Println()
		letsGoShopping(auctions)
		fmt.Println()
		arbitrage(auctions, token)
	}
}
