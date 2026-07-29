package battlepet

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/common"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

const (
	cacheFilename = "./data/petNameCache.gob"
)

var (
	PetCageItemId = int64(82800)
	allNames      = map[int64]string{}
	allOwned      = map[int64][]wowitem.PetInfo{}
	mu            sync.Mutex
)

func Init(oauthAvailable bool) {
	load()
	if oauthAvailable {
		allOwned = owned()
	}
	fmt.Printf("-- #Pets owned: %d/%d\n", len(allOwned), len(allNames))
}

// load loads the disk cache file into memory
func load() {
	mu.Lock()
	file, err := os.Open(cacheFilename)
	if err != nil {
		fmt.Printf("*** error opening petNameCache: %v, creating new one\n", err)
		allNames = petNames()
		err = save()
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&allNames)
	if err != nil {
		log.Fatalf("error reading petNameCache: %v", err)
	}
	mu.Unlock()
}

// save writes the in-memory cache file to disk
func save() error {
	mu.Lock()
	defer mu.Unlock()

	tmp := cacheFilename + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	encoder := gob.NewEncoder(f)

	if err := encoder.Encode(allNames); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}

	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}

	return os.Rename(tmp, cacheFilename)
}

// owned returns the pets I own
func owned() map[int64][]wowitem.PetInfo {
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

// petNames returns a map of all battle pet names by petId
func petNames() map[int64]string {
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

// IsPetSpell returns true and the corresponding pet ID if the item is a pet summoning spell
func IsPetSpell(i wowitem.Item) (int64, bool) {
	if len(allNames) == 0 {
		log.Fatal("You must call battlepet.Init() before calling battlepet.IsPetSpell()")
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
		log.Fatal("You must call battlepet.Init() before calling battlepet.Name()")
	}
	return allNames[petId]
}

func Own(petId int64) bool {
	if len(allOwned) == 0 {
		log.Fatal("You must call battlepet.Init() before calling battlepet.Own()")
	}
	return len(allOwned[petId]) > 0
}
