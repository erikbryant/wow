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

// getPetNames downloads all pet names from the WoW web API
func getPetNames() (map[int64]string, error) {
	names := map[int64]string{}

	allPets, err := wowapi.Pets()
	if err != nil {
		return nil, fmt.Errorf("unable to obtain battle pet names: %w", err)
	}

	for _, petRaw := range allPets {
		pet := petRaw.(map[string]any)
		id := web.ToInt64(pet["id"])
		names[id] = pet["name"].(string)
	}

	return names, nil
}

// getPetsOwned downloads the list of pets I own from the WoW web API
func getPetsOwned() (map[int64]int64, error) {
	owned := map[int64]int64{}

	pets, err := wowapi.CollectionsPets()
	if err != nil {
		return nil, fmt.Errorf("unable to obtain battle pets owned: %w", err)
	}

	for _, petRaw := range pets {
		pet := petRaw.(map[string]any)

		species, ok := pet["species"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("unable to obtain battle pet species")
		}
		speciesID := web.ToInt64(species["id"])

		owned[speciesID]++
	}

	return owned, nil
}

func New() (*BattlePet, error) {
	var err error
	bp := BattlePet{}

	bp.names, err = getPetNames()
	if err != nil {
		return nil, err
	}

	bp.owned, err = getPetsOwned()
	if err != nil {
		return nil, err
	}

	fmt.Printf("-- #Battle pet species owned: %d/%d\n", len(bp.owned), len(bp.names))

	return &bp, nil
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
