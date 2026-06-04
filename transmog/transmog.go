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

// flaky appearance IDs; WoW says I own the transmogs, but this app thinks I don't
var flaky = map[int64]bool{
	573:   true, // Various equippable profession items
	577:   true, // Various equippable profession items
	870:   true, // Ammo
	2016:  true, // Various fish held in offhand
	31863: true, // Vintage Duskwatch Cinch
	31934: true, // Mana-Cord of Deception
	38409: true, // Crushproof Vambraces
}

// NeedId returns true if I need this transmog appearance ID
func NeedId(id int64) bool {
	if id <= 0 {
		return false
	}
	if flaky[id] {
		return false
	}
	if len(allOwned) == 0 {
		Init(true)
	}
	return !allOwned[id]
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
