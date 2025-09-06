package transmog

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/item"
	"github.com/erikbryant/wow/itemCache"
	"github.com/erikbryant/wow/wowAPI"
)

var (
	allOwned            = map[int64]bool{}
	appearanceCacheFile = "./generated/appearanceCache.gob"
	allSetIds           = map[int64]bool{}
)

func Init(oauthAvailable bool) {
	if !oauthAvailable {
		return
	}

	allOwned = owned()
	fmt.Printf("-- #Transmogs: %d/%d\n", len(allOwned), 44344)
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	load()
	fmt.Printf("-- #Appearance set cache: %d\n", len(allSetIds))
}

// load loads the disk cache file into memory
func load() {
	file, err := os.Open(appearanceCacheFile)
	if err != nil {
		fmt.Printf("*** error opening appearance cache file: %v, creating new one\n", err)
		allItemAppearanceSetIds()
		fmt.Printf("Found %d appearance set IDs\n", len(allSetIds))
		save()
		return
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&allSetIds)
	if err != nil {
		log.Fatalf("error reading itemCache: %v", err)
	}
}

// save writes the in-memory cache file to disk
func save() {
	file, err := os.Create(appearanceCacheFile)
	if err != nil {
		log.Fatalf("error creating appearance cache file: %v", err)
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(allSetIds)
	if err != nil {
		log.Fatalf("error encoding allSetIds: %v", err)
	}
}

// allItemAppearanceSetIds returns a map of all item IDs that are in appearance sets
func allItemAppearanceSetIds() {
	ids := wowAPI.ItemAppearanceSetsIndexIds()
	count := len(ids)
	for setId, setName := range ids {
		fmt.Printf("%d\tAppearance set: %d   %s\n", count, setId, setName)
		count--
		for _, id := range wowAPI.ItemAppearanceSetIds(setId) {
			//fmt.Printf("   Appearance: %d\n", id)
			allSetIds[id] = true
		}
	}
}

// owned returns the IDs of the transmogs I own
func owned() map[int64]bool {
	myTransmogs := map[int64]bool{}

	t, ok := wowAPI.CollectionsTransmogs()
	if !ok {
		log.Fatal("ERROR: Unable to obtain transmogs owned.")
	}

	transmogs := t.(map[string]interface{})

	// Appearance sets
	for _, appearanceSet := range transmogs["appearance_sets"].([]interface{}) {
		appearanceSet := appearanceSet.(map[string]interface{})
		id := web.ToInt64(appearanceSet["id"])
		myTransmogs[id] = true
	}

	//	"slots": [
	//	{
	//		"slot": {
	//			"type": "HEAD",
	//			"name": "Head"
	//		},
	//		"appearances": [
	//		{
	//			"key": {
	//				"href": "https://us.api.blizzard.com/data/wow/item-appearance/358?namespace=static-11.1.5_60179-us"
	//			},
	//			"id": 358
	//		},
	//		{
	//			"key": {
	//				"href": "https://us.api.blizzard.com/data/wow/item-appearance/476?namespace=static-11.1.5_60179-us"
	//			},
	//			"id": 476
	//		},
	//	},
	//	...
	//	]
	for _, slot := range transmogs["slots"].([]interface{}) {
		slot := slot.(map[string]interface{})
		for _, appearance := range slot["appearances"].([]interface{}) {
			appearance := appearance.(map[string]interface{})
			id := web.ToInt64(appearance["id"])
			myTransmogs[id] = true
		}
	}

	return myTransmogs
}

// NeedId returns true if I need this transmog
func NeedId(id int64) bool {
	if len(allOwned) == 0 {
		Init(true)
	}
	if id <= 0 {
		return false
	}
	return !allOwned[id]
}

// flaky item IDs; WoW says I own the transmogs, but this app says I don't
var flaky = map[int64]bool{
	// I don't think these have appearances
	itemCache.Search("15 Pound Salmon").Id():             true,
	itemCache.Search("18 Pound Salmon").Id():             true,
	itemCache.Search("22 Pound Salmon").Id():             true,
	itemCache.Search("25 Pound Salmon").Id():             true,
	itemCache.Search("29 Pound Salmon").Id():             true,
	itemCache.Search("17 Pound Catfish").Id():            true,
	itemCache.Search("19 Pound Catfish").Id():            true,
	itemCache.Search("22 Pound Catfish").Id():            true,
	itemCache.Search("26 Pound Catfish").Id():            true,
	itemCache.Search("32 Pound Catfish").Id():            true,
	itemCache.Search("70 Pound Mightfish").Id():          true,
	itemCache.Search("85 Pound Mightfish").Id():          true,
	itemCache.Search("92 Pound Mightfish").Id():          true,
	itemCache.Search("Arclight Spanner").Id():            true,
	itemCache.Search("Blacksmith Hammer").Id():           true,
	itemCache.Search("Brumdysla, Hammer of Vrorsk").Id(): true,
	itemCache.Search("Crafted Light Shot").Id():          true,
	itemCache.Search("Exploding Shot").Id():              true,
	itemCache.Search("Heavy Shot").Id():                  true,
	itemCache.Search("Light Shot").Id():                  true,
	itemCache.Search("Shatter Rounds").Id():              true,
	itemCache.Search("Solid Shot").Id():                  true,

	// These give false negatives
	itemCache.Search("Anthemic Bracers").Id():              true,
	itemCache.Search("Anthemic Coif").Id():                 true,
	itemCache.Search("Anthemic Gauntlets").Id():            true,
	itemCache.Search("Anthemic Links").Id():                true,
	itemCache.Search("Anthemic Shoulders").Id():            true,
	itemCache.Search("Choral Amice").Id():                  true,
	itemCache.Search("Choral Leggings").Id():               true,
	itemCache.Search("Choral Sash").Id():                   true,
	itemCache.Search("Choral Slippers").Id():               true,
	itemCache.Search("Choral Vestments").Id():              true,
	itemCache.Search("Choral Wraps").Id():                  true,
	itemCache.Search("Harmonium Gauntlets").Id():           true,
	itemCache.Search("Harmonium Percussive Stompers").Id(): true,
	itemCache.Search("Harmonium Spaulders").Id():           true,
	itemCache.Search("Harmonium Vambrace").Id():            true,
	itemCache.Search("Staccato Belt").Id():                 true,
	itemCache.Search("Staccato Helm").Id():                 true,
	itemCache.Search("Staccato Mantle").Id():               true,
	//itemCache.Search("").Id():                  true,
	//itemCache.Search("").Id():                  true,
}

// NeedItem returns true if I need any of the transmogs this item provides
func NeedItem(i item.Item) bool {
	if flaky[i.Id()] {
		return false
	}

	for _, id := range i.Appearances() {
		if NeedId(id) {
			return true
		}
	}
	return false
}

// InAppearanceSet returns true if this item is in an appearance set
func InAppearanceSet(i item.Item) bool {
	if len(allSetIds) == 0 {
		Init(true)
	}
	for _, id := range i.Appearances() {
		if allSetIds[id] {
			return true
		}
	}
	return false
}
