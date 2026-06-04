package wowAPI

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/erikbryant/aes"
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/oauth2"
)

var (
	clientIDCrypt      = "f7FhewxUd0lWQz/zPb27ZcwI/ZqkaMyd5YyuskFyEugQEeiKsfL7dvr11Kx1Y+Mi23qMciOAPe5ksCOy"
	clientSecretCrypt  = "CtJH62iU6V3ZeqiHyKItECHahdUYgAFyfHmQ4DRabhWIv6JeK5K4dT7aiybot6MS4JitmDzuWSz1UHHv"
	clientID           string
	clientSecret       string
	accessToken        string
	profileAccessToken string

	skipItems = map[int64]struct{}{
		// Items not found in the WoW database
		23704:  {},
		23942:  {},
		23943:  {},
		23955:  {},
		23958:  {},
		23972:  {},
		29557:  {},
		29558:  {},
		29566:  {},
		42929:  {},
		43557:  {},
		54629:  {},
		56054:  {},
		56055:  {},
		56056:  {},
		60390:  {},
		60405:  {},
		60406:  {},
		62370:  {},
		62770:  {},
		123865: {},
		123868: {},
		123869: {},
		147455: {},
		178149: {},
		198485: {},
		201420: {},
		201421: {},
		203932: {},
		204834: {},
		204835: {},
		204836: {},
		204837: {},
		204838: {},
		204839: {},
		204840: {},
		204841: {},
		204842: {},
		212531: {},
		212533: {},
		212534: {},
		213234: {},
		213235: {},
		213236: {},
		213237: {},
		213238: {},
		213239: {},
		213240: {},
		213241: {},
		213242: {},
		213245: {},
		213246: {},
		213247: {},
		213248: {},
		213249: {},
		213250: {},
		213251: {},
		213252: {},
		213253: {},
		213254: {},
		213255: {},
		213256: {},
		213257: {},
		213258: {},
		213259: {},
		213260: {},
		213261: {},
		213262: {},
		213263: {},
		213265: {},
		213266: {},
		213267: {},
		213268: {},
		217387: {},
		217958: {},
		217959: {},
		217962: {},
		217969: {},
		222906: {},
		224153: {},
		224154: {},
		224155: {},
		225218: {},
		225219: {},
		225236: {},
		225237: {},
		225254: {},
		225784: {},
		225787: {},
		225839: {},
		225840: {},
		226001: {},
		226002: {},
		226003: {},
		226004: {},
		226005: {},
		228386: {},
		244052: {},
		246040: {},
		262792: {},
		262793: {},
		262794: {},
		262795: {},
		262796: {},
		262797: {},
		262798: {},
		262799: {},
		262800: {},
		268944: {},
		268945: {},
		268946: {},
		268947: {},
		268948: {},
		268949: {},
	}
)

func Init(passPhrase string, oauthAvailable bool) {
	var err error

	clientID, err = aes.Decrypt(clientIDCrypt, passPhrase)
	if err != nil {
		log.Fatal("unable to decrypt clientID", err)
	}

	clientSecret, err = aes.Decrypt(clientSecretCrypt, passPhrase)
	if err != nil {
		log.Fatal("unable to decrypt clientSecret", err)
	}

	accessToken, err = wowAccessToken()
	if err != nil {
		log.Fatal("unable to get access token", err)
	}

	if !oauthAvailable {
		return
	}

	var ok bool
	profileAccessToken, ok = wowProfileAccessToken()
	if !ok {
		log.Fatal("unable to get oauthAvailable access token", err)
	}
}

// SkipItem returns true if the caller should ignore this item
func SkipItem(item int64) bool {
	_, ok := skipItems[item]
	return ok
}

// realmToSlug returns the slug form of a given realm name
func realmToSlug(realm string) string {
	slug := strings.ToLower(realm)
	slug = strings.ReplaceAll(slug, "-", "")
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}

func request(url, token, caller string) (interface{}, bool) {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}

	response, err := web.RequestJSON(url, headers)
	if err != nil {
		fmt.Printf("%s: no data returned: %s", caller, err)
		return nil, false
	}

	return response, true
}

func requestKey(url, token, key, caller string) ([]interface{}, bool) {
	r, ok := request(url, token, caller)
	if !ok {
		return nil, false
	}
	response := r.(map[string]interface{})
	return response[key].([]interface{}), true
}

// wowProfileAccessToken returns a profile access token (to authenticate user profile API calls)
func wowProfileAccessToken() (string, bool) {
	return oauth2.ProfileAccessToken(clientID, clientSecret)
}

// wowAccessToken retrieves an access token. This token is used to authenticate API calls.
func wowAccessToken() (string, error) {
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
		return "", err
	}

	var jsonObject map[string]interface{}

	err = json.Unmarshal(contents, &jsonObject)
	if err != nil {
		return "", err
	}

	return jsonObject["access_token"].(string), nil
}

// ConnectedRealm returns all realms connected to the given realm ID
func ConnectedRealm(realmId string) map[string]interface{} {
	url := "https://us.api.blizzard.com/data/wow/connected-realm/" + realmId + "?namespace=dynamic-us&locale=en_US"
	r, ok := request(url, accessToken, "ConnectedRealm")
	if !ok {
		return nil
	}

	response := r.(map[string]interface{})
	if response["code"] != nil {
		fmt.Println("ConnectedRealm: Failed to get connected realm:", response)
		return nil
	}

	return response
}

// ConnectedRealmSearch returns the set of all connected realms
func ConnectedRealmSearch() map[string]interface{} {
	url := "https://us.api.blizzard.com/data/wow/search/connected-realm?namespace=dynamic-us&status.type=UP"
	r, ok := request(url, accessToken, "ConnectedRealm")
	if !ok {
		return nil
	}

	response := r.(map[string]interface{})
	if response["code"] != nil {
		fmt.Println("ConnectedRealmSearch: Failed to get connected realms:", response)
		return nil
	}

	return response
}

var crIDs = map[string]string{
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
	"IceCrown":          "104",
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

// ConnectedRealmId returns the connected realm ID of the given realm
func ConnectedRealmId(realm string) (string, bool) {
	id, ok := crIDs[realm]
	if ok {
		return id, true
	}

	connectedRealms := ConnectedRealmSearch()
	if connectedRealms == nil {
		return "", false
	}

	slug := realmToSlug(realm)

	results := connectedRealms["results"].([]interface{})
	for _, result := range results {
		r := result.(map[string]interface{})
		data := r["data"].(map[string]interface{})
		cRealmId := web.ToString(data["id"])
		cr := ConnectedRealm(cRealmId)
		if cr == nil {
			continue
			//return "", false
		}
		realms := cr["realms"].([]interface{})
		for _, cRealm := range realms {
			realmSlug := cRealm.(map[string]interface{})["slug"].(string)
			if slug == realmSlug {
				mapItem := fmt.Sprintf("  \"%s\": \"%s\",\n", realm, cRealmId)
				fmt.Println(mapItem)
				return cRealmId, true
			}
		}
	}

	fmt.Println("ConnectedRealmId: Failed to find realm:", realm)
	return "", false
}

// Auctions returns the current auctions from the auction house
func Auctions(realm string) ([]interface{}, bool) {
	connectedRealmId, ok := ConnectedRealmId(realm)
	if !ok {
		fmt.Println("Auctions: no connected realm id found")
		return nil, false
	}

	url := "https://us.api.blizzard.com/data/wow/connected-realm/" + connectedRealmId + "/auctions?namespace=dynamic-us&locale=en_US"
	r, ok := request(url, accessToken, "Auctions")
	if !ok {
		return nil, false
	}

	response := r.(map[string]interface{})
	if response["code"] != nil {
		fmt.Println("Auctions: HTTP error:", response)
		return nil, false
	}

	auctions := response["auctions"].([]interface{})
	return auctions, true
}

// Commodities returns the current commodity auctions from the auction house
func Commodities() ([]interface{}, bool) {
	url := "https://us.api.blizzard.com/data/wow/auctions/commodities?namespace=dynamic-us&locale=en_US"
	return requestKey(url, accessToken, "auctions", "Commodities")
}

// WowItem retrieves a single item from the WoW web API
func WowItem(id string) (map[string]interface{}, bool) {
	url := "https://us.api.blizzard.com/data/wow/item/" + id + "?namespace=static-us&locale=en_US"
	r, ok := request(url, accessToken, "Auctions")
	if !ok {
		return nil, false
	}

	response := r.(map[string]interface{})
	if response["status"] == "nok" {
		fmt.Println("INFO: ", response["reason"], "id: ", id)
		return nil, false
	}
	_, ok = response["code"]
	if ok {
		fmt.Println("Error retrieving id: ", id, response)
		return nil, false
	}

	return response, true
}

// Pets returns a list of all battle pets in the game
func Pets() ([]interface{}, bool) {
	url := "https://us.api.blizzard.com/data/wow/pet/index?namespace=static-us&locale=en_US"
	return requestKey(url, profileAccessToken, "pets", "Pets")
}

// CollectionsPets returns the battle pets the user owns
func CollectionsPets() ([]interface{}, bool) {
	url := "https://us.api.blizzard.com/profile/user/wow/collections/pets?namespace=profile-us&locale=en_US"
	return requestKey(url, profileAccessToken, "pets", "CollectionsPets")
}

// Toys returns a list of all toys in the game
func Toys() ([]interface{}, bool) {
	url := "https://us.api.blizzard.com/data/wow/toy/index?namespace=static-us&locale=en_US"
	return requestKey(url, profileAccessToken, "toys", "Toys")
}

// CollectionsToys returns the toys the user owns
func CollectionsToys() ([]interface{}, bool) {
	url := "https://us.api.blizzard.com/profile/user/wow/collections/toys?namespace=profile-us&locale=en_US"
	return requestKey(url, profileAccessToken, "toys", "CollectionsToys")
}

// ItemAppearanceSetsIndex returns IDs of each appearance set
func ItemAppearanceSetsIndex() ([]interface{}, bool) {
	url := "https://us.api.blizzard.com/data/wow/item-appearance/set/index?namespace=static-us&locale=en_US"
	return requestKey(url, accessToken, "appearance_sets", "ItemAppearanceSetsIndex")
}

// ItemAppearanceSetsIndexIds returns the ID and name of each appearance set
func ItemAppearanceSetsIndexIds() map[int64]string {
	index, ok := ItemAppearanceSetsIndex()
	if !ok {
		return nil
	}

	indexMap := map[int64]string{}
	for _, i := range index {
		i := i.(map[string]interface{})
		id := web.ToInt64(i["id"])
		name := web.ToString(i["name"])
		indexMap[id] = name
	}

	return indexMap
}

// ItemAppearanceSet returns the appearance IDs of the given appearance set
func ItemAppearanceSet(appearanceId int64) ([]interface{}, bool) {
	url := fmt.Sprintf("https://us.api.blizzard.com/data/wow/item-appearance/set/%d?namespace=static-us&locale=en_US", appearanceId)
	return requestKey(url, accessToken, "appearances", "ItemAppearanceSet")
}

// ItemAppearanceSetIds returns a slice of the IDs for the given appearance set
func ItemAppearanceSetIds(appearanceId int64) []int64 {
	itemSet, ok := ItemAppearanceSet(appearanceId)
	if !ok {
		return nil
	}

	ids := []int64{}
	for _, i := range itemSet {
		i := i.(map[string]interface{})
		ids = append(ids, web.ToInt64(i["id"]))
	}

	return ids
}

// ItemAppearanceSlotIndex returns a list of slot names
func ItemAppearanceSlotIndex() []string {
	// This would query the Blizzard URL:
	//url := "https://us.api.blizzard.com/data/wow/item-appearance/slot/index?namespace=static-us&locale=en_US"
	// But that returns a static list, so we cache the results and return those.

	return []string{
		"HEAD",
		"SHOULDER",
		"BODY",
		"CHEST",
		"WAIST",
		"LEGS",
		"FEET",
		"WRIST",
		"HAND",
		"WEAPON",
		"SHIELD",
		"RANGED",
		"CLOAK",
		"TWOHWEAPON",
		"TABARD",
		"ROBE",
		"WEAPONMAINHAND",
		"WEAPONOFFHAND",
		"HOLDABLE",
		"AMMO",
		"RANGEDRIGHT",
		"PROFESSION_TOOL",
		"PROFESSION_GEAR",
		"EQUIPABLESPELL_WEAPON",
	}
}

// ItemAppearanceSlot returns a list of appearances for a given slot
func ItemAppearanceSlot(slotName string) ([]interface{}, bool) {
	url := "https://us.api.blizzard.com/data/wow/item-appearance/slot/" + slotName + "?namespace=static-us&locale=en_US"
	return requestKey(url, profileAccessToken, "appearances", "ItemAppearanceSlot")
}

// ItemAppearance returns the details of a given item appearance ID
func ItemAppearance(itemAppearanceId int64) (interface{}, bool) {
	url := fmt.Sprintf("https://us.api.blizzard.com/data/wow/item-appearance/%d?namespace=static-us", itemAppearanceId)
	return request(url, profileAccessToken, "ItemAppearance")
}

// CollectionsTransmogs returns the transmogs the user owns
func CollectionsTransmogs() (interface{}, bool) {
	url := "https://us.api.blizzard.com/profile/user/wow/collections/transmogs?namespace=profile-us&locale=en_US"
	return request(url, profileAccessToken, "CollectionsTransmogs")
}
