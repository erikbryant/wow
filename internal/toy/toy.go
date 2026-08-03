package toy

import (
	"fmt"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

type Toy struct {
	names map[int64]string
	owned map[int64]bool
}

// getNames returns a map of all toy names by ID
func getNames() (map[int64]string, error) {
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

// getOwned returns the toys I own
func getOwned() (map[int64]bool, error) {
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

func New() (*Toy, error) {
	var t Toy
	var err error

	t.names, err = getNames()
	if err != nil {
		return nil, err
	}

	t.owned, err = getOwned()
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (t *Toy) Owned(i wowitem.Item) bool {
	for toyID, name := range t.names {
		if i.Name() == name {
			return t.owned[toyID]
		}
	}
	return false
}
