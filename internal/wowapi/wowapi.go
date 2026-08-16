package wowapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

const defaultAPIBase = "https://us.api.blizzard.com"

type Client struct {
	clientID           string
	clientSecret       string
	accessToken        string
	profileAccessToken string

	apiBase    string
	httpClient *http.Client
}

var (
	defaultClientMu sync.RWMutex
	defaultClient   *Client
)

// NewClient returns a WoW API client. Since there is only one WoW web API,
// we always return the default client.
func NewClient() (*Client, error) {
	defaultClientMu.RLock()
	client := defaultClient
	defaultClientMu.RUnlock()

	if client == nil {
		return nil, fmt.Errorf("wow API is not authenticated, did you call wowapi.Init()")
	}

	return client, nil
}

// NewClientWithHTTP creates a WoW API client using the supplied API base URL
// and HTTP client. It is only used for tests.
func NewClientWithHTTP(
	clientID,
	clientSecret,
	apiBase string,
	httpClient *http.Client,
) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		apiBase:      strings.TrimRight(apiBase, "/"),
		httpClient:   httpClient,
	}
}

func (c *Client) request(rawURL, token, caller string) (any, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: unable to create request: %w", caller, err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: no data returned: %w", caller, err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK ||
		response.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf(
			"%s: HTTP status %d",
			caller,
			response.StatusCode,
		)
	}

	var result any

	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf(
			"%s: unable to decode response: %w",
			caller,
			err,
		)
	}

	return result, nil
}

func (c *Client) requestKey(
	rawURL,
	token,
	key,
	caller string,
) ([]any, error) {
	r, err := c.request(rawURL, token, caller)
	if err != nil {
		return nil, err
	}

	response, ok := r.(map[string]any)
	if !ok {
		return nil, fmt.Errorf(
			"%s: expected object response, got %T",
			caller,
			r,
		)
	}

	value, ok := response[key]
	if !ok {
		return nil, fmt.Errorf(
			"%s: response is missing key %q",
			caller,
			key,
		)
	}

	result, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf(
			"%s: response key %q has type %T, want []any",
			caller,
			key,
			value,
		)
	}

	return result, nil
}

// realmToSlug returns the slug form of a given realm name, based on WoW
// naming rules.
func realmToSlug(realm string) string {
	slug := strings.ToLower(realm)
	slug = strings.ReplaceAll(slug, "-", "")
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}

// connectedRealmIDCache caches the calls to find a connected realm ID,
// which are slow.
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

// ConnectedRealm returns all realms connected to the given realm ID.
func (c *Client) ConnectedRealm(realmID string) (map[string]any, error) {
	rawURL := fmt.Sprintf(
		"%s/data/wow/connected-realm/%s?namespace=dynamic-us&locale=en_US",
		c.apiBase,
		realmID,
	)

	r, err := c.request(rawURL, c.accessToken, "ConnectedRealm")
	if err != nil {
		return nil, err
	}

	response, ok := r.(map[string]any)
	if !ok {
		return nil, fmt.Errorf(
			"ConnectedRealm: expected object response, got %T",
			r,
		)
	}

	if response["code"] != nil {
		return nil, fmt.Errorf(
			"ConnectedRealm failed to get connected realm: %v",
			response["code"],
		)
	}

	return response, nil
}

// ConnectedRealmSearch returns the set of all connected realms.
func (c *Client) ConnectedRealmSearch() (map[string]any, error) {
	rawURL := c.apiBase +
		"/data/wow/search/connected-realm?namespace=dynamic-us&status.type=UP"

	r, err := c.request(rawURL, c.accessToken, "ConnectedRealm")
	if err != nil {
		return nil, err
	}

	response, ok := r.(map[string]any)
	if !ok {
		return nil, fmt.Errorf(
			"ConnectedRealmSearch: expected object response, got %T",
			r,
		)
	}

	if response["code"] != nil {
		return nil, fmt.Errorf(
			"ConnectedRealmSearch failed to get connected realms: %v",
			response,
		)
	}

	return response, nil
}

// ConnectedRealmID returns the connected realm ID of the given realm.
func (c *Client) ConnectedRealmID(realm string) (string, error) {
	id, ok := connectedRealmIDCache[realm]
	if ok {
		return id, nil
	}

	connectedRealms, err := c.ConnectedRealmSearch()
	if err != nil {
		return "", fmt.Errorf("no connected realm found: %w", err)
	}

	slug := realmToSlug(realm)

	results, ok := connectedRealms["results"].([]any)
	if !ok {
		return "", fmt.Errorf(
			"ConnectedRealmID: response results has type %T, want []any",
			connectedRealms["results"],
		)
	}

	for _, result := range results {
		r, ok := result.(map[string]any)
		if !ok {
			return "", fmt.Errorf(
				"ConnectedRealmID: result has type %T, want object",
				result,
			)
		}

		data, ok := r["data"].(map[string]any)
		if !ok {
			return "", fmt.Errorf(
				"ConnectedRealmID: result data has type %T, want object",
				r["data"],
			)
		}

		cRealmID, err := jsonString(data["id"])
		if err != nil {
			return "", fmt.Errorf(
				"ConnectedRealmID: invalid realm ID: %w",
				err,
			)
		}

		cr, err := c.ConnectedRealm(cRealmID)
		if err != nil {
			return "", err
		}

		realms, ok := cr["realms"].([]any)
		if !ok {
			return "", fmt.Errorf(
				"ConnectedRealmID: response realms has type %T, want []any",
				cr["realms"],
			)
		}

		for _, connectedRealm := range realms {
			connectedRealm, ok := connectedRealm.(map[string]any)
			if !ok {
				return "", fmt.Errorf(
					"ConnectedRealmID: realm has type %T, want object",
					connectedRealm,
				)
			}

			realmSlug, ok := connectedRealm["slug"].(string)
			if !ok {
				return "", fmt.Errorf(
					"ConnectedRealmID: realm slug has type %T, want string",
					connectedRealm["slug"],
				)
			}

			if slug == realmSlug {
				return cRealmID, nil
			}
		}
	}

	return "", fmt.Errorf(
		"ConnectedRealmID failed to find realm: %s",
		realm,
	)
}

// Auctions returns the current auctions from the auction house.
func (c *Client) Auctions(realm string) ([]any, error) {
	connectedRealmID, err := c.ConnectedRealmID(realm)
	if err != nil {
		return nil, fmt.Errorf(
			"auctions: no connected realm ID found: %w",
			err,
		)
	}

	rawURL := fmt.Sprintf(
		"%s/data/wow/connected-realm/%s/auctions?namespace=dynamic-us&locale=en_US",
		c.apiBase,
		connectedRealmID,
	)

	r, err := c.request(rawURL, c.accessToken, "Auctions")
	if err != nil {
		return nil, err
	}

	response, ok := r.(map[string]any)
	if !ok {
		return nil, fmt.Errorf(
			"auctions: expected object response, got %T",
			r,
		)
	}

	if response["code"] != nil {
		return nil, fmt.Errorf(
			"auctions: HTTP error: %v",
			response["code"],
		)
	}

	auctions, ok := response["auctions"].([]any)
	if !ok {
		return nil, fmt.Errorf(
			"auctions: response auctions has type %T, want []any",
			response["auctions"],
		)
	}

	return auctions, nil
}

// Commodities returns the current commodity auctions from the auction house.
func (c *Client) Commodities() ([]any, error) {
	rawURL := c.apiBase +
		"/data/wow/auctions/commodities?namespace=dynamic-us&locale=en_US"

	return c.requestKey(
		rawURL,
		c.accessToken,
		"auctions",
		"Commodities",
	)
}

// Item retrieves a single item from the WoW web API.
func (c *Client) Item(id string) (map[string]any, error) {
	rawURL := fmt.Sprintf(
		"%s/data/wow/item/%s?namespace=static-us&locale=en_US",
		c.apiBase,
		id,
	)

	r, err := c.request(rawURL, c.accessToken, "Item")
	if err != nil {
		return nil, err
	}

	response, ok := r.(map[string]any)
	if !ok {
		return nil, fmt.Errorf(
			"item: expected object response, got %T",
			r,
		)
	}

	if response["status"] == "nok" {
		return nil, fmt.Errorf(
			"item: bad status for itemID %s: %v",
			id,
			response["reason"],
		)
	}

	if response["code"] != nil {
		return nil, fmt.Errorf(
			"item: error retrieving itemID %s: %v",
			id,
			response,
		)
	}

	return response, nil
}

// Pets returns a list of all battle pets in the game.
func (c *Client) Pets() ([]any, error) {
	rawURL := c.apiBase +
		"/data/wow/pet/index?namespace=static-us&locale=en_US"

	return c.requestKey(
		rawURL,
		c.profileAccessToken,
		"pets",
		"Pets",
	)
}

// CollectionsPets returns the battle pets the user owns.
func (c *Client) CollectionsPets() ([]any, error) {
	rawURL := c.apiBase +
		"/profile/user/wow/collections/pets?namespace=profile-us&locale=en_US"

	return c.requestKey(
		rawURL,
		c.profileAccessToken,
		"pets",
		"CollectionsPets",
	)
}

// Toys returns a list of all toys in the game.
func (c *Client) Toys() ([]any, error) {
	rawURL := c.apiBase +
		"/data/wow/toy/index?namespace=static-us&locale=en_US"

	return c.requestKey(
		rawURL,
		c.profileAccessToken,
		"toys",
		"Toys",
	)
}

// CollectionsToys returns the toys the user owns.
func (c *Client) CollectionsToys() ([]any, error) {
	rawURL := c.apiBase +
		"/profile/user/wow/collections/toys?namespace=profile-us&locale=en_US"

	return c.requestKey(
		rawURL,
		c.profileAccessToken,
		"toys",
		"CollectionsToys",
	)
}

// ItemAppearanceSetsIndex returns IDs of each appearance set.
func (c *Client) ItemAppearanceSetsIndex() ([]any, error) {
	rawURL := c.apiBase +
		"/data/wow/item-appearance/set/index?namespace=static-us&locale=en_US"

	return c.requestKey(
		rawURL,
		c.accessToken,
		"appearance_sets",
		"ItemAppearanceSetsIndex",
	)
}

// ItemAppearanceSetsIndexIDs returns the ID and name of each appearance set.
func (c *Client) ItemAppearanceSetsIndexIDs() (map[int64]string, error) {
	index, err := c.ItemAppearanceSetsIndex()
	if err != nil {
		return nil, err
	}

	indexMap := make(map[int64]string)

	for _, item := range index {
		item, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf(
				"ItemAppearanceSetsIndexIDs: item has type %T, want object",
				item,
			)
		}

		id, err := jsonInt64(item["id"])
		if err != nil {
			return nil, fmt.Errorf(
				"ItemAppearanceSetsIndexIDs: invalid ID: %w",
				err,
			)
		}

		name, err := jsonString(item["name"])
		if err != nil {
			return nil, fmt.Errorf(
				"ItemAppearanceSetsIndexIDs: invalid name: %w",
				err,
			)
		}

		indexMap[id] = name
	}

	return indexMap, nil
}

// ItemAppearanceSet returns the appearance IDs of the given appearance set.
func (c *Client) ItemAppearanceSet(appearanceID int64) ([]any, error) {
	rawURL := fmt.Sprintf(
		"%s/data/wow/item-appearance/set/%d?namespace=static-us&locale=en_US",
		c.apiBase,
		appearanceID,
	)

	return c.requestKey(
		rawURL,
		c.accessToken,
		"appearances",
		"ItemAppearanceSet",
	)
}

// ItemAppearanceSetIDs returns the appearance IDs that comprise the given
// appearance set.
func (c *Client) ItemAppearanceSetIDs(appearanceID int64) ([]int64, error) {
	itemSet, err := c.ItemAppearanceSet(appearanceID)
	if err != nil {
		return nil, err
	}

	appearanceIDs := make([]int64, 0, len(itemSet))

	for _, item := range itemSet {
		item, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf(
				"ItemAppearanceSetIDs: item has type %T, want object",
				item,
			)
		}

		id, err := jsonInt64(item["id"])
		if err != nil {
			return nil, fmt.Errorf(
				"ItemAppearanceSetIDs: invalid ID: %w",
				err,
			)
		}

		appearanceIDs = append(appearanceIDs, id)
	}

	return appearanceIDs, nil
}

// CollectionsTransmogs returns the transmogs the user owns.
func (c *Client) CollectionsTransmogs() (any, error) {
	rawURL := c.apiBase +
		"/profile/user/wow/collections/transmogs?namespace=profile-us&locale=en_US"

	return c.request(
		rawURL,
		c.profileAccessToken,
		"CollectionsTransmogs",
	)
}

// Professions returns the professions this alt knows.
func (c *Client) Professions(realm, alt string) (any, error) {
	realm = realmToSlug(realm)
	alt = strings.ToLower(alt)

	rawURL := fmt.Sprintf(
		"%s/profile/wow/character/%s/%s/professions?namespace=profile-us&locale=en_US",
		c.apiBase,
		realm,
		alt,
	)

	return c.request(
		rawURL,
		c.profileAccessToken,
		"Professions",
	)
}

// jsonString converts a JSON-decoded value to a string.
func jsonString(value any) (string, error) {
	switch value := value.(type) {
	case string:
		return value, nil
	case json.Number:
		return value.String(), nil
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64), nil
	case float32:
		return strconv.FormatFloat(float64(value), 'f', -1, 32), nil
	case int:
		return strconv.Itoa(value), nil
	case int64:
		return strconv.FormatInt(value, 10), nil
	case int32:
		return strconv.FormatInt(int64(value), 10), nil
	default:
		return "", fmt.Errorf("cannot convert %T to string", value)
	}
}

// jsonInt64 converts a JSON-decoded value to int64.
func jsonInt64(value any) (int64, error) {
	switch value := value.(type) {
	case int:
		return int64(value), nil
	case int64:
		return value, nil
	case int32:
		return int64(value), nil
	case float64:
		return int64(value), nil
	case float32:
		return int64(value), nil
	case json.Number:
		return value.Int64()
	case string:
		return strconv.ParseInt(value, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", value)
	}
}

// -----------------------------------------------------------------------------
// Package-level API
//
// These functions use the single authenticated client created by Authenticate
// or AuthenticateFromKeychain. Consumers therefore do not need to know the
// client credentials and do not need to carry a *Client around.
// -----------------------------------------------------------------------------

// Auctions returns the current auctions from the auction house.
func Auctions(realm string) ([]any, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return client.Auctions(realm)
}

// Commodities returns the current commodity auctions from the auction house.
func Commodities() ([]any, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return client.Commodities()
}

// Item retrieves a single item from the WoW web API.
func Item(id string) (map[string]any, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return client.Item(id)
}

// Pets returns a list of all battle pets in the game.
func Pets() ([]any, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return client.Pets()
}

// CollectionsPets returns the battle pets the user owns.
func CollectionsPets() ([]any, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return client.CollectionsPets()
}

// Toys returns a list of all toys in the game.
func Toys() ([]any, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return client.Toys()
}

// CollectionsToys returns the toys the user owns.
func CollectionsToys() ([]any, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return client.CollectionsToys()
}

// ItemAppearanceSetsIndexIDs returns the ID and name of each appearance set.
func ItemAppearanceSetsIndexIDs() (map[int64]string, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return client.ItemAppearanceSetsIndexIDs()
}

// ItemAppearanceSetIDs returns the appearance IDs that comprise the given
// appearance set.
func ItemAppearanceSetIDs(appearanceID int64) ([]int64, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return client.ItemAppearanceSetIDs(appearanceID)
}

// CollectionsTransmogs returns the transmogs the user owns.
func CollectionsTransmogs() (any, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return client.CollectionsTransmogs()
}

// Professions returns the professions this alt knows.
func Professions(realm, alt string) (any, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return client.Professions(realm, alt)
}
