package itemcache

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

const (
	cacheFilename = "./data/itemCache.gob"
)

var (
	itemCache    = map[int64]wowitem.Item{}
	readDisabled = false
	mu           sync.Mutex
)

func init() {
	gob.Register(map[string]any{})
	gob.Register([]any{})
	load()
	fmt.Printf("-- #Items in cache: %d\n", len(itemCache))
}

// load loads the disk cache file into memory
func load() {
	file, err := os.Open(cacheFilename)
	if err != nil {
		fmt.Printf("*** error opening item cache file: %v, creating new one\n", err)
		return
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	mu.Lock()
	err = decoder.Decode(&itemCache)
	mu.Unlock()
	if err != nil {
		log.Fatalf("error reading item cache: %v", err)
	}
}

// Save writes the in-memory cache file to disk
func Save() error {
	mu.Lock()
	defer mu.Unlock()

	tmp := cacheFilename + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	encoder := gob.NewEncoder(f)

	if err := encoder.Encode(itemCache); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}

	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}

	return os.Rename(tmp, cacheFilename)
}

// Read returns the in-memory copy (if exists)
func Read(id int64) (wowitem.Item, bool) {
	if readDisabled {
		return wowitem.Item{}, false
	}
	mu.Lock()
	i, ok := itemCache[id]
	mu.Unlock()
	return i, ok
}

// Write writes an entry to the in-memory cache
func Write(id int64, i wowitem.Item) {
	mu.Lock()
	itemCache[id] = i
	mu.Unlock()
}

// Delete deletes an entry from the in-memory cache
func Delete(id int64) {
	mu.Lock()
	delete(itemCache, id)
	mu.Unlock()
}

// IDs returns the sorted list of keys from the item cache file
func IDs() []int64 {
	ids := []int64{}

	mu.Lock()
	for id := range itemCache {
		ids = append(ids, id)
	}
	mu.Unlock()

	slices.Sort(ids)

	return ids
}

// ItemsSlice returns a slice of the cached map values
func ItemsSlice() []wowitem.Item {
	mu.Lock()
	defer mu.Unlock()

	items := make([]wowitem.Item, 0, len(itemCache))

	for _, item := range itemCache {
		items = append(items, item)
	}

	return items
}

// Search returns the item with name 's' or an empty item if not found
func Search(s string) wowitem.Item {
	mu.Lock()
	for id := range itemCache {
		if itemCache[id].Name() == s {
			mu.Unlock()
			return itemCache[id]
		}
	}
	mu.Unlock()

	fmt.Println("Did not find item for search string: ", s)
	return wowitem.Item{}
}

func DisableRead() {
	readDisabled = true
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
		mu.Lock()
		spr := itemCache[id].SellPriceRealizable()
		mu.Unlock()
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
		mu.Lock()
		cosmetic := itemCache[id].Cosmetic()
		mu.Unlock()
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
