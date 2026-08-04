package battlepet

import (
	"fmt"
	"strings"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

// TODO: Convert this to a singleton package (like appearanceset.go)

type BattlePet struct {
	petNames  map[int64]string
	petsOwned map[int64]int64
}

const (
	PetCageItemID = int64(82800)
)

// getPetNames downloads all pet names from the WoW web API
func getPetNames() (map[int64]string, error) {
	petNames := map[int64]string{}

	allPets, err := wowapi.Pets()
	if err != nil {
		return nil, fmt.Errorf("unable to obtain battle pet names: %s", err)
	}

	for _, petRaw := range allPets {
		pet := petRaw.(map[string]any)
		id := web.ToInt64(pet["id"])
		petNames[id] = pet["name"].(string)
	}

	return petNames, nil
}

// getPetsOwned downloads the list of pets I own from the WoW web API
func getPetsOwned() (map[int64]int64, error) {
	myPets := map[int64]int64{}

	pets, err := wowapi.CollectionsPets()
	if err != nil {
		return nil, fmt.Errorf("unable to obtain battle pets owned: %s", err)
	}

	for _, petRaw := range pets {
		pet := petRaw.(map[string]any)

		species, ok := pet["species"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("unable to obtain battle pet species")
		}
		speciesID := web.ToInt64(species["id"])

		myPets[speciesID]++
	}

	return myPets, nil
}

func New() (*BattlePet, error) {
	bp := BattlePet{}
	var err error

	bp.petNames, err = getPetNames()
	if err != nil {
		return nil, err
	}

	bp.petsOwned, err = getPetsOwned()
	if err != nil {
		return nil, err
	}

	// Technically, this is _unique_ battle pets owned, but I don't keep dupes so it still works
	fmt.Printf("-- #Battle pets owned: %d/%d\n", len(bp.petsOwned), len(bp.petNames))

	return &bp, nil
}

// PetSpell returns true and the corresponding pet ID if the item is a pet summoning spell
func (bp *BattlePet) PetSpell(i wowitem.Item) (int64, bool) {
	if i.ItemSubclassName() != "Companion Pets" {
		return 0, false
	}

	for petID, petName := range bp.petNames {
		if petName == i.Name() {
			return petID, true
		}
	}

	return 0, false
}

// Name returns the pet name for the given ID
func (bp *BattlePet) Name(petID int64) string {
	return bp.petNames[petID]
}

// Owned returns true if I own this pet ID
func (bp *BattlePet) Owned(petID int64) bool {
	return bp.petsOwned[petID] > 0
}

// Output returns all petID/names
func (bp *BattlePet) Output() string {
	var output strings.Builder

	for petID, petName := range bp.petNames {
		// Output in a format pastable into shopping.SkipPets
		output.WriteString(fmt.Sprintf("%d: {}, // %s\n", petID, petName))
	}

	return output.String()
}
