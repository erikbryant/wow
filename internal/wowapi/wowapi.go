package wowapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/erikbryant/wow/internal/common"
	"github.com/erikbryant/wow/internal/regions"
)

type Client struct {
	accessToken        string
	profileAccessToken string
	httpClient         *http.Client
	realmList          regions.Realms

	apiBaseTest string // Only used by tests
}

// NewClient returns a WoW API client. Since there is only one WoW web API,
// we always return the default client.
func NewClient(secretPath string) (*Client, error) {
	var err error

	c := Client{
		httpClient: http.DefaultClient,
		realmList:  regions.New(),
	}

	clientID, clientSecret, err := getSecretsFromKeychain(secretPath)
	if err != nil {
		return nil, err
	}

	c.accessToken, c.profileAccessToken, err = authenticate(clientID, clientSecret)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// NewClientWithHTTP creates a WoW API client using the supplied API base URL
// and HTTP client. It is only used for tests.
func NewClientWithHTTP(
	apiBase string,
	httpClient *http.Client,
) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		apiBaseTest: strings.TrimRight(apiBase, "/"),
		httpClient:  httpClient,
		realmList:   regions.New(),
	}
}

func (c *Client) apiBase(realm string) string {
	if realm == "testRealm" || c.apiBaseTest != "" {
		return c.apiBaseTest
	}
	region, _ := c.realmList.Config(realm)
	return fmt.Sprintf("https://%s.api.blizzard.com", strings.ToLower(region))
}

func (c *Client) getRegion(realm string) string {
	region, _ := c.realmList.Config(realm)
	return region
}

func (c *Client) getLanguage(realm string) string {
	_, language := c.realmList.Config(realm)
	return language
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
	decoder.UseNumber()
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

// Auctions returns the current auctions from the auction house.
func (c *Client) Auctions(realm string) ([]any, error) {
	connectedRealmID, err := regions.ConnectedRealmID(realm)
	if err != nil {
		return nil, fmt.Errorf(
			"auctions: no connected realm ID found: %w",
			err,
		)
	}

	rawURL := fmt.Sprintf(
		"%s/data/wow/connected-realm/%s/auctions?namespace=dynamic-%s&locale=%s",
		c.apiBase(realm),
		connectedRealmID,
		c.getRegion(realm),
		c.getLanguage(realm),
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
	// TODO: are the commodity auctions different across regions?
	rawURL := fmt.Sprintf(
		"%s/data/wow/auctions/commodities?namespace=dynamic-%s&locale=%s",
		c.apiBase("Commodities"),
		c.getRegion("Commodities"),
		c.getLanguage("Commodities"),
	)

	return c.requestKey(
		rawURL,
		c.accessToken,
		"auctions",
		"Commodities",
	)
}

// Item retrieves a single item from the WoW web API.
func (c *Client) Item(id string) (map[string]any, error) {
	// TODO: Allow retrieval of non-US item data
	realm := "Aegwynn"
	rawURL := fmt.Sprintf(
		"%s/data/wow/item/%s?namespace=static-%s&locale=%s",
		c.apiBase(realm),
		id,
		c.getRegion(realm),
		c.getLanguage(realm),
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
	// TODO: Allow retrieval of non-US data?
	realm := "Aegwynn"
	rawURL := fmt.Sprintf(
		"%s/data/wow/pet/index?namespace=static-%s&locale=%s",
		c.apiBase(realm),
		c.getRegion(realm),
		c.getLanguage(realm),
	)

	return c.requestKey(
		rawURL,
		c.profileAccessToken,
		"pets",
		"Pets",
	)
}

// CollectionsPets returns the battle pets the user owns.
func (c *Client) CollectionsPets() ([]any, error) {
	// TODO: Allow retrieval of non-US data?
	realm := "Aegwynn"
	rawURL := fmt.Sprintf(
		"%s/profile/user/wow/collections/pets?namespace=profile-%s&locale=%s",
		c.apiBase(realm),
		c.getRegion(realm),
		c.getLanguage(realm),
	)

	return c.requestKey(
		rawURL,
		c.profileAccessToken,
		"pets",
		"CollectionsPets",
	)
}

// Toys returns a list of all toys in the game.
func (c *Client) Toys() ([]any, error) {
	// TODO: Allow retrieval of non-US data?
	realm := "Aegwynn"
	rawURL := fmt.Sprintf(
		"%s/data/wow/toy/index?namespace=static-%s&locale=%s",
		c.apiBase(realm),
		c.getRegion(realm),
		c.getLanguage(realm),
	)

	return c.requestKey(
		rawURL,
		c.profileAccessToken,
		"toys",
		"Toys",
	)
}

// CollectionsToys returns the toys the user owns.
func (c *Client) CollectionsToys() ([]any, error) {
	// TODO: Allow retrieval of non-US data?
	realm := "Aegwynn"
	rawURL := fmt.Sprintf(
		"%s/profile/user/wow/collections/toys?namespace=profile-%s&locale=%s",
		c.apiBase(realm),
		c.getRegion(realm),
		c.getLanguage(realm),
	)

	return c.requestKey(
		rawURL,
		c.profileAccessToken,
		"toys",
		"CollectionsToys",
	)
}

// ItemAppearanceSetsIndex returns IDs of each appearance set.
func (c *Client) ItemAppearanceSetsIndex() ([]any, error) {
	// TODO: Allow retrieval of non-US data?
	realm := "Aegwynn"
	rawURL := fmt.Sprintf(
		"%s/data/wow/item-appearance/set/index?namespace=static-%s&locale=%s",
		c.apiBase(realm),
		c.getRegion(realm),
		c.getLanguage(realm),
	)

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
				"ItemAppearanceSetsIndexIDs need map[string]any, got: %v",
				item,
			)
		}

		id, err := common.JSONInt64(item["id"])
		if err != nil {
			return nil, fmt.Errorf(
				"ItemAppearanceSetsIndexIDs: invalid ID: %v, %w",
				item["id"],
				err,
			)
		}

		indexMap[id] = common.JSONString(item["name"])
	}

	return indexMap, nil
}

// ItemAppearanceSet returns the appearance IDs of the given appearance set.
func (c *Client) ItemAppearanceSet(appearanceID int64) ([]any, error) {
	// TODO: Allow retrieval of non-US data?
	realm := "Aegwynn"
	rawURL := fmt.Sprintf(
		"%s/data/wow/item-appearance/set/%d?namespace=static-%s&locale=%s",
		c.apiBase(realm),
		appearanceID,
		c.getRegion(realm),
		c.getLanguage(realm),
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

		id, err := common.JSONInt64(item["id"])
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
	// TODO: Allow retrieval of non-US data?
	realm := "Aegwynn"
	rawURL := fmt.Sprintf(
		"%s/profile/user/wow/collections/transmogs?namespace=profile-%s&locale=%s",
		c.apiBase(realm),
		c.getRegion(realm),
		c.getLanguage(realm),
	)

	return c.request(
		rawURL,
		c.profileAccessToken,
		"CollectionsTransmogs",
	)
}

// Professions returns the professions this alt knows.
func (c *Client) Professions(realm, alt string) (any, error) {
	realm = regions.RealmToSlug(realm)
	alt = strings.ToLower(alt)

	rawURL := fmt.Sprintf(
		"%s/profile/wow/character/%s/%s/professions?namespace=profile-%s&locale=%s",
		c.apiBase(realm),
		realm,
		alt,
		c.getRegion(realm),
		c.getLanguage(realm),
	)

	return c.request(
		rawURL,
		c.profileAccessToken,
		"Professions",
	)
}
