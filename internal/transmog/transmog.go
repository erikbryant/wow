package transmog

import (
	"fmt"
	"os"
	"sync"

	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/wowapi"
)

type Transmog struct {
	appearanceSetAppearanceIDs *persist.Persistence[int64, bool]
}

const (
	persistName = "appearances"
)

var (
	once sync.Once
	tr   *Transmog
)

// getAppearanceSetsAppearanceIDs loads all appearance IDs that are in any appearance set
func getAppearanceSetsAppearanceIDs() {
	appearanceSetsIDs := wowapi.ItemAppearanceSetsIndexIDs()
	count := len(appearanceSetsIDs)

	for setID, setName := range appearanceSetsIDs {
		fmt.Printf("%d\tLoading appearance set: %d   %s\n", count, setID, setName)
		count--
		for _, appearanceID := range wowapi.ItemAppearanceSetIDs(setID) {
			tr.appearanceSetAppearanceIDs.Set(appearanceID, true)
		}
	}
}

func createFromWeb() {
	getAppearanceSetsAppearanceIDs()
	err := tr.appearanceSetAppearanceIDs.Save()
	if err != nil {
		fmt.Fprintf(os.Stderr, "*** failed to save appearances persist: %s\n", err)
	}
}

func load() {
	tr = &Transmog{
		appearanceSetAppearanceIDs: persist.New[int64, bool](persistName),
	}

	err := tr.appearanceSetAppearanceIDs.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "*** error opening appearances persist, creating new one: %v\n", err)
		createFromWeb()
	}

	fmt.Printf("-- #Appearances known: %d\n", tr.appearanceSetAppearanceIDs.Len())
}

func New() *Transmog {
	once.Do(load)
	return tr
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
