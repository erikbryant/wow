package wowAPI

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/erikbryant/web"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// Item contains the properties of a single auction house item
type Item struct {
	// WARNING: Changing this struct invalidates the cache
	Id         int64
	Name       string
	Equippable bool
	SellPrice  int64
}

// Sample auction response. Some have more or fewer fields.
// map[buyout:1.1111011e+09 id:3.49632108e+08 item:map[id:142075] quantity:1 time_left:VERY_LONG]

// Commodity auction response. All have exactly these fields.
// map[id:3.44371058e+08 item:map[id:192672] quantity:1 time_left:SHORT unit_price:16800]

// Auction contains the properties of a single auction house auction
type Auction struct {
	Id       int64
	ItemId   int64
	Buyout   int64 // For commodity auctions this stores 'unit_price'
	Quantity int64
}

var (
	cache     = map[int64]Item{}
	cacheFile = "cache.gob"
)

func init() {
	cacheLoad()
	fmt.Printf("#Cache items: %d\n", len(cache))
}

// realmToSlug returns the slug form of a given realm name
func realmToSlug(realm string) string {
	slug := strings.ToLower(realm)
	slug = strings.ReplaceAll(slug, "-", "")
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}

// AccessToken retrieves an access token from battle.net. This token is used to authenticate API calls.
func AccessToken(id, secret string) (string, bool) {
	grantString := "grant_type=client_credentials"
	request, err := http.NewRequest("POST", "https://oauth.battle.net/token", bytes.NewBuffer([]byte(grantString)))
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(id, secret)

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return "", false
	}

	var jsonObject map[string]interface{}

	err = json.Unmarshal(contents, &jsonObject)
	if err != nil {
		return "", false
	}

	return jsonObject["access_token"].(string), true
}

// Realm returns info about the given realm
func Realm(realm, accessToken string) map[string]interface{} {
	url := "https://us.api.blizzard.com/data/wow/realm/" + realmToSlug(realm) + "?namespace=dynamic-us&locale=en_US&access_token=" + accessToken

	response, err := web.RequestJSON(url, map[string]string{})
	if err != nil {
		fmt.Println("Realm: error getting realm", err)
		return nil
	}

	return response
}

// ConnectedRealm returns all realms connected to the given realm Id
func ConnectedRealm(realmId, accessToken string) map[string]interface{} {
	url := "https://us.api.blizzard.com/data/wow/connected-realm/" + realmId + "?namespace=dynamic-us&locale=en_US&access_token=" + accessToken

	response, err := web.RequestJSON(url, map[string]string{})
	if err != nil {
		fmt.Println("ConnectedRealm: Error getting connected realm", err)
		return nil
	}
	if response["code"] != nil {
		fmt.Println("ConnectedRealm: Failed to get connected realm", response)
		return nil
	}

	return response
}

// ConnectedRealmSearch returns the set of all connected realms
func ConnectedRealmSearch(accessToken string) map[string]interface{} {
	url := "https://us.api.blizzard.com/data/wow/search/connected-realm?namespace=dynamic-us&status.type=UP&access_token=" + accessToken
	response, err := web.RequestJSON(url, map[string]string{})
	if err != nil {
		fmt.Println("ConnectedRealmSearch: Error getting connected realms", err)
		return nil
	}
	if response["code"] != nil {
		fmt.Println("ConnectedRealmSearch: Failed to get connected realms", response)
		return nil
	}

	return response
}

// ConnectedRealmId returns the connected realm Id of the given realm
func ConnectedRealmId(realm, accessToken string) (string, bool) {
	connectedRealms := ConnectedRealmSearch(accessToken)
	if connectedRealms == nil {
		return "", false
	}

	slug := realmToSlug(realm)

	results := connectedRealms["results"].([]interface{})
	for _, result := range results {
		r := result.(map[string]interface{})
		data := r["data"].(map[string]interface{})
		cRealmId := web.ToString(data["id"])
		cr := ConnectedRealm(cRealmId, accessToken)
		if cr == nil {
			return "", false
		}
		realms := cr["realms"].([]interface{})
		for _, realm := range realms {
			realmSlug := realm.(map[string]interface{})["slug"].(string)
			if slug == realmSlug {
				return cRealmId, true
			}
		}
	}

	fmt.Println("ConnectedRealmId: Failed to find realm:", realm)
	return "", false
}

// Auctions returns the current auctions from the auction house
func Auctions(realm, accessToken string) ([]interface{}, bool) {
	connectedRealmId, ok := ConnectedRealmId(realm, accessToken)
	if !ok {
		fmt.Println("Auctions: no connected realm id found")
		return nil, false
	}

	url := "https://us.api.blizzard.com/data/wow/connected-realm/" + connectedRealmId + "/auctions?namespace=dynamic-us&locale=en_US&access_token=" + accessToken
	response, err := web.RequestJSON(url, map[string]string{})
	if err != nil {
		fmt.Println("Auctions: no auction data returned", err)
		return nil, false
	}

	if response["code"] != nil {
		fmt.Println("Auctions: HTTP error", response)
		return nil, false
	}

	auctions := response["auctions"].([]interface{})
	return auctions, true
}

// Commodities returns the current commodity auctions from the auction house
func Commodities(accessToken string) ([]interface{}, bool) {
	url := "https://us.api.blizzard.com/data/wow/auctions/commodities?namespace=dynamic-us&locale=en_US&access_token=" + accessToken
	response, err := web.RequestJSON(url, map[string]string{})
	if err != nil {
		fmt.Println("Commodities: no auction data returned", err)
		return nil, false
	}

	return response["auctions"].([]interface{}), true
}

// wowItem retrieves a single item from the WoW web API
func wowItem(id, accessToken string) (map[string]interface{}, bool) {
	url := "https://us.api.blizzard.com/data/wow/item/" + id + "?namespace=static-us&locale=en_US&access_token=" + accessToken
	response, err := web.RequestJSON(url, map[string]string{})
	if err != nil {
		fmt.Println("ItemId: failed to retrieve item", err)
		return nil, false
	}
	if response["status"] == "nok" {
		fmt.Println("INFO: ", response["reason"], "id: ", id)
		return nil, false
	}
	_, ok := response["code"]
	if ok {
		fmt.Println("Error retrieving id: ", id, response)
		return nil, false
	}

	return response, true
}

func cacheLoad() {
	file, err := os.Open(cacheFile)
	if err != nil {
		fmt.Printf("error opening cache file: %v", err)
		panic(err)
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&cache)
	if err != nil {
		fmt.Printf("error reading cache: %v", err)
		panic(err)
	}
}

func cacheSave() {
	file, err := os.Create(cacheFile)
	if err != nil {
		fmt.Printf("error creating cache file: %v", err)
		panic(err)
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	encoder.Encode(cache)
}

func cacheRead(id int64) (Item, bool) {
	item, ok := cache[id]
	return item, ok
}

func cacheWrite(id int64, item Item) {
	cache[id] = item
	cacheSave()
}

// LookupItem retrieves the data for a single item. It retrieves from the database if it is there, or the web if it is not. If it retrieves it from the web it also caches it.
func LookupItem(id int64, accessToken string) (Item, bool) {
	// Is it cached?
	item, ok := cacheRead(id)
	if ok {
		return item, true
	}

	i, ok := wowItem(web.ToString(id), accessToken)
	if !ok {
		return item, false
	}

	_, ok = i["name"]
	if !ok {
		fmt.Println("ItemId had no name:", id, i)
		return item, false
	}
	item.Name = i["name"].(string)

	item.Id = web.ToInt64(i["id"])

	item.Equippable = i["is_equippable"].(bool)

	_, ok = i["sell_price"]
	item.SellPrice = web.ToInt64(i["sell_price"])

	cacheWrite(id, item)

	return item, true
}
