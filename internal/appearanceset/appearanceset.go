package appearanceset

import (
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/wowapi"
)

type AppearanceSets struct {
	IDs *persist.Persistence[int64, bool]
}

const (
	persistName = "appearances"
)

var (
	as = &AppearanceSets{
		IDs: persist.New[int64, bool](persistName),
	}
)

// getAppearanceSetsAppearanceIDs loads all appearance IDs that are in any appearance set
func getAppearanceSetsAppearanceIDs() error {
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
			as.IDs.Set(appearanceID, true)
		}
	}

	return nil
}

func createFromWeb() error {
	err := getAppearanceSetsAppearanceIDs()
	if err != nil {
		return fmt.Errorf("failed to get appearances: %s", err)
	}

	err = as.IDs.Save()
	if err != nil {

		return fmt.Errorf("failed to save appearances persist: %s", err)
	}

	return nil
}

func New() (*AppearanceSets, error) {
	if as.IDs.Loaded() {
		return as, nil
	}

	err := as.IDs.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "*** error opening appearances persist, creating new one: %s\n", err)
		err = createFromWeb()
		if err != nil {
			return nil, fmt.Errorf("failed to load or create appearances: %s", err)
		}
	}

	fmt.Printf("-- #Appearances known: %d\n", as.IDs.Len())

	return as, nil
}

// Contains returns true if any of these appearance IDs are in an appearance set
func (as *AppearanceSets) Contains(appearanceIDs []int64) bool {
	for _, appearanceID := range appearanceIDs {
		inSet, ok := as.IDs.Get(appearanceID)
		if !ok {
			continue
		}
		if inSet {
			return true
		}
	}
	return false
}
