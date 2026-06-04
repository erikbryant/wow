package transmog

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"

	"github.com/erikbryant/web"
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

// NeedId returns true if I need this transmog appearance ID
func NeedId(id int64) bool {
	if id <= 0 {
		return false
	}
	if id == 573 || id == 577 {
		// Various equippable profession items
		return false
	}
	if id == 870 {
		// Ammo
		return false
	}
	if id == 2016 {
		// Various offhand fish
		return false
	}
	if len(allOwned) == 0 {
		Init(true)
	}
	return !allOwned[id]
}

// flaky item IDs; WoW says I own the transmogs, but this app says I don't
var flaky = map[int64]bool{
	// Need these, but almost never available -- part of an appearance set
	//itemCache.Search("Aristocrat's Winter Drape").Id():   true,
	//itemCache.Search("Bloody Experimenter's Wraps").Id(): true,
	//itemCache.Search("Cord of Zandalari Resolve").Id():   true,
	//itemCache.Search("Fashionable Autumn Cloak").Id():    true,
	//itemCache.Search("Mana-Cord of Deception").Id():      true,
	//itemCache.Search("Skyless Coif").Id():                true,
	//itemCache.Search("Skyless Epaulets").Id():            true,
	//itemCache.Search("Vintage Duskwatch Cinch").Id():     true,

	// Need these, but almost never available -- NOT part of an appearance set
	//itemCache.Search("Anthemic Shoulders").Id():    true,
	//itemCache.Search("Choral Handwraps").Id():      true,
	//itemCache.Search("Choral Leggings").Id():       true,
	//itemCache.Search("Choral Slippers").Id():       true,
	//itemCache.Search("Choral Wraps").Id():          true,
	//itemCache.Search("Harmonium Breastplate").Id(): true,
	//itemCache.Search("Harmonium Gauntlets").Id():   true,
	//itemCache.Search("Harmonium Girdle").Id():      true,
	//itemCache.Search("Protective Gloves").Id():     true,
	//itemCache.Search("Round Buckler").Id():         true,
	//itemCache.Search("Staccato Belt").Id():         true,
	//itemCache.Search("Staccato Cuffs").Id():        true,
	//itemCache.Search("Staccato Grips").Id():        true,

	//itemCache.Search("").Id():                  true,
	//itemCache.Search("").Id():                  true,
}

// NeedAppearance returns true if I need any of these appearance IDs
func NeedAppearance(appearances []int64) bool {
	for _, appearance := range appearances {
		if NeedId(appearance) {
			return true
		}
	}
	return false
}

// InAppearanceSet returns true if any of these appearance IDs are in an appearance set
func InAppearanceSet(appearances []int64) bool {
	if len(allSetIds) == 0 {
		Init(true)
	}
	for _, appearance := range appearances {
		if allSetIds[appearance] {
			return true
		}
	}
	return false
}
