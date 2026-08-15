package appearanceset

import (
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/wowapi"
)

type AppearanceSets struct {
	asaIDs *persist.Persistence[int64, bool]
}

// NewFromWeb returns an object populated from the web API
func NewFromWeb(persistencePath string) (*AppearanceSets, error) {
	as := AppearanceSets{
		asaIDs: persist.New[int64, bool](persistencePath + ".new"),
	}

	appearanceSetsIDs, err := wowapi.ItemAppearanceSetsIndexIDs()
	if err != nil {
		return nil, err
	}

	total := len(appearanceSetsIDs)
	count := 1

	for setID, setName := range appearanceSetsIDs {
		fmt.Fprintf(os.Stderr, "Loading appearance set %4d/%4d: %5d  %s\n", count, total, setID, setName)
		count++
		asIDs, err := wowapi.ItemAppearanceSetIDs(setID)
		if err != nil {
			return nil, err
		}
		for _, appearanceID := range asIDs {
			as.asaIDs.Set(appearanceID, true)
		}
	}

	return &as, nil
}

// New returns an object populated from the persistence
func New(persistencePath string) (*AppearanceSets, error) {
	as := AppearanceSets{
		asaIDs: persist.New[int64, bool](persistencePath),
	}

	err := as.asaIDs.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load appearance sets: %w", err)
	}

	fmt.Printf("-- #Appearances known: %d\n", as.asaIDs.Len())

	return &as, nil
}

// Contains returns true if any of these appearance IDs are in an appearance set
func (as *AppearanceSets) Contains(appearanceIDs []int64) bool {
	for _, appearanceID := range appearanceIDs {
		inSet, ok := as.asaIDs.Get(appearanceID)
		if !ok {
			continue
		}
		if inSet {
			return true
		}
	}
	return false
}

func (as *AppearanceSets) Len() int {
	return as.asaIDs.Len()
}

func (as *AppearanceSets) Path() string {
	return as.asaIDs.Path()
}

func (as *AppearanceSets) Save() error {
	return as.asaIDs.Save()
}
