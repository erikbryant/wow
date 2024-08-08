package cache

import (
	"encoding/gob"
	"fmt"
	"github.com/erikbryant/wow/common"
	"os"
	"sort"
	"time"
)

var (
	itemCache     = map[int64]common.Item{}
	itemCacheFile = "itemCache.gob"
	readDisabled  = false
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

// Migrate writes the in-memory cache file to disk in the NewItem format
func Migrate() {
	newItemCache := map[int64]common.NewItem{}

	for key, value := range itemCache {
		newValue := common.NewItem{
			Id:         value.Id,
			Name:       value.Name,
			Equippable: value.Equippable,
			SellPrice:  value.SellPrice,
			ItemLevel:  value.ItemLevel,
			Updated:    time.Now(),
		}
		newItemCache[key] = newValue
	}

	file, err := os.Create("new." + itemCacheFile)
	if err != nil {
		fmt.Printf("error creating newItemCache file: %v", err)
		panic(err)
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	encoder.Encode(newItemCache)
}

// Read returns the in-memory copy (if exists)
func Read(id int64) (common.Item, bool) {
	if readDisabled {
		return common.Item{}, false
	}
	item, ok := itemCache[id]
	return item, ok
}

// Write writes an entry to the in-memory cache
func Write(id int64, item common.Item) {
	itemCache[id] = item
	save()
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

// Print writes a text version of the in-memory cache to stdout
func Print() {
	for _, id := range IDs() {
		item := itemCache[id]
		fmt.Println(item.Format())
	}
}

func DisableRead() {
	readDisabled = true
}

func EnableRead() {
	readDisabled = false
}

// PrintLuaVendorPrice writes the cached vendor sell prices to stdout as a lua table and accessor
func PrintLuaVendorPrice() {
	fmt.Println("local VendorSellPriceCache = {")
	for _, id := range IDs() {
		// The auction house does not deal in copper; skip any items <= a full silver
		if itemCache[id].SellPrice > 100 && !itemCache[id].Equippable {
			fmt.Printf("  [\"%d\"] = %d,\n", id, itemCache[id].SellPrice)
		}
	}
	fmt.Println("}")

	luaFunc := `
local function VendorSellPrice(itemID)
    return VendorSellPriceCache[tostring(itemID)] or 0
end`

	fmt.Println(luaFunc)
}
