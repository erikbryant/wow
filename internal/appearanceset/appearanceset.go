package appearanceset

import (
	"fmt"

	"github.com/erikbryant/wow/internal/persist"
)

type AppearanceSets struct {
}

var (
	asaIDs *persist.Persistence[int64, bool]
)

func New(persistencePath string) (*AppearanceSets, error) {
	if asaIDs == nil {
		asaIDs = persist.New[int64, bool](persistencePath)
	}

	if asaIDs.Loaded() {
		return &AppearanceSets{}, nil
	}

	err := asaIDs.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load appearance sets: %w", err)
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
