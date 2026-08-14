package appearanceset

import (
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/wowapi"
)

type AppearanceSets struct {
}

const (
	persistName = "appearances"
)

var (
	asaIDs *persist.Persistence[int64, bool]
)

// getFromWeb loads all appearance IDs that are in any appearance set
func getFromWeb() error {
	// TODO: If this exits on error, asaIDs is partially set. Verify load is complete before assigning to asaIDs.

	appearanceSetsIDs, err := wowapi.ItemAppearanceSetsIndexIDs()
	if err != nil {
		return err
	}
	count := len(appearanceSetsIDs)

	for setID, setName := range appearanceSetsIDs {
		fmt.Fprintf(os.Stderr, "%d\tLoading appearance set: %d   %s\n", count, setID, setName)
		count--
		asIDs, err := wowapi.ItemAppearanceSetIDs(setID)
		if err != nil {
			return err
		}
		for _, appearanceID := range asIDs {
			asaIDs.Set(appearanceID, true)
		}
	}

	err = asaIDs.Save()
	if err != nil {

		return fmt.Errorf("failed to save appearances persist: %w", err)
	}

	return nil
}

func New() (*AppearanceSets, error) {
	if asaIDs == nil {
		asaIDs = persist.New[int64, bool](persistName)
	}

	if asaIDs.Loaded() {
		return &AppearanceSets{}, nil
	}

	err := asaIDs.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "*** error opening appearances persist, creating new one: %s\n", err)
		err = getFromWeb()
		if err != nil {
			return nil, err
		}
	}

	fmt.Printf("-- #Appearances known: %d\n", asaIDs.Len())

	return &AppearanceSets{}, nil
}

// Contains returns true if any of these appearance IDs are in an appearance set
func (as *AppearanceSets) Contains(appearanceIDs []int64) bool {
	for _, appearanceID := range appearanceIDs {
		inSet, ok := asaIDs.Get(appearanceID)
		if !ok {
			continue
		}
		if inSet {
			return true
		}
	}
	return false
}
