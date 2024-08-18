package wowAPI

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/erikbryant/aes"
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/cache"
	"github.com/erikbryant/wow/item"
	"github.com/erikbryant/wow/secrets"
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
		123868: true,
		123869: true,
		147455: true,
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
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}

// ProfileAccessToken retrieves a profile access token. This token is used to authenticate user profile API calls.
func ProfileAccessToken(passPhrase string) (string, bool) {
	clientID, err := aes.Decrypt(clientIDCrypt, passPhrase)
	if err != nil {
		log.Fatal(err)
	}

	url := "https://oauth.battle.net/oauth/authorize?response_type=code&client_id=" + clientID + "&scope=wow.profile%20sc2.profile&redirect_uri=https://develop.battle.net/documentation/world-of-warcraft/profile-apis"

	response, err := web.RequestBody(url, nil)
	if err != nil {
		fmt.Println("ProfileAccessToken: Error getting token:", err)
		return "", false
	}

	// Replace this with code to unpack the profile access token
	response = secrets.Pat

	return response, true
}

// AccessToken retrieves an access token. This token is used to authenticate API calls.
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

func stale(i item.Item) bool {
	return time.Now().Sub(i.Updated()) > 7*24*time.Hour
}

// LookupItem retrieves the data for a single item. It retrieves from the database if it is there, or the web if it is not. If it retrieves it from the web it also caches it.
func LookupItem(id int64, accessToken string) (item.Item, bool) {
	// Use the cached value if exists and not stale
	i, ok := cache.Read(id)
	if ok {
		if !stale(i) {
			return i, true
		}
		fmt.Println("Refreshing stale item:", i.Format())
	}

	result, ok := wowItem(web.ToString(id), accessToken)
	if !ok {
		return item.Item{}, false
	}
	i = item.NewItem(result)
	cache.Write(i.Id(), i)

	return i, true
}

// Pets returns a list of all battle pets in the game
func Pets(accessToken string) ([]interface{}, bool) {
	url := "https://us.api.blizzard.com/data/wow/pet/index?namespace=static-us&locale=en_US&access_token=" + accessToken

	response, err := web.RequestJSON(url, map[string]string{})
	if err != nil {
		fmt.Println("Pets: no data returned:", err)
		return nil, false
	}

	return response["pets"].([]interface{}), true
}

// CollectionsPets returns the battle pets the user owns
func CollectionsPets(profileAccessToken string) ([]interface{}, bool) {
	url := "https://us.api.blizzard.com/profile/user/wow/collections/pets?namespace=profile-us&locale=en_US&access_token=" + profileAccessToken

	response, err := web.RequestJSON(url, map[string]string{})
	if err != nil {
		fmt.Println("CollectionsPets: no data returned:", err)
		return nil, false
	}

	return response["pets"].([]interface{}), true
}
