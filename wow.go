package main

// https://develop.battle.net/documentation
// $ go get github.com/go-sql-driver/mysql

import (
	"./database"
	"./web"
	"encoding/json"
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

// ShoppingItem is a single item that the user can purchase.
type ShoppingItem struct {
	name     string
	owner    string
	quantity int64
	discount int64
}

var (
	clientID     = flag.String("clientID", "", "WoW API client ID")
	clientSecret = flag.String("clientSecret", "", "WoW API client secret")
	realm        = flag.String("realm", "icecrown", "WoW realm")
)

// getAccessToken retrieves an access token from battle.net. This token is used to authenticate API calls.
func webGetAccessToken(id, secret string) string {
	url := "https://us.battle.net/oauth/token?client_id=" + id + "&client_secret=" + secret + "&grant_type=client_credentials"
	response := web.RequestJSON(url)
	return response["access_token"].(string)
}

// getAuctionURL retrieves the URL for the latest auction house data.
func webGetAuctionURL(realm, accessToken string) string {
	url := "https://us.api.blizzard.com/wow/auction/data/" + realm + "?locale=en_US&access_token=" + accessToken
	response := web.RequestJSON(url)
	data := response["files"].([]interface{})[0].(map[string]interface{})
	return data["url"].(string)
}

// getAuctions retrieves the latest auctions from the auction house.
func webGetAuctions(auctionURL string) []interface{} {
	response := web.RequestJSON(auctionURL)
	auctions := response["auctions"].([]interface{})
	return auctions
}

// webGetItem retrieves a single item from the WoW web API.
func webGetItem(id, accessToken string) map[string]interface{} {
	url := "https://us.api.blizzard.com/wow/item/" + id + "?locale=en_US&access_token=" + accessToken
	return web.RequestJSON(url)
}

// lookupItem retrieves the data for a single item. It retrieves from the database if it is there, or the web if it is not. If it retrieves it from the web it also stores it in the database.
func webLookupItem(id int64, accessToken string) (item database.Item) {
	var ok bool
	cache := true

	// Do we have it cached in the database?
	item, ok = database.LookupItem(id)
	if ok {
		return
	}

	i := webGetItem(web.ToString(id), accessToken)
	item.Id = web.ToInt64(i["id"])
	b, _ := json.Marshal(i)
	item.JSON = fmt.Sprintf("%s", b)
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

// getAllItems prefetches item data for every requested item. This is faster than querying for each item and each of its repeats. It also makes the tests simpler.
func webGetAllItems(auctions []interface{}, accessToken string) map[int64]database.Item {
	var items = map[int64]database.Item{}

	for _, a := range auctions {
		auction := a.(map[string]interface{})
		id := web.ToInt64(auction["item"])
		if _, ok := items[id]; ok {
			continue
		}
		item := webLookupItem(id, accessToken)
		items[id] = item
	}

	return items
}

// itemsWeNeed returns all of the auctions for which the given goods are below our desired prices.
func itemsWeNeed(auctions []interface{}, goods []Good) map[int64][]ShoppingItem {
	shoppingList := make(map[int64][]ShoppingItem)

	for _, a := range auctions {
		auction := a.(map[string]interface{})
		id := web.ToInt64(auction["item"])
		quantity := web.ToInt64(auction["quantity"])
		buyout := web.ToInt64(auction["buyout"])
		unitBuyout := buyout / quantity

		for _, good := range goods {
			if id == good.item && unitBuyout < good.maxPrice {
				discount := good.maxPrice - unitBuyout
				if _, ok := shoppingList[id]; !ok {
					shoppingList[id] = []ShoppingItem{}
				}
				shoppingItem := ShoppingItem{
					name:     good.name,
					owner:    web.ToString(auction["owner"]),
					quantity: quantity,
					discount: discount,
				}
				shoppingList[id] = append(shoppingList[id], shoppingItem)
			}
		}
	}

	return shoppingList
}

func jsonToStruct(auc map[string]interface{}) database.Auction {
	var auction database.Auction
	var ok bool

	auction.Auc = web.ToInt64(auc["auc"])
	auction.Item = web.ToInt64(auc["item"])
	auction.Owner = web.ToString(auc["owner"])
	auction.Bid = web.ToInt64(auc["bid"])
	auction.Buyout = web.ToInt64(auc["buyout"])
	auction.Quantity = web.ToInt64(auc["quantity"])
	auction.TimeLeft = web.ToString(auc["timeLeft"])
	auction.Rand = web.ToInt64(auc["rand"])
	auction.Seed = web.ToInt64(auc["seed"])
	auction.Context = web.ToInt64(auc["context"])
	_, auction.HasBonusLists = auc["bonusLists"]
	_, auction.HasModifiers = auc["bonusModifiers"]
	if _, ok = auc["petBreedId"]; ok {
		auction.PetBreedId = web.ToInt64(auc["petBreedId"])
	}
	if _, ok = auc["petLevel"]; ok {
		auction.PetLevel = web.ToInt64(auc["petLevel"])
	}
	if _, ok = auc["petQualityId"]; ok {
		auction.PetQualityId = web.ToInt64(auc["petQualityId"])
	}
	if _, ok = auc["petSpeciesId"]; ok {
		auction.PetSpeciesId = web.ToInt64(auc["petSpeciesId"])
	}
	b, _ := json.Marshal(auc)
	auction.JSON = fmt.Sprintf("%s", b)

	database.SaveAuction(auction)

	return auction
}

func saveAuctions(auctions []interface{}) []database.Auction {
	var aucs []database.Auction

	for _, a := range auctions {
		s := jsonToStruct(a.(map[string]interface{}))
		aucs = append(aucs, s)
	}

	return aucs
}

// arbitrage flags auction prices that are lower than vendor prices.
func arbitrage(auctions []interface{}, items map[int64]database.Item) map[int64][]ShoppingItem {
	shoppingList := make(map[int64][]ShoppingItem)

	for _, a := range auctions {
		auction := a.(map[string]interface{})
		item := items[web.ToInt64(auction["item"])]
		var j interface{}
		json.Unmarshal([]byte(item.JSON), &j)
		js := j.(map[string]interface{})
		if js["equippable"].(bool) {
			// I do not understand how to price these.
			continue
		}
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

	return shoppingList
}

func main() {
	flag.Parse()

	database.Open()
	defer database.Close()

	lastAuctionURL := ""
	retries := 0
	for {
		accessToken := webGetAccessToken(*clientID, *clientSecret)

		auctionURL := webGetAuctionURL(*realm, accessToken)
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

		auctions := webGetAuctions(auctionURL)
		saveAuctions(auctions)
		items := webGetAllItems(auctions, accessToken)

		var goods = []Good{
			// Health
			{33447, 30000, "Runic Healing Potion"},
			{34721, 28000, "Frostweave Bandage"},
			{34722, 40000, "Heavy Frostweave Bandage"},

			// Enchanting: Boots
			{34054, 25000, "Infinite Dust"},
			{34056, 30000, "Lesser Cosmic Essence"},

			// Enchanting: Runed Copper Rod
			{10940, 800, "Strange Dust"},
			{10938, 800, "Lesser Magic Essence"},
		}

		shoppingList := itemsWeNeed(auctions, goods)
		for _, list := range shoppingList {
			for _, item := range list {
				fmt.Printf("Shop  '%s' %s quantity: %d save: %d\n", item.name, item.owner, item.quantity, item.discount*item.quantity)
			}
		}
		fmt.Println()

		shoppingList = arbitrage(auctions, items)
		for _, list := range shoppingList {
			for _, item := range list {
				fmt.Printf("Shop  '%s' %s quantity: %d save: %d\n", item.name, item.owner, item.quantity, item.discount*item.quantity)
			}
		}
		fmt.Println()
	}
}
