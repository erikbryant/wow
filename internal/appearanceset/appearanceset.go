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

// NewEmpty creates a new Persistence with no items in it.
func NewEmpty(persistencePath string) *Persistence {
	return &Persistence{
		Persistence: persist.New[int64, bool](persistencePath + ".new"),
	}
}

// New creates a new Persistence, populated with data from its persistence store.
func New(persistencePath string) (*Persistence, error) {
	as := &Persistence{
		Persistence: persist.New[int64, bool](persistencePath),
	}

	if err := as.Load(); err != nil {
		return nil, fmt.Errorf("failed to load appearance sets: %w", err)
	}

	return as, nil
}

func (as *Persistence) LoadFromWeb(client *wowapi.Client) error {
	appearanceSetsIDs, err := client.ItemAppearanceSetsIndexIDs()
	if err != nil {
		return err
	}

	total := len(appearanceSetsIDs)
	count := 1

	for setID, setName := range appearanceSetsIDs {
		fmt.Fprintf(os.Stderr, "Loading appearance set %4d/%4d: %5d  %s\n", count, total, setID, setName)
		count++

		appearanceIDs, err := client.ItemAppearanceSetIDs(setID)
		if err != nil {
			return err
		}

		for _, appearanceID := range appearanceIDs {
			as.Set(appearanceID, true)
		}
	}

	return nil
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
