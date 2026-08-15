package appearanceset

import (
	"fmt"

	"github.com/erikbryant/wow/internal/persist"
)

type AppearanceSets struct {
	asaIDs *persist.Persistence[int64, bool]
}

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
