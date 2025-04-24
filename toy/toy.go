package toy

import (
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/item"
	"github.com/erikbryant/wow/wowAPI"
	"log"
)

var (
	allNames = map[int64]string{}
	allOwned = map[int64]bool{}
)

func Init(profileAccessToken string) {
	allNames = toyNames()
	allOwned = owned(profileAccessToken)
}

// owned returns the toys I own
func owned(profileAccessToken string) map[int64]bool {
	myToys := map[int64]bool{}

	toys, ok := wowAPI.CollectionsToys(profileAccessToken)
	if !ok {
		log.Fatal("ERROR: Unable to obtain toys owned.")
	}

	for _, toyRaw := range toys {
		toy := toyRaw.(map[string]interface{})
		id, _ := web.MsiValued(toy, []string{"toy", "id"}, 0)
		myToys[web.ToInt64(id)] = true
	}

	return myToys
}

// toyNames returns a map of all toy names by Id
func toyNames() map[int64]string {
	toys := map[int64]string{}

	allToys, ok := wowAPI.Toys()
	if !ok {
		log.Fatal("ERROR: Unable to obtain toys.")
	}

	for _, toyRaw := range allToys {
		toy := toyRaw.(map[string]interface{})
		id := web.ToInt64(toy["id"])
		toys[id] = toy["name"].(string)
	}

	return toys
}

func Own(i item.Item) bool {
	if len(allOwned) == 0 {
		log.Fatal("ERROR: You must call toy.Init() before calling toy.Own()")
	}

	for _, name := range allNames {
		if i.Name() == name {
			return true
		}
	}

	return false
}
