package battlePet

import (
	"fmt"
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/common"
	"github.com/erikbryant/wow/item"
	"github.com/erikbryant/wow/wowAPI"
	"log"
	"sort"
	"strings"
)

var (
	PetCageItemId = int64(82800)
	allNames      = map[int64]string{}
	owned         = map[int64][]item.PetInfo{}
)

func Init(profileAccessToken string) {
	allNames = PetNames()
	owned = Owned(profileAccessToken)
}

// Owned returns the pets I own
func Owned(profileAccessToken string) map[int64][]item.PetInfo {
	myPets := map[int64][]item.PetInfo{}

	pets, ok := wowAPI.CollectionsPets(profileAccessToken)
	if !ok {
		log.Fatal("ERROR: Unable to obtain pets owned.")
	}

	for _, petRaw := range pets {
		pet := petRaw.(map[string]interface{})

		var p item.PetInfo

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
			myPets[p.SpeciesId] = []item.PetInfo{}
		}
		myPets[p.SpeciesId] = append(myPets[p.SpeciesId], p)
	}

	return myPets
}

// PetNames returns a map of all battle pet names by petId
func PetNames() map[int64]string {
	pets := map[int64]string{}

	allPets, ok := wowAPI.Pets()
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

// IsPetSpell returns true and the corresponding pet ID if the item is a pet summoning spell
func IsPetSpell(i item.Item) (int64, bool) {
	if len(allNames) == 0 {
		log.Fatal("ERROR: You must call battlePet.Init() before calling battlePet.IsPetSpell()")
	}

	if i.ItemSubclassName() != "Companion Pets" {
		return 0, false
	}

	for petId, petName := range allNames {
		if i.Name() == petName {
			return petId, true
		}
	}

	return 0, false
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

func Format(pet item.PetInfo) string {
	return fmt.Sprintf("%4d  %2d  %-8s  %s", pet.SpeciesId, pet.Level, common.QualityName(pet.QualityId), allNames[pet.SpeciesId])
}

func LuaPetId() string {
	lua := ""

	rows := []string{}
	lua += fmt.Sprintf("local SpeciesIdCache = {\n")
	for petId, name := range allNames {
		row := fmt.Sprintf("  [\"%s\"] = %d,", name, petId)
		rows = append(rows, row)
	}
	sort.Strings(rows)
	lua += strings.Join(rows, "\n")

	lua += fmt.Sprintf("\n}\n")
	lua += fmt.Sprintf(`
local function SpeciesId(name)
    return SpeciesIdCache[name] or 0
end
`)

	lua += "\n"

	highestLevelOwned := map[int64]int64{}
	for id, pets := range owned {
		for _, pet := range pets {
			if pet.Level > highestLevelOwned[id] {
				highestLevelOwned[id] = pet.Level
			}
		}
	}

	rows = []string{}
	lua += fmt.Sprintf("local OwnedLevelCache = {\n")
	for petId, level := range highestLevelOwned {
		name := allNames[petId]
		row := fmt.Sprintf("  [\"%s\"] = %d,", name, level)
		rows = append(rows, row)
	}
	sort.Strings(rows)
	lua += strings.Join(rows, "\n")

	lua += fmt.Sprintf("\n}\n")
	lua += fmt.Sprintf(`
local function OwnedLevel(speciesID)
    return OwnedLevelCache[speciesID] or 0
end

AhaPetCache = {
  SpeciesId = SpeciesId,
  OwnedLevel = OwnedLevel,
}
`)

	return lua
}
