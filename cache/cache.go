package cache

import (
	"encoding/gob"
	"fmt"
	"github.com/erikbryant/wow/common"
	"os"
	"sort"
)

var (
	itemCache     = map[int64]common.Item{}
	itemCacheFile = "itemCache.gob"
)

func init() {
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
		fmt.Printf("error reading itemCache: %v", err)
		panic(err)
	}
}

// save writes the in-memory cache file to disk
func save() {
	file, err := os.Create(itemCacheFile)
	if err != nil {
		fmt.Printf("error creating itemCache file: %v", err)
		panic(err)
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	encoder.Encode(itemCache)
}

// Read returns the in-memory copy (if exists)
func Read(id int64) (common.Item, bool) {
	item, ok := itemCache[id]
	return item, ok
}

// Write writes an entry to the in-memory cache
func Write(id int64, item common.Item) {
	itemCache[id] = item
	save()
}

// Print writes a text version of the in-memory cache to stdout
func Print() {
	for id, item := range itemCache {
		fmt.Printf("%-50s %d - %v\n", item.Name, id, item)
	}
}

// sortItemCacheKeys returns the sorted list of keys from itemCache
func sortItemCacheKeys(dict map[int64]common.Item) []int64 {
	keys := []int64{}

	for k := range dict {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	return keys
}

// PrintLuaEquippable writes the cached equippable status to stdout as a lua table and accessor
func PrintLuaEquippable() {
	fmt.Println("local ItemIsEquippableCache = {")
	for _, key := range sortItemCacheKeys(itemCache) {
		if itemCache[key].Equippable {
			fmt.Printf("  [\"%d\"] = true,\n", key)
		}
	}
	fmt.Println("}")

	luaFunc := `
function ItemCache:ItemIsEquippable(itemID)
	return ItemIsEquippableCache[tostring(itemID)]
end`

	fmt.Println(luaFunc)
}

// PrintLuaVendorPrice writes the cached vendor sell prices to stdout as a lua table and accessor
func PrintLuaVendorPrice() {
	fmt.Println("local VendorSellPriceCache = {")
	for _, key := range sortItemCacheKeys(itemCache) {
		fmt.Printf("  [\"%d\"] = %d,\n", key, itemCache[key].SellPrice)
	}
	fmt.Println("}")

	luaFunc := `
function ItemCache:VendorSellPrice(itemID)
    local sellPrice = VendorSellPriceCache[tostring(itemID)]

    if sellPrice == nil then
        print("Arbitrages: No cached vendor price for ", itemID)
        local itemInfo = { C_Item.GetItemInfo(itemID) }
        sellPrice = itemInfo[11]
        if sellPrice == nil then
			sellPrice = 0
        end
    end

    return sellPrice
end`

	fmt.Println(luaFunc)
}
