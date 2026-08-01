package transmog

import (
	"fmt"
	"log"
	"slices"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/config"
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

// needAppearanceID returns true if I need this appearance ID
func needAppearanceID(id int64) bool {
	if config.FlakyAppearanceIDs[id] {
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
