package itempersistence

import (
	"encoding/gob"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

const (
	persistName = "items"
)

var (
	items = persist.New[int64, wowitem.Item](persistName)
)

func init() {
	gob.Register(map[string]any{})
	gob.Register([]any{})
	err := items.Load()
	if err != nil {
		fmt.Printf("*** error opening items persist, creating new one: %v\n", err)
	}
	fmt.Printf("-- #Items in cache: %d\n", items.Len())
}

func Save() error {
	return items.Save()
}

// Read returns the in-memory copy (if exists)
func Read(id int64) (wowitem.Item, bool) {
	return items.Get(id)
}

// Write writes an entry to the in-memory cache
func Write(id int64, i wowitem.Item) {
	items.Set(id, i)
}

// Delete deletes an entry from the in-memory cache
func Delete(id int64) {
	items.Delete(id)
}

// IDs returns the sorted list of keys from the item cache file
func IDs() []int64 {
	keys := items.Keys()
	slices.Sort(keys)
	return keys
}

// ItemValues returns a slice of the cached map values
func ItemValues() []wowitem.Item {
	return items.Values()
}

// Search returns the item with name 's' or an empty item if not found
func Search(s string) wowitem.Item {
	_, i, ok := items.Search(func(v wowitem.Item) bool {
		return v.Name() == s
	})
	if !ok {
		fmt.Println("Did not find item for search string: ", s)
	}
	return i
}

// LookupItem retrieves data for a single item. From the cache if present, or web if not. If it retrieves it from the web it also caches it.
func LookupItem(id int64, age time.Duration) (wowitem.Item, bool) {
	// Use the cached value if exists and not stale
	i, ok := Read(id)
	if ok {
		// A cache hit, but is the cache stale?
		if !i.Stale(age) {
			return i, true
		}
		fmt.Println("Refreshing stale item:", i.Format())
	}

	result, ok := wowapi.Item(web.ToString(id))
	if !ok {
		return wowitem.Item{}, false
	}
	i = wowitem.NewItem(result)
	Write(i.ID(), i)

	return i, true
}

func luaVendorPrice() (string, []string) {
	var lua strings.Builder

	lua.WriteString(fmt.Sprintf("local VendorSellPriceCache = {\n"))
	for _, id := range IDs() {
		i, _ := items.Get(id)
		spr := i.SellPriceRealizable()
		if spr <= 100 {
			// To keep the lua table small, ignore anything that can't ever be a bargain
			// Skip prices that are zero
			// Skip prices <= one silver (the auction house does not deal in copper)
			continue
		}
		lua.WriteString(fmt.Sprintf("  [\"%d\"] = %d,\n", id, spr))
	}
	lua.WriteString(fmt.Sprintf("}\n"))

	lua.WriteString(fmt.Sprintf(`
-- VendorSellPrice returns the cached vendor sell price
local function VendorSellPrice(itemID)
    return VendorSellPriceCache[tostring(itemID)] or 0
end

-- ValidatePriceCache verifies each cached sell price matches the actual sell price
local function ValidatePriceCache()
    for itemID, cachedPrice in pairs(VendorSellPriceCache) do
        itemID = tonumber(itemID)
        local item = Item:CreateFromItemID(itemID)
        item:ContinueOnItemLoad(
                function()
                    local itemInfo = { C_Item.GetItemInfo(itemID) }
                    local sellPrice = itemInfo[11]
                    if cachedPrice ~= sellPrice then
                        MerchUtil.PrettyPrint("Cached price mismatch!", itemID, GetCoinTextureString(cachedPrice), "~=", GetCoinTextureString(sellPrice))
                    end
                end
        )
    end
end
`))

	return lua.String(), []string{"VendorSellPrice", "ValidatePriceCache"}
}

func luaCosmetic() (string, []string) {
	var lua strings.Builder

	lua.WriteString(fmt.Sprintf("local Cosmetics = {\n"))
	for _, id := range IDs() {
		i, _ := items.Get(id)
		cosmetic := i.Cosmetic()
		if !cosmetic {
			continue
		}
		lua.WriteString(fmt.Sprintf("  [\"%d\"] = true,\n", id))
	}
	lua.WriteString(fmt.Sprintf("}\n"))

	lua.WriteString(fmt.Sprintf(`
-- Cosmetic returns true if the item is a Cosmetic
local function Cosmetic(itemID)
    return Cosmetics[tostring(itemID)] or false
end
`))

	return lua.String(), []string{"Cosmetic"}
}

// Lua the cached vendor sell prices to stdout as a lua table and accessor
func Lua() string {
	lua := ""

	lua, fcns := luaVendorPrice()

	lua2, fcns2 := luaCosmetic()

	lua += "\n" + lua2
	for _, fn := range fcns2 {
		fcns = append(fcns, fn)
	}

	lua += fmt.Sprintf("\nPriceCache = {\n")
	for _, fn := range fcns {
		lua += fmt.Sprintf("  %s = %s,\n", fn, fn)
	}
	lua += fmt.Sprintf("}\n")

	return lua
}
