package cache

import (
	"encoding/gob"
	"fmt"
	"github.com/erikbryant/wow/common"
	"os"
)

var (
	itemCache     = map[int64]common.Item{}
	itemCacheFile = "itemCache.gob"
)

func init() {
	load()
	//fmt.Printf("#Cache items: %d\n\n", len(itemCache))
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

// PrintLua writes a text version of the in-memory cache to stdout as a lua table
func PrintLua() {
	fmt.Println("VendorPriceCache = {")
	for _, item := range itemCache {
		fmt.Printf("  [\"%d\"] = %d,\n", item.Id, item.SellPrice)
	}
	fmt.Println("}")
}
