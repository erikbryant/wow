package wowitem

import (
	"fmt"
	"os"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/wowapi"
)

type WoWItem struct {
	Items *persist.Persistence[int64, Item]
}

const (
	persistName = "items"
)

var (
	wowItems = &WoWItem{
		Items: persist.New[int64, Item](persistName),
	}
)

func New() *WoWItem {
	if wowItems.Items.Loaded() {
		return wowItems
	}

	err := wowItems.Items.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "*** error opening items persist, using an empty one: %s\n", err)
	}
	fmt.Printf("-- #Items persisted  : %d\n", wowItems.Items.Len())

	return wowItems
}

// Search returns the first item with name 's' (duplicates are very rare) or an empty item if not found
func (wi *WoWItem) Search(s string) Item {
	_, i, ok := wi.Items.Search(func(v Item) bool {
		return v.Name() == s
	})
	if !ok {
		fmt.Fprintf(os.Stderr, "*** did not find item for search string: %s\n", s)
	}
	return i
}

// GetWeb retrieves a single item from the web and persists it.
func (wi *WoWItem) GetWeb(id int64) (Item, error) {
	fmt.Println("Downloading item:", id)
	result, err := wowapi.Item(web.ToString(id))
	if err != nil {
		return Item{}, err
	}

	i := NewItem(result)
	wi.Items.Set(i.ID(), i)

	return i, nil
}

// Get retrieves a single item. From persistence if present, web if not.
func (wi *WoWItem) Get(id int64) (Item, error) {
	i, ok := wi.Items.Get(id)
	if ok {
		return i, nil
	}
	return wi.GetWeb(id)
}
