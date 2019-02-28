package main

// https://develop.battle.net/documentation
// $ go get github.com/go-sql-driver/mysql

import (
	"github.com/erikbryant/database"
	"github.com/erikbryant/web"
	"encoding/json"
	"flag"
	"fmt"
	"sort"
	"time"
)

var (
	clientID     = flag.String("clientID", "", "WoW API client ID")
	clientSecret = flag.String("clientSecret", "", "WoW API client secret")
	realm        = flag.String("realm", "icecrown", "WoW realm")
)

// getAccessToken retrieves an access token from battle.net. This token is used to authenticate API calls.
func webGetAccessToken(id, secret string) (string, bool) {
	url := "https://us.battle.net/oauth/token?client_id=" + id + "&client_secret=" + secret + "&grant_type=client_credentials"
	response, err := web.RequestJSON(url)
	if err != nil {
		fmt.Println("webGetAccessToken:", err)
		return "", false
	}

	if response == nil {
		fmt.Println("webGetAccessToken: no response from battle.net")
		return "", false
	}

	return response["access_token"].(string), true
}

// getAuctionURL retrieves the URL for the latest auction house data.
func webGetAuctionURL(realm, accessToken string) (string, int64, bool) {
	url := "https://us.api.blizzard.com/wow/auction/data/" + realm + "?locale=en_US&access_token=" + accessToken
	response, err := web.RequestJSON(url)
	if err != nil {
		fmt.Println("webGetAuctionURL: no response from api.blizzard.com", err)
		return "", 0, false
	}

	data := response["files"].([]interface{})[0].(map[string]interface{})
	return web.ToString(data["url"]), web.ToInt64(data["lastModified"]), true
}

// getAuctions retrieves the latest auctions from the auction house.
func webGetAuctions(auctionURL string) ([]interface{}, bool) {
	response, err := web.RequestJSON(auctionURL)
	if err != nil {
		fmt.Println("webGetAuction: no auction data returned", err)
		return nil, false
	}

	auctions := response["auctions"].([]interface{})

	return auctions, true
}

// webGetItem retrieves a single item from the WoW web API.
func webGetItem(id, accessToken string) (map[string]interface{}, bool) {
	url := "https://us.api.blizzard.com/wow/item/" + id + "?locale=en_US&access_token=" + accessToken
	response, err := web.RequestJSON(url)
	if err != nil {
		fmt.Println("webGetItem: failed to retrieve item from blizzard.com", err)
		return nil, false
	}

	return response, true
}

// lookupItem retrieves the data for a single item. It retrieves from the database if it is there, or the web if it is not. If it retrieves it from the web it also stores it in the database.
func webLookupItem(id int64, accessToken string) (database.Item, bool) {
	cache := true

	// Is it cached in the database?
	item, ok := database.LookupItem(id)
	if ok {
		return item, true
	}

	i, ok := webGetItem(web.ToString(id), accessToken)
	if !ok {
		return item, false
	}

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

	return item, true
}

// getAllItems prefetches item data for every requested item. This is faster than querying for each item and each of its repeats. It also makes the tests simpler.
func webGetAllItems(auctions map[int64]database.Auction, accessToken string) map[int64]database.Item {
	var items = map[int64]database.Item{}

	for _, auction := range auctions {
		if _, ok := items[auction.Item]; ok {
			continue
		}
		items[auction.Item], _ = webLookupItem(auction.Item, accessToken)
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

	// Keep the results stable across calls.
	sort.Slice(toBid, func(i, j int) bool { return toBid[i] < toBid[j] })
	sort.Slice(toBuy, func(i, j int) bool { return toBuy[i] < toBuy[j] })

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

// coinsToString returns a human-readable, formatted version of the coin amount.
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

// printShoppingList prints a list of auctions the user should consider bidding/buying.
func printShoppingList(action string, toGet []int64, auctions map[int64]database.Auction, items map[int64]database.Item) {
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

func main() {
	flag.Parse()

	database.Open()
	defer database.Close()

	lastAuctionURL := ""
	lastModified := int64(0)
	for {
		// Make sure our credentials are current.
		accessToken, ok := webGetAccessToken(*clientID, *clientSecret)
		if !ok {
			fmt.Println("ERROR: Unable to obtain access token.")
			return
		}

		// Sleep until a new auction file is published.
		auctionURL, modified, ok := webGetAuctionURL(*realm, accessToken)
		if !ok {
			continue
		}
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
		response, ok := webGetAuctions(auctionURL)
		if !ok {
			continue
		}
		auctions := unpackAuctions(response)
		items := webGetAllItems(auctions, accessToken)

		var goods = map[int64]int64{
			// Health
			34722: 40000, // Heavy Frostweave Bandage

			// Enchanting: Boots
			34056: 30000, // Lesser Cosmic Essence

			// Enchanting: Runed Copper Rod
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
