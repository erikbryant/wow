package cache

import (
	"encoding/gob"
	"fmt"
	"github.com/erikbryant/wow/item"
	"log"
	"os"
	"sort"
)

var (
	itemCache     = map[int64]item.Item{}
	itemCacheFile = "./generated/itemCache.gob"
	readDisabled  = false
	writeDisabled = false
)

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	load()
	fmt.Printf("-- #items in cache: %d\n\n", len(itemCache))
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
	err = decoder.Decode(&itemCache)
	if err != nil {
		log.Fatalf("error reading itemCache: %v", err)
	}
}

// Save writes the in-memory cache file to disk
func Save() {
	if writeDisabled {
		return
	}

	file, err := os.Create(itemCacheFile)
	if err != nil {
		log.Fatalf("error creating itemCache file: %v", err)
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(itemCache)
	if err != nil {
		log.Fatalf("error encoding itemCache: %v", err)
	}
}

// Read returns the in-memory copy (if exists)
func Read(id int64) (item.Item, bool) {
	if readDisabled {
		return item.Item{}, false
	}
	i, ok := itemCache[id]
	return i, ok
}

// Write writes an entry to the in-memory cache
func Write(id int64, i item.Item) {
	itemCache[id] = i
	Save()
}

// IDs returns the sorted list of keys from itemCache
func IDs() []int64 {
	ids := []int64{}

	for id := range itemCache {
		ids = append(ids, id)
	}

	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	return ids
}

// Items returns the map of cached items
func Items() map[int64]item.Item {
	return itemCache
}

// Print writes a text version of the in-memory cache to stdout
func Print() {
	for _, id := range IDs() {
		i := itemCache[id]
		fmt.Println(i.Format())
	}
}

func DisableRead() {
	readDisabled = true
}

func EnableRead() {
	readDisabled = false
}

func DisableWrite() {
	writeDisabled = true
}

func EnableWrite() {
	writeDisabled = false
}

// LuaVendorPrice writes the cached vendor sell prices to stdout as a lua table and accessor
func LuaVendorPrice() string {
	lua := ""

	lua += fmt.Sprintf("local VendorSellPriceCache = {\n")
	for _, id := range IDs() {
		if itemCache[id].SellPriceRealizable() <= 100 {
			// To keep the lua table small ignore anything that can't ever be a bargain
			// Skip prices that are zero
			// Skip prices <= one silver (the auction house does not deal in copper)
			continue
		}
		lua += fmt.Sprintf("  [\"%d\"] = %d,\n", id, itemCache[id].SellPriceRealizable())
	}
	lua += fmt.Sprintf("}\n")

	lua += fmt.Sprintf(`
local function VendorSellPrice(itemID)
    return VendorSellPriceCache[tostring(itemID)] or 0
end
`)

	lua += fmt.Sprintf(`
PriceCache = {
  VendorSellPrice = VendorSellPrice,
}
`)

	return lua
}
