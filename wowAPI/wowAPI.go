package wowAPI

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/erikbryant/web"
	"github.com/erikbryant/wowdb"
	"io"
	"log"
	"net/http"
	"strings"
)

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

// Item retrieves a single item from the WoW web API
func Item(id, accessToken string) (map[string]interface{}, bool) {
	url := "https://us.api.blizzard.com/data/wow/item/" + id + "?namespace=static-us&locale=en_US&access_token=" + accessToken
	response, err := web.RequestJSON(url, map[string]string{})
	if err != nil {
		fmt.Println("Item: failed to retrieve item", err)
		return nil, false
	}
	if response["status"] == "nok" {
		fmt.Println("INFO: ", response["reason"], "id: ", id)
		return nil, false
	}

	return response, true
}

// LookupItem retrieves the data for a single item. It retrieves from the database if it is there, or the web if it is not. If it retrieves it from the web it also caches it in the wowdb.
func LookupItem(id int64, accessToken string) (wowdb.Item, bool) {
	cache := true

	// Is it cached in the database?
	item, ok := wowdb.LookupItem(id)
	if ok {
		return item, true
	}

	i, ok := Item(web.ToString(id), accessToken)
	if !ok {
		return item, false
	}

	item.ID = web.ToInt64(i["id"])
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
		wowdb.SaveItem(item)
	}

	return item, true
}
