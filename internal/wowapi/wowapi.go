package wowapi

import (
	"fmt"
	"os"
	"strings"

	"github.com/erikbryant/web"
)

// connectedRealmIDCache the calls to find a connected realm ID are slow, cache the responses here
var connectedRealmIDCache = map[string]string{
	"Aegwynn":           "1136",
	"Agamaggan":         "1129",
	"Aggramar":          "106",
	"Akama":             "84",
	"Alexstrasza":       "1070",
	"Alleria":           "52",
	"Altar of Storms":   "78",
	"Alterac Mountains": "71",
	"Andorhal":          "96",
	"Anub'arak":         "1138",
	"Argent Dawn":       "75",
	"Azgalor":           "77",
	"Azjol-Nerub":       "121",
	"Azuremyst":         "160",
	"Baelgun":           "1190",
	"Blackhand":         "54",
	"Blackwing Lair":    "154",
	"Bloodhoof":         "64",
	"Bloodscalp":        "1185",
	"Bronzebeard":       "117",
	"Cairne":            "1168",
	"Coilfang":          "157",
	"Darrowmere":        "113",
	"Deathwing":         "155",
	"Dentarg":           "55",
	"Draenor":           "115",
	"Dragonblight":      "114",
	"Drak'thul":         "86",
	"Durotan":           "63",
	"Eitrigg":           "47",
	"Elune":             "67",
	"Eredar":            "53",
	"Farstriders":       "12",
	"Feathermoon":       "118",
	"Frostwolf":         "127",
	"Ghostlands":        "1175",
	"Greymane":          "158",
	"Icecrown":          "104",
	"Kilrogg":           "4",
	"Kirin Tor":         "1071",
	"Kul Tiras":         "1147",
	"Lightninghoof":     "163",
	"Llane":             "99",
	"Misha":             "1151",
	"Nazgrel":           "1184",
	"Ravencrest":        "1072",
	"Runetotem":         "151",
	"Sisters of Elune":  "125",

	// Remote realms: Oceanic
	"Aman'Thul":   "3726",
	"Barthilas":   "3723",
	"Caelestrasz": "3721",
	"Dath'Remar":  "3726",
	"Dreadmaul":   "3725",
	"Frostmourne": "3725",
	"Gundrak":     "3725",
	"Jubei'Thos":  "3725",
	"Khaz'goroth": "3726",
	"Nagrand":     "3721",
	"Saurfang":    "3721",
	"Thaurissan":  "3725",

	// Remote realms: Brazil
	"Azralon":   "3209",
	"Gallywix":  "3234",
	"Goldrinn":  "3207",
	"Nemesis":   "3208",
	"Tol Barad": "3208",

	// Remote realms: Latin America
	"Drakkari":    "1425",
	"Quel'Thalas": "1428",
	"Ragnaros":    "1427",
}

// realmToSlug returns the slug form of a given realm name
func realmToSlug(realm string) string {
	slug := strings.ToLower(realm)
	slug = strings.ReplaceAll(slug, "-", "")
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}

func request(url, token, caller string) (any, bool) {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	response, err := web.RequestJSON(url, headers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: no data returned: %s\n", caller, err)
		return nil, false
	}

	return response, true
}

func requestKey(url, token, key, caller string) ([]any, bool) {
	r, ok := request(url, token, caller)
	if !ok {
		return nil, false
	}
	response := r.(map[string]any)
	return response[key].([]any), true
}

// ConnectedRealm returns all realms connected to the given realm ID
func ConnectedRealm(realmID string) map[string]any {
	url := "https://us.api.blizzard.com/data/wow/connected-realm/" + realmID + "?namespace=dynamic-us&locale=en_US"
	r, ok := request(url, accessToken, "ConnectedRealm")
	if !ok {
		return nil
	}

	response := r.(map[string]any)
	if response["code"] != nil {
		fmt.Fprintf(os.Stderr, "ConnectedRealm failed to get connected realm: %v\n", response)
		return nil
	}

	return response
}

// ConnectedRealmSearch returns the set of all connected realms
func ConnectedRealmSearch() map[string]any {
	url := "https://us.api.blizzard.com/data/wow/search/connected-realm?namespace=dynamic-us&status.type=UP"
	r, ok := request(url, accessToken, "ConnectedRealm")
	if !ok {
		return nil
	}

	response := r.(map[string]any)
	if response["code"] != nil {
		fmt.Fprintf(os.Stderr, "ConnectedRealmSearch failed to get connected realms: %v\n", response)
		return nil
	}

	return response
}

// ConnectedRealmID returns the connected realm ID of the given realm
func ConnectedRealmID(realm string) (string, bool) {
	id, ok := connectedRealmIDCache[realm]
	if ok {
		return id, true
	}

	connectedRealms := ConnectedRealmSearch()
	if connectedRealms == nil {
		return "", false
	}

	slug := realmToSlug(realm)

	results := connectedRealms["results"].([]any)
	for _, result := range results {
		r := result.(map[string]any)
		data := r["data"].(map[string]any)
		cRealmID := web.ToString(data["id"])
		cr := ConnectedRealm(cRealmID)
		if cr == nil {
			continue
		}
		realms := cr["realms"].([]any)
		for _, cRealm := range realms {
			realmSlug := cRealm.(map[string]any)["slug"].(string)
			if slug == realmSlug {
				return cRealmID, true
			}
		}
	}

	fmt.Fprintf(os.Stderr, "ConnectedRealmID failed to find realm: %s\n", realm)
	return "", false
}

// Auctions returns the current auctions from the auction house
func Auctions(realm string) ([]any, bool) {
	connectedRealmID, ok := ConnectedRealmID(realm)
	if !ok {
		fmt.Fprintln(os.Stderr, "Auctions: no connected realm ID found")
		return nil, false
	}

	url := "https://us.api.blizzard.com/data/wow/connected-realm/" + connectedRealmID + "/auctions?namespace=dynamic-us&locale=en_US"
	r, ok := request(url, accessToken, "Auctions")
	if !ok {
		return nil, false
	}

	response := r.(map[string]any)
	if response["code"] != nil {
		fmt.Fprintf(os.Stderr, "Auctions: HTTP error: %v\n", response)
		return nil, false
	}

	auctions := response["auctions"].([]any)
	return auctions, true
}

// Commodities returns the current commodity auctions from the auction house
func Commodities() ([]any, bool) {
	url := "https://us.api.blizzard.com/data/wow/auctions/commodities?namespace=dynamic-us&locale=en_US"
	return requestKey(url, accessToken, "auctions", "Commodities")
}

// Item retrieves a single item from the WoW web API
func Item(id string) (map[string]any, bool) {
	url := "https://us.api.blizzard.com/data/wow/item/" + id + "?namespace=static-us&locale=en_US"
	r, ok := request(url, accessToken, "Item")
	if !ok {
		return nil, false
	}

	response := r.(map[string]any)
	if response["status"] == "nok" {
		fmt.Fprintf(os.Stderr, "Item: bad status for itemID %s: %v\n", id, response["reason"])
		return nil, false
	}
	_, ok = response["code"]
	if ok {
		fmt.Fprintf(os.Stderr, "Item: error retrieving itemID %s: %v\n", id, response)
		return nil, false
	}

	return response, true
}

// Pets returns a list of all battle pets in the game
func Pets() ([]any, bool) {
	url := "https://us.api.blizzard.com/data/wow/pet/index?namespace=static-us&locale=en_US"
	return requestKey(url, profileAccessToken, "pets", "Pets")
}

// CollectionsPets returns the battle pets the user owns
func CollectionsPets() ([]any, bool) {
	url := "https://us.api.blizzard.com/profile/user/wow/collections/pets?namespace=profile-us&locale=en_US"
	return requestKey(url, profileAccessToken, "pets", "CollectionsPets")
}

// Toys returns a list of all toys in the game
func Toys() ([]any, bool) {
	url := "https://us.api.blizzard.com/data/wow/toy/index?namespace=static-us&locale=en_US"
	return requestKey(url, profileAccessToken, "toys", "Toys")
}

// CollectionsToys returns the toys the user owns
func CollectionsToys() ([]any, bool) {
	url := "https://us.api.blizzard.com/profile/user/wow/collections/toys?namespace=profile-us&locale=en_US"
	return requestKey(url, profileAccessToken, "toys", "CollectionsToys")
}

// ItemAppearanceSetsIndex returns IDs of each appearance set
func ItemAppearanceSetsIndex() ([]any, bool) {
	url := "https://us.api.blizzard.com/data/wow/item-appearance/set/index?namespace=static-us&locale=en_US"
	return requestKey(url, accessToken, "appearance_sets", "ItemAppearanceSetsIndex")
}

// ItemAppearanceSetsIndexIDs returns the ID and name of each appearance set
func ItemAppearanceSetsIndexIDs() map[int64]string {
	index, ok := ItemAppearanceSetsIndex()
	if !ok {
		return nil
	}

	indexMap := map[int64]string{}
	for _, i := range index {
		i := i.(map[string]any)
		id := web.ToInt64(i["id"])
		name := web.ToString(i["name"])
		indexMap[id] = name
	}

	return indexMap
}

// ItemAppearanceSet returns the appearance IDs of the given appearance set
func ItemAppearanceSet(appearanceID int64) ([]any, bool) {
	url := fmt.Sprintf("https://us.api.blizzard.com/data/wow/item-appearance/set/%d?namespace=static-us&locale=en_US", appearanceID)
	return requestKey(url, accessToken, "appearances", "ItemAppearanceSet")
}

// ItemAppearanceSetIDs returns the appearance IDs that comprise the given appearance set
func ItemAppearanceSetIDs(appearanceID int64) []int64 {
	itemSet, ok := ItemAppearanceSet(appearanceID)
	if !ok {
		return nil
	}

	appearanceIDs := []int64{}
	for _, i := range itemSet {
		i := i.(map[string]any)
		appearanceIDs = append(appearanceIDs, web.ToInt64(i["id"]))
	}

	return appearanceIDs
}

// CollectionsTransmogs returns the transmogs the user owns
func CollectionsTransmogs() (any, bool) {
	url := "https://us.api.blizzard.com/profile/user/wow/collections/transmogs?namespace=profile-us&locale=en_US"
	return request(url, profileAccessToken, "CollectionsTransmogs")
}

// Professions returns the professions this alt knows
func Professions(realm, alt string) (any, bool) {
	realm = strings.ToLower(realm)
	realm = realmToSlug(realm)
	alt = strings.ToLower(alt)
	url := "https://us.api.blizzard.com/profile/wow/character/" + realm + "/" + alt + "/professions?namespace=profile-us&locale=en_US"
	return request(url, profileAccessToken, "Professions")
}
