package toy

import (
	"log"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

var (
	allNames = map[int64]string{}
	allOwned = map[int64]bool{}
)

func Init(oauthAvailable bool) {
	if !oauthAvailable {
		return
	}

	allNames = toyNames()
	allOwned = owned()
}

// owned returns the toys I own
func owned() map[int64]bool {
	myToys := map[int64]bool{}

	toys, ok := wowapi.CollectionsToys()
	if !ok {
		log.Fatal("Unable to obtain toys owned.")
	}

	for _, toyRaw := range toys {
		toy := toyRaw.(map[string]any)
		id, _ := web.MsiValued(toy, []string{"toy", "id"}, 0)
		myToys[web.ToInt64(id)] = true
	}

	return myToys
}

// toyNames returns a map of all toy names by ID
func toyNames() map[int64]string {
	toys := map[int64]string{}

	allToys, ok := wowapi.Toys()
	if !ok {
		log.Fatal("Unable to obtain toy names.")
	}

	for _, toyRaw := range allToys {
		toy := toyRaw.(map[string]any)
		id := web.ToInt64(toy["id"])
		toys[id] = toy["name"].(string)
	}

	return toys
}

func Own(i wowitem.Item) bool {
	if len(allOwned) == 0 {
		log.Fatal("You must call toy.Init() before calling toy.Own()")
	}

	for toyID, name := range allNames {
		if i.Name() == name {
			return allOwned[toyID]
		}
	}

	return false
}
