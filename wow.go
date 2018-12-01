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
func webGetAuctionURL(realm, accessToken string) (string, int64) {
	url := "https://us.api.blizzard.com/wow/auction/data/" + realm + "?locale=en_US&access_token=" + accessToken
	response := web.RequestJSON(url)
	data := response["files"].([]interface{})[0].(map[string]interface{})
	return web.ToString(data["url"]), web.ToInt64(data["lastModified"])
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
func webGetAllItems(auctions map[int64]database.Auction, accessToken string) map[int64]database.Item {
	var items = map[int64]database.Item{}

	for _, auction := range auctions {
		if _, ok := items[auction.Item]; ok {
			continue
		}
		items[auction.Item] = webLookupItem(auction.Item, accessToken)
	}

	return items
}

// bargains returns all of the auctions for which the given goods are below our desired prices.
func bargains(auctions map[int64]database.Auction, goods map[int64]int64) (toBid []int64, toBuy []int64) {
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

	return toBid, toBuy
}

// jsonToStruct converts a single auction json string into a struct that is much easier to work with.
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

// unpackAuction unpacks the []interface{} format we get from the web into a map of structs. Same data, but in a format that is much easier to work with.
func unpackAuctions(auctions []interface{}) map[int64]database.Auction {
	aucs := make(map[int64]database.Auction)

	for _, a := range auctions {
		s := jsonToStruct(a.(map[string]interface{}))
		aucs[s.Auc] = s
	}

	return aucs
}

// arbitrage finds auction prices that are lower than vendor prices.
func arbitrage(auctions map[int64]database.Auction, items map[int64]database.Item) (toBid []int64, toBuy []int64) {
	for _, auction := range auctions {
		item := items[auction.Item]

		var j interface{}
		json.Unmarshal([]byte(item.JSON), &j)
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

// printShoppingList prints a list of auctions the user should consider bidding/buying.
func printShoppingList(header string, toGet []int64, auctions map[int64]database.Auction, items map[int64]database.Item) {
	printed := false
	for _, b := range toGet {
		fmt.Printf("%s <%s> %s quantity: %d\n", header, items[auctions[b].Item].Name, auctions[b].Owner, auctions[b].Quantity)
		printed = true
	}
	if printed {
		fmt.Println()
	}
}

func main() {
	flag.Parse()

	database.Open()
	defer database.Close()

	lastAuctionURL := ""
	lastModified := int64(0)
	for {
		// Make sure our credentials are current.
		accessToken := webGetAccessToken(*clientID, *clientSecret)

		// Sleep until a new auction file is published.
		auctionURL, modified := webGetAuctionURL(*realm, accessToken)
		if auctionURL == lastAuctionURL && modified == lastModified {
			fmt.Printf(".")
			time.Sleep(60 * time.Second)
			continue
		}
		fmt.Println()
		lastAuctionURL = auctionURL
		lastModified = modified

		// Database stats are fun to see! :-)
		fmt.Printf("#Items: %d #Auctions: %d\n\n", database.CountItems(), database.CountAuctions())

		// Download the auction file and all items for sale.
		auctions := unpackAuctions(webGetAuctions(auctionURL))
		items := webGetAllItems(auctions, accessToken)

		var goods = map[int64]int64{
			// Health
			33447: 30000, // Runic Healing Potion
			34721: 28000, // Frostweave Bandage
			34722: 40000, // Heavy Frostweave Bandage

			// Enchanting: Boots
			34054: 25000, // Infinite Dust
			34056: 30000, // Lesser Cosmic Essence

			// Enchanting: Runed Copper Rod
			10940: 800, // Strange Dust
			10938: 800, // Lesser Magic Essence
		}

		// Look for bargains on items we need.
		toBid, toBuy := bargains(auctions, goods)
		printShoppingList("Bid ", toBid, auctions, items)
		printShoppingList("Buy!", toBuy, auctions, items)

		// Look for items listed lower than what vendors will pay for them.
		toBid, toBuy = arbitrage(auctions, items)
		printShoppingList("Bid ", toBid, auctions, items)
		printShoppingList("Buy!", toBuy, auctions, items)
	}
}
