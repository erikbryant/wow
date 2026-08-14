package toy

import (
	"fmt"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

type Toy struct {
	// Toy IDs. These are not the same as item IDs.
	names map[string]int64
	owned map[int64]bool
}

var (
	toys *Toy
)

// getNames returns a map of all toy names by toy ID
func getNames() (map[string]int64, error) {
	toyNames := map[string]int64{}

	allToys, err := wowapi.Toys()
	if err != nil {
		return nil, fmt.Errorf("unable to obtain toy names: %w", err)
	}

	for _, toyRaw := range allToys {
		toy := toyRaw.(map[string]any)
		id := web.ToInt64(toy["id"])
		toyNames[toy["name"].(string)] = id
	}

	return toyNames, nil
}

// getOwned returns the toys I own by toy ID
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

func get() (*Toy, error) {
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

func New() (*Toy, error) {
	if toys != nil {
		return toys, nil
	}

	var err error
	toys, err = get()
	if err != nil {
		return nil, err
	}

	return toys, err
}

// Owned returns true if the item name is a toy and we own it
func (t *Toy) Owned(i wowitem.Item) bool {
	toyID, ok := t.names[i.Name()]
	if !ok {
		return false
	}
	return t.owned[toyID]
}
