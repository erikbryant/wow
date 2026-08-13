package toy

import (
	"fmt"
	"sync"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

type Toy struct {
	names map[int64]string
	owned map[int64]bool
}

var (
	once = sync.Once{}
	toys *Toy
)

// getNames returns a map of all toy names by ID
func getNames() (map[int64]string, error) {
	toyNames := map[int64]string{}

	allToys, err := wowapi.Toys()
	if err != nil {
		return nil, fmt.Errorf("unable to obtain toy names: %w", err)
	}

	for _, toyRaw := range allToys {
		toy := toyRaw.(map[string]any)
		id := web.ToInt64(toy["id"])
		toyNames[id] = toy["name"].(string)
	}

	return toyNames, nil
}

// getOwned returns the toys I own
func getOwned() (map[int64]bool, error) {
	myToys := map[int64]bool{}

	toysOwned, err := wowapi.CollectionsToys()
	if err != nil {
		return nil, fmt.Errorf("unable to obtain toys owned: %w", err)
	}

	for _, toyRaw := range toysOwned {
		toy := toyRaw.(map[string]any)
		id, _ := web.MsiValued(toy, []string{"toy", "id"}, 0)
		myToys[web.ToInt64(id)] = true
	}

	return myToys, nil
}

func get() error {
	var t Toy
	var err error

	t.names, err = getNames()
	if err != nil {
		return err
	}

	t.owned, err = getOwned()
	if err != nil {
		return err
	}

	toys = &t

	return nil
}

func New() (*Toy, error) {
	var err error
	once.Do(func() {
		err = get()
	})
	return toys, err
}

func (t *Toy) Owned(i wowitem.Item) bool {
	for toyID, name := range t.names {
		if i.Name() == name {
			return t.owned[toyID]
		}
	}
	return false
}
