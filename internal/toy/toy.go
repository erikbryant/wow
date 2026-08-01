package toy

import (
	"fmt"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

var (
	allNames = map[int64]string{}
	allOwned = map[int64]bool{}
)

// toyNames returns a map of all toy names by ID
func toyNames() (map[int64]string, error) {
	toys := map[int64]string{}

	allToys, ok := wowapi.Toys()
	if !ok {
		return nil, fmt.Errorf("unable to obtain toy names")
	}

	for _, toyRaw := range allToys {
		toy := toyRaw.(map[string]any)
		id := web.ToInt64(toy["id"])
		toys[id] = toy["name"].(string)
	}

	return toys, nil
}

// owned returns the toys I own
func owned() (map[int64]bool, error) {
	myToys := map[int64]bool{}

	toys, ok := wowapi.CollectionsToys()
	if !ok {
		return nil, fmt.Errorf("unable to obtain toys owned")
	}

	for _, toyRaw := range toys {
		toy := toyRaw.(map[string]any)
		id, _ := web.MsiValued(toy, []string{"toy", "id"}, 0)
		myToys[web.ToInt64(id)] = true
	}

	return myToys, nil
}

func Init() error {
	var err error

	allNames, err = toyNames()
	if err != nil {
		return err
	}

	allOwned, err = owned()
	if err != nil {
		return err
	}

	return nil
}

func Own(i wowitem.Item) bool {
	for toyID, name := range allNames {
		if i.Name() == name {
			return allOwned[toyID]
		}
	}
	return false
}
