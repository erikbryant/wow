package transmog

import (
	"encoding/gob"
	"fmt"
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/item"
	"github.com/erikbryant/wow/wowAPI"
	"log"
	"os"
)

var (
	allOwned            = map[int64]bool{}
	appearanceCacheFile = "./generated/appearanceCache.gob"
	allSetIds           = map[int64]bool{}
)

func Init() {
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
			fmt.Printf("   Appearance: %d\n", id)
			allSetIds[id] = true
		}
	}
}

// appearances returns a list of all item appearance IDs
func appearances() map[int64]bool {
	ids := map[int64]bool{}

	for _, slot := range wowAPI.ItemAppearanceSlotIndex() {
		appearances, ok := wowAPI.ItemAppearanceSlot(slot)
		if !ok {
			log.Fatal("ERROR: Unable to obtain appearances for slot:", slot)
		}
		if appearances == nil {
			log.Fatal("ERROR: no appearances for slot:", slot)
		}
		for _, appearance := range appearances {
			id := web.ToInt64(appearance.(map[string]interface{})["id"])
			ids[id] = true
		}
	}

	return ids
}

// ItemIdsForAppearance returns a list of item IDs that have the given appearance
func ItemIdsForAppearance(appearanceId int64) ([]int64, bool) {
	ids := []int64{}

	appearance, ok := wowAPI.ItemAppearance(appearanceId)
	if !ok {
		return nil, false
	}

	items := appearance.(map[string]interface{})["items"].([]interface{})

	for _, i := range items {
		id := web.ToInt64(i.(map[string]interface{})["id"])
		ids = append(ids, id)
	}

	return ids, true
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
		Init()
	}
	if id <= 0 {
		return false
	}
	return !allOwned[id]
}

// NeedItem returns true if I need any of the transmogs this item provides
func NeedItem(i item.Item) bool {
	// WoW says I have these transmogs, but this app says I don't. Ignore them.
	flaky := map[int64]bool{
		2516:   true, // Light Shot
		3033:   true, // Solid Shot
		3465:   true, // Exploding Shot
		5956:   true, // Blacksmith Hammer
		6219:   true, // Arclight Spanner
		6309:   true, // 17 Pound Catfish
		6310:   true, // 19 Pound Catfish
		6311:   true, // 22 Pound Catfish
		13901:  true, // 15 Pound Salmon
		13902:  true, // 18 Pound Salmon
		52020:  true, // Shatter Rounds
		144405: true, // Waistclasp of Unethical Power
		188007: true, // Choral Slippers
		188011: true, // Choral Sash
		188017: true, // Staccato Belt
		188019: true, // Anthemic Cuirass
		188021: true, // Anthemic Gauntlets
		188026: true, // Anthemic Bracers
		188037: true, // Choral Armice
	}

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
		Init()
	}
	for _, id := range i.Appearances() {
		if allSetIds[id] {
			return true
		}
	}
	return false
}
