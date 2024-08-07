package wowAPI

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/erikbryant/aes"
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/cache"
	"github.com/erikbryant/wow/common"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	clientIDCrypt     = "f7FhewxUd0lWQz/zPb27ZcwI/ZqkaMyd5YyuskFyEugQEeiKsfL7dvr11Kx1Y+Mi23qMciOAPe5ksCOy"
	clientSecretCrypt = "CtJH62iU6V3ZeqiHyKItECHahdUYgAFyfHmQ4DRabhWIv6JeK5K4dT7aiybot6MS4JitmDzuWSz1UHHv"
	clientID          string
	clientSecret      string

	skipItems = map[int64]bool{
		// Items not found in the WoW database
		12034:  true,
		25308:  true,
		38517:  true,
		54629:  true,
		56054:  true,
		56055:  true,
		56056:  true,
		60390:  true,
		60405:  true,
		60406:  true,
		62370:  true,
		62770:  true,
		123865: true,
		123869: true,
		158078: true,
		159217: true,
		178149: true,
		178150: true,
		201420: true,
		201421: true,
		203932: true,
		204834: true,
		204835: true,
		204836: true,
		204837: true,
		204838: true,
		204839: true,
		204840: true,
		204841: true,
		204842: true,
	}
)

// SkipItem returns true if the caller should ignore this item
func SkipItem(item int64) bool {
	return skipItems[item]
}

// realmToSlug returns the slug form of a given realm name
func realmToSlug(realm string) string {
	slug := strings.ToLower(realm)
	slug = strings.ReplaceAll(slug, "-", "")
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}

// AccessToken retrieves an access token from battle.net. This token is used to authenticate API calls.
func AccessToken(passPhrase string) (string, bool) {
	clientID, err := aes.Decrypt(clientIDCrypt, passPhrase)
	if err != nil {
		log.Fatal(err)
	}

	clientSecret, err = aes.Decrypt(clientSecretCrypt, passPhrase)
	if err != nil {
		log.Fatal(err)
	}

	grantString := "grant_type=client_credentials"
	request, err := http.NewRequest("POST", "https://oauth.battle.net/token", bytes.NewBuffer([]byte(grantString)))
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(clientID, clientSecret)

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
		fmt.Println("Realm: error getting realm:", err)
		return nil
	}

	return response
}

// ConnectedRealm returns all realms connected to the given realm ID
func ConnectedRealm(realmId, accessToken string) map[string]interface{} {
	url := "https://us.api.blizzard.com/data/wow/connected-realm/" + realmId + "?namespace=dynamic-us&locale=en_US&access_token=" + accessToken

	response, err := web.RequestJSON(url, map[string]string{})
	if err != nil {
		fmt.Println("ConnectedRealm: Error getting connected realm:", err)
		return nil
	}
	if response["code"] != nil {
		fmt.Println("ConnectedRealm: Failed to get connected realm:", response)
		return nil
	}

	return response
}

// ConnectedRealmSearch returns the set of all connected realms
func ConnectedRealmSearch(accessToken string) map[string]interface{} {
	url := "https://us.api.blizzard.com/data/wow/search/connected-realm?namespace=dynamic-us&status.type=UP&access_token=" + accessToken
	response, err := web.RequestJSON(url, map[string]string{})
	if err != nil {
		fmt.Println("ConnectedRealmSearch: Error getting connected realms:", err)
		return nil
	}
	if response["code"] != nil {
		fmt.Println("ConnectedRealmSearch: Failed to get connected realms:", response)
		return nil
	}

	return response
}

// ConnectedRealmId returns the connected realm ID of the given realm
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
		fmt.Println("Auctions: no auction data returned:", err)
		return nil, false
	}

	if response["code"] != nil {
		fmt.Println("Auctions: HTTP error:", response)
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
		fmt.Println("Commodities: no auction data returned:", err)
		return nil, false
	}

	return response["auctions"].([]interface{}), true
}

// wowItem retrieves a single item from the WoW web API
func wowItem(id, accessToken string) (map[string]interface{}, bool) {
	url := "https://us.api.blizzard.com/data/wow/item/" + id + "?namespace=static-us&locale=en_US&access_token=" + accessToken
	response, err := web.RequestJSON(url, map[string]string{})
	if err != nil {
		fmt.Println("ItemId: failed to retrieve item:", err)
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

func age(item common.Item) time.Duration {
	return time.Now().Sub(item.Updated)
}

// LookupItem retrieves the data for a single item. It retrieves from the database if it is there, or the web if it is not. If it retrieves it from the web it also caches it.
func LookupItem(id int64, accessToken string) (common.Item, bool) {
	// Use the cached value if we have it
	item, ok := cache.Read(id)
	if ok {
		if age(item) <= 7*24*time.Hour {
			// Value in cache is fresh; use that
			return item, true
		}
		fmt.Println("Refreshing stale item:", item)
	}

	i, ok := wowItem(web.ToString(id), accessToken)
	if !ok {
		return item, false
	}

	item.Id = web.ToInt64(i["id"])

	_, ok = i["name"]
	if !ok {
		fmt.Println("ItemId had no name:", id, i)
		return item, false
	}
	item.Name = i["name"].(string)

	item.Equippable = i["is_equippable"].(bool)

	switch item.Id {
	case 141284:
		// Nor'danil Ampoule
		item.Equippable = true
	case 141286:
		// Rite of the Val'kyr
		item.Equippable = true
	case 141292:
		// Crystallizing Mana
		item.Equippable = true
	case 141293:
		// Spellfire Oil
		item.Equippable = true
	case 194829:
		// Fated Fortune Card (can't be sold until read)
		item.SellPrice = 10000
	default:
		_, ok = i["sell_price"]
		item.SellPrice = web.ToInt64(i["sell_price"])
	}

	_, ok = i["preview_item"]
	if ok {
		previewItem := i["preview_item"].(map[string]interface{})
		_, ok = previewItem["level"]
		if ok {
			level := previewItem["level"].(map[string]interface{})
			_, ok = level["value"]
			if ok {
				item.ItemLevel = web.ToInt64(level["value"])
			}
		}
	}

	item.Updated = time.Now()

	cache.Write(id, item)

	return item, true
}
