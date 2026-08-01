package transmog

import (
	"fmt"
	"log"
	"slices"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/wowapi"
)

const (
	persistName = "appearances"
)

var (
	appearanceIDsOwned         = map[int64]bool{}
	appearanceSetAppearanceIDs = persist.New[int64, bool](persistName)
)

// getAppearanceSetAppearanceIDs returns all appearance IDs that are in any appearance set
func getAppearanceSetAppearanceIDs() {
	appearanceSetIDs := wowapi.ItemAppearanceSetsIndexIDs()
	count := len(appearanceSetIDs)
	for setID, setName := range appearanceSetIDs {
		fmt.Printf("%d\tAppearance set: %d   %s\n", count, setID, setName)
		count--
		for _, appearanceID := range wowapi.ItemAppearanceSetIDs(setID) {
			appearanceSetAppearanceIDs.Set(appearanceID, true)
		}
	}
}

// getAppearanceIDsOwned returns the appearance IDs I own
func getAppearanceIDsOwned() map[int64]bool {
	myAppearanceIDs := map[int64]bool{}

	t, ok := wowapi.CollectionsTransmogs()
	if !ok {
		log.Fatal("Unable to obtain transmogs owned.")
	}

	transmogs := t.(map[string]any)

	for _, slot := range transmogs["slots"].([]any) {
		slot := slot.(map[string]any)
		for _, appearance := range slot["appearances"].([]any) {
			appearance := appearance.(map[string]any)
			id := web.ToInt64(appearance["id"])
			myAppearanceIDs[id] = true
		}
	}

	return myAppearanceIDs
}

func Init() {
	err := appearanceSetAppearanceIDs.Load()
	if err != nil {
		fmt.Printf("*** error opening appearances persist, creating new one: %v\n", err)
		getAppearanceSetAppearanceIDs()
		err = appearanceSetAppearanceIDs.Save()
		if err != nil {
			log.Fatalf("Failed to save appearances persist: %v\n", err)
		}
	}
	fmt.Printf("-- #Appearances persisted: %d\n", appearanceSetAppearanceIDs.Len())

	appearanceIDsOwned = getAppearanceIDsOwned()
	fmt.Printf("-- #Appearances owned    : %d\n", len(appearanceIDsOwned))
}

// flaky appearance IDs; WoW says I own them, but this app thinks I don't
var flaky = map[int64]bool{
	// These are not real appearances; they generate false positives
	573:   true, // Various equippable profession items
	577:   true, // Various equippable profession items
	870:   true, // Ammo
	1884:  true, // Various fish held in offhand
	2016:  true, // Various fish held in offhand
	2019:  true, // Various fish held in offhand
	70361: true, // Elegant Artisan's Cooking Hat
	78217: true, // Elegant Artisan's Fishing Hat

	// NOT part of an appearance set (so less interesting)
	1172:  true, // Ghostly Bracers
	22334: true, // Shrediron's Shredder
	22335: true, // Shrediron's Shredder
	22392: true, // Shadowtome
	22547: true, // Shadowtome
	22750: true, // Truesteel Waistguard
	22902: true, // Hexweave Cowl
	22905: true, // Hexweave Mantle
	22911: true, // Hexweave Cowl
	22914: true, // Hexweave Mantle
	22939: true, // Steelforged Saber
	22940: true, // Steelforged Saber
	23254: true, // Truesteel Waistguard
	56701: true, // Choral Hood
	56702: true, // Choral Amice
	56703: true, // Choral Vestments
	56704: true, // Choral Sash
	56705: true, // Choral Leggings
	56706: true, // Choral Slippers
	56707: true, // Choral Wraps
	56708: true, // Choral Handwraps
	56859: true, // Staccato Helm
	56860: true, // Staccato Mantle
	56861: true, // Staccato Vest
	56862: true, // Staccato Belt
	56863: true, // Staccato Leggings
	56864: true, // Staccato Boots
	56865: true, // Staccato Cuffs
	56866: true, // Staccato Grips
	57167: true, // Harmonium Helm
	57168: true, // Harmonium Spaulders
	57169: true, // Harmonium Breastplate
	57170: true, // Harmonium Girdle
	57171: true, // Harmonium Legplates
	57172: true, // Harmonium Percussive Stompers
	57173: true, // Harmonium Vambrace
	57174: true, // Harmonium Gauntlets
	57175: true, // Antecedent Drape
	57224: true, // Anthemic Coif
	57225: true, // Anthemic Shoulders
	57226: true, // Anthemic Cuirass
	57227: true, // Anthemic Links
	57228: true, // Anthemic Legguards
	57230: true, // Anthemic Bracers
	57231: true, // Anthemic Gauntlets
	78230: true, // Scepter of Spectacle: Order

	// NOT part of an appearance set  (so less interesting) [Horde only]
	37116: true, // Enchanter's Sorcerous Scepter

	// Part of an appearance set (so of great interest), but rarely available
	//18561: true, // Fists of Lightning
	//18575: true, // Nightfire Robe
	//18715: true, // Greyshadow Gloves
	//22757: true, // Truesteel Armguards
	//22906: true, // Hexweave Bracers
	//22915: true, // Hexweave Bracers
	//23247: true, // Truesteel Armguards
	//24178: true, // {Brilliant, Nimble, Powerful} Burnished Cloak
	//24180: true, // {Brilliant, Nimble, Powerful} Burnished Cloak
	32066: true, // Fashionable Autumn Cloak
	32237: true, // Aristocrat's Winter Drape
	//33357: true, // Sash of the Unredeemed
	//33365: true, // Sash of the Unredeemed
	33423: true, // Treads of Panicked Escape
	33439: true, // Treads of Panicked Escape
	33496: true, // Cord of Pilfered Rosaries
	33497: true, // Treads of Violent Intrusion
	//33716: true, // Moon-Wrought Clasp
	//34314: true, // Pristine Moon-Wrought Clasp
	//34558: true, // Cuffs of the Viridian Flameweavers
	//34870: true, // Gloves of Abhorrent Strategies
	//34886: true, // Gloves of Abhorrent Strategies
	35092: true, // Wristguards of Ominous Forging
	35101: true, // Wristguards of Ominous Forging
	//38275: true, // Reinforced Test Subject Shackles
	//38291: true, // Reinforced Test Subject Shackles
	38359: true, // Bloody Experimenter's Wraps
	38409: true, // Crushproof Vambraces
	38830: true, // Cord of Zandalari Resolve
	39969: true, // Gauntlets of Crashing Tides
	39976: true, // Gauntlets of Crashing Tides
	39980: true, // Brutish Myrmidon's Vambraces
	39983: true, // Brutish Myrmidon's Vambraces
	39987: true, // Gauntlets of Crashing Tides
	40325: true, // Cloak of Blessed Depths
	40811: true, // Belt of Concealed Intent
	40813: true, // Belt of Concealed Intent
	40883: true, // Footpads of Terrible Delusions
	40886: true, // Footpads of Terrible Delusions
	40967: true, // Gauntlets of Nightmare Manifest
	40970: true, // Gauntlets of Nightmare Manifest
	44135: true, // Bindings of the Subjugated
	44178: true, // Bindings of the Subjugated
	//57228: true, // Anthemic Legguards
	//80187: true, // Skyless Coif
	//80188: true, // Skyless Epaulets

}

// needAppearanceID returns true if I need this appearance ID
func needAppearanceID(id int64) bool {
	if flaky[id] {
		return false
	}
	if !appearanceIDsOwned[id] {
		fmt.Println("NEED APPEARANCE ID: ", id)
	}
	return !appearanceIDsOwned[id]
}

// NeedAppearance returns true if I need any of these appearance IDs
func NeedAppearance(appearanceIDs []int64) bool {
	return slices.ContainsFunc(appearanceIDs, needAppearanceID)
}

// InAppearanceSet returns true if any of these appearance IDs are in an appearance set
func InAppearanceSet(appearanceIDs []int64) bool {
	for _, appearanceID := range appearanceIDs {
		inSet, ok := appearanceSetAppearanceIDs.Get(appearanceID)
		if !ok {
			continue
		}
		if inSet {
			return true
		}
	}
	return false
}
