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

type Transmog struct {
	appearanceIDsOwned         map[int64]bool
	appearanceSetAppearanceIDs appearanceSetStore
}

const (
	persistName = "appearances"
)

var (
	// Define external dependencies such that they can be mocked in the tests.
	collectionsTransmogs   = wowapi.CollectionsTransmogs
	appearanceSetsIndexIDs = wowapi.ItemAppearanceSetsIndexIDs
	appearanceSetIDs       = wowapi.ItemAppearanceSetIDs
	flakyAppearanceID      = userconfig.FlakyAppearanceID
)

// getAppearanceSetsAppearanceIDs loads all appearance IDs that are in any appearance set
func (t *Transmog) getAppearanceSetsAppearanceIDs() {
	appearanceSetsIDs := appearanceSetsIndexIDs()
	count := len(appearanceSetsIDs)
	for setID, setName := range appearanceSetsIDs {
		fmt.Printf("%d\tAppearance set: %d   %s\n", count, setID, setName)
		count--
		for _, appearanceID := range appearanceSetIDs(setID) {
			t.appearanceSetAppearanceIDs.Set(appearanceID, true)
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

func New(useWeb bool) (*Transmog, error) {
	t := Transmog{
		appearanceSetAppearanceIDs: persist.New[int64, bool](persistName),
	}

	err := t.appearanceSetAppearanceIDs.Load()
	if err != nil {
		fmt.Printf("*** error opening appearances persist, creating new one: %v\n", err)
		t.getAppearanceSetsAppearanceIDs()
		err = t.appearanceSetAppearanceIDs.Save()
		if err != nil {
			return nil, fmt.Errorf("failed to save appearances persist: %s", err)
		}
	}

	if !useWeb {
		return &t, nil
	}

	t.appearanceIDsOwned, err = getAppearanceIDsOwned()
	if err != nil {
		return nil, err
	}

	fmt.Printf("-- #Appearances owned: %d/%d\n", len(t.appearanceIDsOwned), t.appearanceSetAppearanceIDs.Len())

	return &t, nil
}

// needAppearanceID returns true if I need this appearance ID
func (t *Transmog) needAppearanceID(id int64) bool {
	if flakyAppearanceID(id) {
		return false
	}
	if !t.appearanceIDsOwned[id] {
		fmt.Println("NEED APPEARANCE ID: ", id)
	}
	return !t.appearanceIDsOwned[id]
}

// NeedAppearance returns true if I need any of these appearance IDs
func (t *Transmog) NeedAppearance(appearanceIDs []int64) bool {
	return slices.ContainsFunc(appearanceIDs, t.needAppearanceID)
}

// InAppearanceSet returns true if any of these appearance IDs are in an appearance set
func (t *Transmog) InAppearanceSet(appearanceIDs []int64) bool {
	for _, appearanceID := range appearanceIDs {
		inSet, ok := t.appearanceSetAppearanceIDs.Get(appearanceID)
		if !ok {
			continue
		}
		if inSet {
			return true
		}
	}
	return false
}
