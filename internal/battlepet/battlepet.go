package battlepet

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/common"
	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

const (
	persistName   = "battlePets"
	PetCageItemID = int64(82800)
)

var (
	petNames  = persist.New[int64, string](persistName)
	petsOwned = map[int64][]wowitem.PetInfo{}
)

// refreshPetNames downloads all pet names from the WoW web API
func refreshPetNames() {
	allPets, ok := wowapi.Pets()
	if !ok {
		log.Fatal("Unable to obtain pet names.")
	}

	for _, petRaw := range allPets {
		pet := petRaw.(map[string]any)
		id := web.ToInt64(pet["id"])
		petNames.Set(id, pet["name"].(string))
	}
}

// getPetsOwned downloads the list of pets I own from the WoW web API
func getPetsOwned() map[int64][]wowitem.PetInfo {
	myPets := map[int64][]wowitem.PetInfo{}

	pets, ok := wowapi.CollectionsPets()
	if !ok {
		log.Fatal("Unable to obtain pets owned.")
	}

	for _, petRaw := range pets {
		pet := petRaw.(map[string]any)

		var p wowitem.PetInfo

		stats, ok := pet["stats"].(map[string]any)
		if !ok {
			log.Fatal("Unable to obtain stats.")
		}
		p.BreedID = web.ToInt64(stats["breed_id"])

		p.Level = web.ToInt64(pet["level"])

		quality, ok := pet["quality"].(map[string]any)
		if !ok {
			log.Fatal("Unable to obtain quality.")
		}
		p.QualityID = common.QualityID(quality["name"].(string))

		species, ok := pet["species"].(map[string]any)
		if !ok {
			log.Fatal("Unable to obtain species.")
		}
		p.SpeciesID = web.ToInt64(species["id"])

		_, ok = myPets[p.SpeciesID]
		if ok {
			name, _ := pet["species"].(map[string]any)
			fmt.Println("Duplicate pet:", name)
		} else {
			myPets[p.SpeciesID] = []wowitem.PetInfo{}
		}
		myPets[p.SpeciesID] = append(myPets[p.SpeciesID], p)
	}

	return myPets
}

func Init() {
	err := petNames.Load()
	if err != nil {
		fmt.Printf("*** error opening pet name persist, creating new one: %v\n", err)
		refreshPetNames()
		petNames.Save()
	}
	petsOwned = getPetsOwned()
	fmt.Printf("-- #Battle pets owned    : %d/%d\n", len(petsOwned), petNames.Len())
}

// PetSpell returns true and the corresponding pet ID if the item is a pet summoning spell
func PetSpell(i wowitem.Item) (int64, bool) {
	if i.ItemSubclassName() != "Companion Pets" {
		return 0, false
	}
	key, _, ok := petNames.Search(func(v string) bool { return v == i.Name() })
	return key, ok
}

// Name returns the pet name for the given ID
func Name(petID int64) string {
	name, ok := petNames.Get(petID)
	if !ok {
		log.Fatal("Pet not found:", petID)
	}
	return name
}

// Owned returns true if I own this pet ID
func Owned(petID int64) bool {
	return len(petsOwned[petID]) > 0
}

// Output returns all petID/names
func Output() string {
	var output strings.Builder

	keys := petNames.Keys()
	slices.Sort(keys)

	for _, petID := range keys {
		name, ok := petNames.Get(petID)
		if !ok {
			continue
		}
		output.WriteString(fmt.Sprintf("%d %s\n", petID, name))
	}

	return output.String()
}
