package appearanceset

import (
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/wowapi"
)

type Persistence struct {
	*persist.Persistence[int64, bool]
}

// NewFromWeb creates a new Persistence, populated with data from the WoW web API.
func NewFromWeb(persistencePath string) (*Persistence, error) {
	as := &Persistence{
		Persistence: persist.New[int64, bool](persistencePath + ".new"),
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

		appearanceIDs, err := wowapi.ItemAppearanceSetIDs(setID)
		if err != nil {
			return nil, err
		}

		for _, appearanceID := range appearanceIDs {
			as.Set(appearanceID, true)
		}
	}

	return as, nil
}

// New creates a new Persistence, populated with data from its persistence store.
func New(persistencePath string) (*Persistence, error) {
	as := &Persistence{
		Persistence: persist.New[int64, bool](persistencePath),
	}

	if err := as.Load(); err != nil {
		return nil, fmt.Errorf("failed to load appearance sets: %w", err)
	}

	fmt.Printf("-- #Appearances known: %d\n", as.Len())

	return as, nil
}

// Contains returns true if any of these appearance IDs are in an appearance set.
func (as *Persistence) Contains(appearanceIDs []int64) bool {
	for _, appearanceID := range appearanceIDs {
		inSet, ok := as.Get(appearanceID)
		if ok && inSet {
			return true
		}
	}

	return false
}
