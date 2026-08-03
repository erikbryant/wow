package transmog

import (
	"fmt"
	"slices"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/userconfig"
	"github.com/erikbryant/wow/internal/wowapi"
)

type appearanceSetStore interface {
	Load() error
	Save() error
	Set(int64, bool)
	Get(int64) (bool, bool)
	Len() int
}

const (
	persistName = "appearances"
)

var (
	appearanceIDsOwned = map[int64]bool{}

	// Define external dependencies such that they can be mocked in the tests.
	appearanceSetAppearanceIDs appearanceSetStore = persist.New[int64, bool](persistName)

	collectionsTransmogs   = wowapi.CollectionsTransmogs
	appearanceSetsIndexIDs = wowapi.ItemAppearanceSetsIndexIDs
	appearanceSetIDs       = wowapi.ItemAppearanceSetIDs
	flakyAppearanceID      = userconfig.FlakyAppearanceID
)

// getAppearanceSetsAppearanceIDs returns all appearance IDs that are in any appearance set
func getAppearanceSetsAppearanceIDs() {
	appearanceSetsIDs := appearanceSetsIndexIDs()
	count := len(appearanceSetsIDs)
	for setID, setName := range appearanceSetsIDs {
		fmt.Printf("%d\tAppearance set: %d   %s\n", count, setID, setName)
		count--
		for _, appearanceID := range appearanceSetIDs(setID) {
			appearanceSetAppearanceIDs.Set(appearanceID, true)
		}
	}
}

// getAppearanceIDsOwned returns the appearance IDs I own
func getAppearanceIDsOwned() (map[int64]bool, error) {
	myAppearanceIDs := map[int64]bool{}

	t, ok := collectionsTransmogs()
	if !ok {
		return nil, fmt.Errorf("unable to obtain transmogs owned")
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

	return myAppearanceIDs, nil
}

func Init() error {
	err := appearanceSetAppearanceIDs.Load()
	if err != nil {
		fmt.Printf("*** error opening appearances persist, creating new one: %v\n", err)
		getAppearanceSetsAppearanceIDs()
		err = appearanceSetAppearanceIDs.Save()
		if err != nil {
			return fmt.Errorf("failed to save appearances persist: %s", err)
		}
	}

	appearanceIDsOwned, err = getAppearanceIDsOwned()
	if err != nil {
		return err
	}

	fmt.Printf("-- #Appearances owned: %d/%d\n", len(appearanceIDsOwned), appearanceSetAppearanceIDs.Len())

	return nil
}

// needAppearanceID returns true if I need this appearance ID
func needAppearanceID(id int64) bool {
	if flakyAppearanceID(id) {
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
