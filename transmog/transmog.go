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

func Init(includeOwned bool) {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	load()
	fmt.Printf("-- #Appearance set cache: %d\n", len(allSetIds))

	if !includeOwned {
		return
	}

	allOwned = owned()
	fmt.Printf("-- #Transmogs: %d/%d\n", len(allOwned), 44344)
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
	// These are not real appearances
	573:  true, // Various equippable profession items
	577:  true, // Various equippable profession items
	870:  true, // Ammo
	1884: true, // Various fish held in offhand
	2016: true, // Various fish held in offhand
	2019: true, // Various fish held in offhand

	// NOT part of an appearance set
	5133:  true, // Round Buckler
	22392: true, // Shadowtome
	22547: true, // Shadowtome
	22902: true, // Hexweave Cowl
	22905: true, // Hexweave Mantle
	22911: true, // Hexweave Cowl
	22914: true, // Hexweave Mantle
	22939: true, // Steelforged Saber
	22940: true, // Steelforged Saber
	56703: true, // Choral Vestments
	56704: true, // Choral Sash
	56705: true, // Choral Leggings
	56707: true, // Choral Wraps
	56859: true, // Staccato Helm
	56860: true, // Staccato Mantle
	56861: true, // Staccato Vest
	56862: true, // Staccato Belt
	56865: true, // Staccato Cuffs
	56866: true, // Staccato Grips
	57169: true, // Harmonium Breastplate
	57173: true, // Harmonium Vambrace
	57174: true, // Harmonium Gauntlets
	57175: true, // Antecedent Drape
	57225: true, // Anthemic Shoulders
	57227: true, // Anthemic Links
	57230: true, // Anthemic Bracers
	57231: true, // Anthemic Gauntlets
	78230: true, // Scepter of Spectacle: Order

	// Part of an appearance set, but rarely available
	18561: true, // Fists of Lightning
	18575: true, // Nightfire Robe
	18715: true, // Greyshadow Gloves
	22757: true, // Truesteel Armguards
	22906: true, // Hexweave Bracers
	22915: true, // Hexweave Bracers
	23247: true, // Truesteel Armguards
	24178: true, // {Brilliant, Nimble, Powerful} Burnished Cloak
	24180: true, // {Brilliant, Nimble, Powerful} Burnished Cloak
	26016: true, // Cursed Demonchain Belt
	31863: true, // Vintage Duskwatch Cinch
	31934: true, // Mana-Cord of Deception
	32066: true, // Fashionable Autumn Cloak
	32237: true, // Aristocrat's Winter Drape
	33357: true, // Sash of the Unredeemed
	33365: true, // Sash of the Unredeemed
	33423: true, // Treads of Panicked Escape
	33439: true, // Treads of Panicked Escape
	33497: true, // Treads of Violent Intrusion
	33716: true, // Moon-Wrought Clasp
	34314: true, // Pristine Moon-Wrought Clasp
	34558: true, // Cuffs of the Viridian Flameweavers
	35092: true, // Wristguards of Ominous Forging
	35101: true, // Wristguards of Ominous Forging
	38275: true, // Reinforced Test Subject Shackles
	38291: true, // Reinforced Test Subject Shackles
	38325: true, // Antiseptic Specimen Handlers
	38359: true, // Bloody Experimenter's Wraps
	38409: true, // Crushproof Vambraces
	38830: true, // Cord of Zandalari Resolve
	39969: true, // Gauntlets of Crashing Tides
	39976: true, // Gauntlets of Crashing Tides
	39987: true, // Gauntlets of Crashing Tides
	40811: true, // Belt of Concealed Intent
	40813: true, // Belt of Concealed Intent
	40967: true, // Gauntlets of Nightmare Manifest
	40970: true, // Gauntlets of Nightmare Manifest
	57228: true, // Anthemic Legguards
	80187: true, // Skyless Coif
	80188: true, // Skyless Epaulets
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
	if !allOwned[id] {
		fmt.Println("NEED APPEARANCE ID: ", id)
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
		Init(false)
	}
	for _, appearance := range appearances {
		if allSetIds[appearance] {
			return true
		}
	}
	return false
}
