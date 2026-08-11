package battlepet

import (
	"fmt"
	"strings"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

type BattlePet struct {
	names map[int64]string
	owned map[int64]int64
}

const (
	PetCageItemID = int64(82800)
)

var (
	bp = &BattlePet{}
)

// getPetNames downloads all pet names from the WoW web API
func getPetNames() error {
	if bp.names != nil {
		return nil
	}

	bp.names = map[int64]string{}

	allPets, err := wowapi.Pets()
	if err != nil {
		return fmt.Errorf("unable to obtain battle pet names: %s", err)
	}

	for _, petRaw := range allPets {
		pet := petRaw.(map[string]any)
		id := web.ToInt64(pet["id"])
		bp.names[id] = pet["name"].(string)
	}

	return nil
}

// getPetsOwned downloads the list of pets I own from the WoW web API
func getPetsOwned() error {
	if bp.owned != nil {
		return nil
	}

	bp.owned = map[int64]int64{}

	pets, err := wowapi.CollectionsPets()
	if err != nil {
		return fmt.Errorf("unable to obtain battle pets owned: %s", err)
	}

	for _, petRaw := range pets {
		pet := petRaw.(map[string]any)

		species, ok := pet["species"].(map[string]any)
		if !ok {
			return fmt.Errorf("unable to obtain battle pet species")
		}
		speciesID := web.ToInt64(species["id"])

		bp.owned[speciesID]++
	}

	return nil
}

func New() (*BattlePet, error) {
	var err error

	err = getPetNames()
	if err != nil {
		return nil, err
	}

	err = getPetsOwned()
	if err != nil {
		return nil, err
	}

	// Technically, this is _unique_ battle pets owned, but I don't keep dupes so it still works
	fmt.Printf("-- #Battle pets owned: %d/%d\n", len(bp.owned), len(bp.names))

	return bp, nil
}

// PetSpell returns true and the corresponding pet ID if the item is a pet summoning spell
func (bp *BattlePet) PetSpell(i wowitem.Item) (int64, bool) {
	if i.ItemSubclassName() != "Companion Pets" {
		return 0, false
	}

	for petID, petName := range bp.names {
		if petName == i.Name() {
			return petID, true
		}
	}

	return 0, false
}

// Name returns the pet name for the given ID
func (bp *BattlePet) Name(petID int64) string {
	return bp.names[petID]
}

// Owned returns true if I own this pet ID
func (bp *BattlePet) Owned(petID int64) bool {
	return bp.owned[petID] > 0
}

// Output returns all petID/names
func (bp *BattlePet) Output() string {
	var output strings.Builder

	for petID, petName := range bp.names {
		// Output in a format pastable into shopping.SkipPets
		output.WriteString(fmt.Sprintf("%d: {}, // %s\n", petID, petName))
	}

	return output.String()
}
