package battlepet

import (
	"fmt"
	"log"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/cache"
	"github.com/erikbryant/wow/internal/common"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

const (
	cacheFilename = "./data/petNameCache.gob"
	PetCageItemId = int64(82800)
)

var (
	allNames = cache.New(cacheFilename, map[int64]string{})
	allOwned = map[int64][]wowitem.PetInfo{}
)

// refreshPetNames downloads all pet names from the WoW web API
func refreshPetNames() map[int64]string {
	pets := map[int64]string{}

	allPets, ok := wowapi.Pets()
	if !ok {
		log.Fatal("Unable to obtain pet names.")
	}

	for _, petRaw := range allPets {
		pet := petRaw.(map[string]any)
		id := web.ToInt64(pet["id"])
		pets[id] = pet["name"].(string)
	}

	return pets
}

// petsOwned downloads the list of pets I own from the WoW web API
func petsOwned() map[int64][]wowitem.PetInfo {
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
		p.BreedId = web.ToInt64(stats["breed_id"])

		p.Level = web.ToInt64(pet["level"])

		quality, ok := pet["quality"].(map[string]any)
		if !ok {
			log.Fatal("Unable to obtain quality.")
		}
		p.QualityId = common.QualityId(quality["name"].(string))

		species, ok := pet["species"].(map[string]any)
		if !ok {
			log.Fatal("Unable to obtain species.")
		}
		p.SpeciesId = web.ToInt64(species["id"])

		_, ok = myPets[p.SpeciesId]
		if ok {
			name, _ := pet["species"].(map[string]any)
			fmt.Println("Duplicate pet:", name)
		} else {
			myPets[p.SpeciesId] = []wowitem.PetInfo{}
		}
		myPets[p.SpeciesId] = append(myPets[p.SpeciesId], p)
	}

	return myPets
}

func Init(oauthAvailable bool) {
	err := allNames.Load()
	if err != nil {
		fmt.Printf("*** error opening pet name cache, creating new one: %v\n", err)
		allNames.Data = refreshPetNames()
		allNames.Save()
	}
	if oauthAvailable {
		allOwned = petsOwned()
	}
	fmt.Printf("-- #Pets owned: %d/%d\n", len(allOwned), len(allNames.Data))
}

// PetSpell returns true and the corresponding pet ID if the item is a pet summoning spell
func PetSpell(i wowitem.Item) (int64, bool) {
	if len(allNames.Data) == 0 {
		log.Fatal("You must call battlepet.Init() before calling battlepet.PetSpell()")
	}

	if i.ItemSubclassName() != "Companion Pets" {
		return 0, false
	}

	for petId, petName := range allNames.Data {
		if i.Name() == petName {
			return petId, true
		}
	}

	return 0, false
}

// Name returns the pet name for the given ID
func Name(petID int64) string {
	if len(allNames.Data) == 0 {
		log.Fatal("You must call battlepet.Init() before calling battlepet.IdToName()")
	}
	return allNames.Data[petID]
}

// Owned returns true if I own this pet ID
func Owned(petID int64) bool {
	if len(allOwned) == 0 {
		log.Fatal("You must call battlepet.Init() before calling battlepet.Own()")
	}
	return len(allOwned[petID]) > 0
}
