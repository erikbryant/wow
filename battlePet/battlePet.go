package battlePet

import (
	"encoding/gob"
	"fmt"
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/common"
	"github.com/erikbryant/wow/item"
	"github.com/erikbryant/wow/wowAPI"
	"log"
	"os"
	"sort"
	"strings"
)

var (
	PetCageItemId    = int64(82800)
	petNameCacheFile = "./generated/petNameCache.gob"
	allNames         = map[int64]string{}
	allOwned         = map[int64][]item.PetInfo{}
)

func Init(oauthAvailable bool) {
	load()
	if oauthAvailable {
		allOwned = owned()
	}
	fmt.Printf("-- #Pets owned %d/%d\n", len(allOwned), len(allNames))
}

// load loads the disk cache file into memory
func load() {
	file, err := os.Open(petNameCacheFile)
	if err != nil {
		fmt.Printf("*** error opening petNameCache: %v, creating new one\n", err)
		allNames = petNames()
		save()
		return
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&allNames)
	if err != nil {
		log.Fatalf("error reading petNameCache: %v", err)
	}
}

// save writes the in-memory cache file to disk
func save() {
	file, err := os.Create(petNameCacheFile)
	if err != nil {
		log.Fatalf("error creating appearance cache file: %v", err)
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(allNames)
	if err != nil {
		log.Fatalf("error encoding allSetIds: %v", err)
	}
}

// owned returns the pets I own
func owned() map[int64][]item.PetInfo {
	myPets := map[int64][]item.PetInfo{}

	pets, ok := wowAPI.CollectionsPets()
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
		if ok {
			name, _ := pet["species"].(map[string]interface{})
			fmt.Println("Duplicate pet:", name)
		} else {
			myPets[p.SpeciesId] = []item.PetInfo{}
		}
		myPets[p.SpeciesId] = append(myPets[p.SpeciesId], p)
	}

	return myPets
}

// petNames returns a map of all battle pet names by petId
func petNames() map[int64]string {
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
	if len(allOwned) == 0 {
		log.Fatal("ERROR: You must call battlePet.Init() before calling battlePet.Own()")
	}
	return len(allOwned[petId]) > 0
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
	for id, pets := range allOwned {
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
