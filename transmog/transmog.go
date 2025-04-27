package transmog

import (
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/item"
	"github.com/erikbryant/wow/wowAPI"
	"log"
)

var (
	allTransmogs = map[int64]bool{}
	allOwned     = map[int64]bool{}
)

func Init() {
	allTransmogs = appearances()
	allOwned = owned()

	// I own some transmogs that are no longer in the API.
	// Delete them so they do not cause lookup errors.
	for id := range allOwned {
		if !allTransmogs[id] {
			delete(allOwned, id)
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

	// Problematic transmog IDs. Pretend we already own them.
	myTransmogs[573] = true  // Blacksmith Hammer
	myTransmogs[577] = true  // Arclight Spanner, Shoni's Disarming Tool, Tork Wrench
	myTransmogs[2016] = true // {17,19,22,26,32} Pound Catfish, {15,18,22,25,29,32} Pound Salmon, OldCrafty

	return myTransmogs
}

// NeedId returns true if I need this transmog
func NeedId(id int64) bool {
	if len(allOwned) == 0 {
		log.Fatal("ERROR: You must call transmog.Init() before calling transmog.NeedId()")
	}
	if id <= 0 {
		return false
	}
	return !allOwned[id]
}

// NeedItem returns true if I need the transmog this item provides
func NeedItem(i item.Item) bool {
	for _, id := range i.Appearances() {
		if NeedId(id) {
			return true
		}
	}
	return false
}
