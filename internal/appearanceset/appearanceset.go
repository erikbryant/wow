package appearanceset

import (
	"fmt"
	"os"
	"sync"

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
	once sync.Once
	as   *AppearanceSets
)

// getAppearanceSetsAppearanceIDs loads all appearance IDs that are in any appearance set
func getAppearanceSetsAppearanceIDs() {
	appearanceSetsIDs := wowapi.ItemAppearanceSetsIndexIDs()
	count := len(appearanceSetsIDs)

	for setID, setName := range appearanceSetsIDs {
		fmt.Fprintf(os.Stderr, "%d\tLoading appearance set: %d   %s\n", count, setID, setName)
		count--
		for _, appearanceID := range wowapi.ItemAppearanceSetIDs(setID) {
			as.IDs.Set(appearanceID, true)
		}
	}
}

func createFromWeb() {
	getAppearanceSetsAppearanceIDs()
	err := as.IDs.Save()
	if err != nil {
		fmt.Fprintf(os.Stderr, "*** failed to save appearances persist: %s\n", err)
	}
}

func load() {
	as = &AppearanceSets{
		IDs: persist.New[int64, bool](persistName),
	}

	err := as.IDs.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "*** error opening appearances persist, creating new one: %v\n", err)
		createFromWeb()
	}

	fmt.Printf("-- #Appearances known: %d\n", as.IDs.Len())
}

func New() *AppearanceSets {
	once.Do(load)
	return as
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
