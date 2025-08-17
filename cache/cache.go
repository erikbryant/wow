package cache

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"

	"github.com/erikbryant/wow/item"
)

var (
	itemCache     = map[int64]item.Item{}
	itemCacheFile = "./generated/itemCache.gob"
	readDisabled  = false
	mu            sync.Mutex
)

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	load()
	fmt.Printf("-- #Items in cache: %d\n", len(itemCache))
}

// load loads the disk cache file into memory
func load() {
	file, err := os.Open(itemCacheFile)
	if err != nil {
		fmt.Printf("*** error opening itemCache file: %v, creating new one\n", err)
		return
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	mu.Lock()
	err = decoder.Decode(&itemCache)
	mu.Unlock()
	if err != nil {
		log.Fatalf("error reading itemCache: %v", err)
	}
}

// Save writes the in-memory cache file to disk
func Save() {
	file, err := os.Create(itemCacheFile)
	if err != nil {
		log.Fatalf("error creating itemCache file: %v", err)
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	mu.Lock()
	err = encoder.Encode(itemCache)
	mu.Unlock()
	if err != nil {
		log.Fatalf("error encoding itemCache: %v", err)
	}
}

// Read returns the in-memory copy (if exists)
func Read(id int64) (item.Item, bool) {
	if readDisabled {
		return item.Item{}, false
	}
	mu.Lock()
	i, ok := itemCache[id]
	mu.Unlock()
	return i, ok
}

// Write writes an entry to the in-memory cache
func Write(id int64, i item.Item) {
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

// IDs returns the sorted list of keys from itemCache
func IDs() []int64 {
	ids := []int64{}

	mu.Lock()
	for id := range itemCache {
		ids = append(ids, id)
	}
	mu.Unlock()

	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	return ids
}

// Items returns the map of cached items
func Items() map[int64]item.Item {
	// I am not sure that this is thread-safe
	return itemCache
}

// Print writes a text version of the in-memory cache to stdout
func Print() {
	for _, id := range IDs() {
		mu.Lock()
		i := itemCache[id]
		mu.Unlock()
		fmt.Println(i.Format())
	}
}

// Search returns the item with name 's' or an empty item if not found
func Search(s string) item.Item {
	mu.Lock()
	for id := range itemCache {
		if itemCache[id].Name() == s {
			mu.Unlock()
			return itemCache[id]
		}
	}
	mu.Unlock()

	fmt.Println("Did not find item for search string: ", s)
	return item.Item{}
}

func DisableRead() {
	readDisabled = true
}

func EnableRead() {
	readDisabled = false
}

// LuaVendorPrice writes the cached vendor sell prices to stdout as a lua table and accessor
func LuaVendorPrice() string {
	lua := ""

	lua += fmt.Sprintf("local VendorSellPriceCache = {\n")
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
		lua += fmt.Sprintf("  [\"%d\"] = %d,\n", id, spr)
	}
	lua += fmt.Sprintf("}\n")

	lua += fmt.Sprintf(`
-- VendorSellPrice returns the cached vendor sell price
local function VendorSellPrice(itemID)
    return VendorSellPriceCache[tostring(itemID)] or 0
end

-- validatePriceCache verifies each cached sell price matches the actual sell price
local function validatePriceCache()
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

-- Validate the sell price cache
C_Timer.After(1, validatePriceCache)

AhaPriceCache = {
  VendorSellPrice = VendorSellPrice,
}
`)

	return lua
}
