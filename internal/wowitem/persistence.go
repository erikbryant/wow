package wowitem

import (
	"fmt"
	"os"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/wowapi"
)

// TODO: Convert this to a singleton package (like appearanceset.go)

const (
	persistName = "items"
)

type WoWItem struct {
	Items *persist.Persistence[int64, Item]
}

func New() *WoWItem {
	wi := &WoWItem{
		Items: persist.New[int64, Item](persistName),
	}

	err := wi.Items.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "*** error opening items persist, using an empty one: %s\n", err)
	}
	fmt.Printf("-- #Items persisted  : %d\n", wi.Items.Len())

	return wi
}

// Search returns the item with name 's' or an empty item if not found
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
func (wi *WoWItem) GetWeb(id int64) (Item, bool) {
	fmt.Println("Downloading item:", id)
	result, ok := wowapi.Item(web.ToString(id))
	if !ok {
		return Item{}, false
	}

	i := NewItem(result)
	wi.Items.Set(i.ID(), i)

	return i, true
}

// Get retrieves a single item. From persistence if present, web if not.
func (wi *WoWItem) Get(id int64) (Item, bool) {
	i, ok := wi.Items.Get(id)
	if ok {
		return i, true
	}
	return wi.GetWeb(id)
}
