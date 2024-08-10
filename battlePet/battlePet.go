package battlePet

import (
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/common"
	"github.com/erikbryant/wow/wowAPI"
	"log"
)

var (
	PetCageItemId = int64(82800)
	allNames      = map[int64]string{}
	owned         = map[int64][]common.PetInfo{}
)

// Owned returns the pets I own
func Owned(accessToken string) map[int64][]common.PetInfo {
	myPets := map[int64][]common.PetInfo{}

	pets, ok := wowAPI.CollectionsPets(accessToken)
	if !ok {
		log.Fatal("ERROR: Unable to obtain pets owned.")
	}

	for _, petRaw := range pets {
		pet := petRaw.(map[string]interface{})

		var p common.PetInfo

		stats, ok := pet["stats"].(map[string]interface{})
		if !ok {
			log.Fatal("ERROR: Unable to obtain stats.")
		}
		p.BreedId = web.ToInt64(stats["breed_id"])

		p.Level = web.ToInt64(pet["level"])

		quality, ok := pet["quality"].(map[string]interface{})
		if !ok {
			log.Fatal("ERROR: Unable to obtain quality.")
		}
		p.QualityId = common.QualityId(quality["name"].(string))

		species, ok := pet["species"].(map[string]interface{})
		if !ok {
			log.Fatal("ERROR: Unable to obtain species.")
		}
		p.SpeciesId = web.ToInt64(species["id"])

		_, ok = myPets[p.SpeciesId]
		if !ok {
			myPets[p.SpeciesId] = []common.PetInfo{}
		}
		myPets[p.SpeciesId] = append(myPets[p.SpeciesId], p)
	}

	return myPets
}

// PetNames returns a map of all battle pet names by petId
func PetNames(accessToken string) map[int64]string {
	pets := map[int64]string{}

	allPets, ok := wowAPI.Pets(accessToken)
	if !ok {
		log.Fatal("ERROR: Unable to obtain pets.")
	}

	for _, petRaw := range allPets {
		pet := petRaw.(map[string]interface{})
		id := web.ToInt64(pet["id"])
		pets[id] = pet["name"].(string)
	}

	return pets
}

func Name(petId int64) string {
	if len(allNames) == 0 {
		log.Fatal("ERROR: You must call battlePet.Init() before calling battlePet.Name()")
	}
	return allNames[petId]
}

func Own(petId int64) bool {
	if len(owned) == 0 {
		log.Fatal("ERROR: You must call battlePet.Init() before calling battlePet.OwnPet()")
	}
	return len(owned[petId]) > 0
}

func Init(accessToken string) {
	allNames = PetNames(accessToken)
	owned = Owned(accessToken)
}
