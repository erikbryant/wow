package wowitem

import (
	"fmt"
	"os"
	"strconv"

	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/wowapi"
)

type WoWItem struct {
}

const (
	persistName = "items"
)

var (
	itemPersistence *persist.Persistence[int64, Item]
)

func New() *WoWItem {
	if itemPersistence == nil {
		itemPersistence = persist.New[int64, Item](persistName)
	}

	if itemPersistence.Loaded() {
		return &WoWItem{}
	}

	err := itemPersistence.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "*** error opening items persist, using an empty one: %s\n", err)
	}
	fmt.Printf("-- #Items persisted  : %d\n", itemPersistence.Len())

	return &WoWItem{}
}

// Search returns the first item with name 's' (duplicates are very rare) or an empty item if not found
func (wi *WoWItem) Search(s string) Item {
	_, i, ok := itemPersistence.Search(func(v Item) bool {
		return v.Name() == s
	})
	if !ok {
		fmt.Fprintf(os.Stderr, "*** did not find item for search string: %s\n", s)
	}
	return i
}

// GetLive retrieves a single item from the WoW web API and persists it.
func (wi *WoWItem) GetLive(id int64) (Item, error) {
	fmt.Println("Downloading item:", id)
	result, err := wowapi.Item(strconv.FormatInt(id, 10))
	if err != nil {
		return Item{}, err
	}

	i := NewItem(result)
	itemPersistence.Set(i.ID(), i)

	return i, nil
}

// Get retrieves a single item. From persistence if present, web if not.
func (wi *WoWItem) Get(id int64) (Item, error) {
	i, ok := itemPersistence.Get(id)
	if ok {
		return i, nil
	}
	return wi.GetLive(id)
}

func (wi *WoWItem) Delete(id int64) {
	itemPersistence.Delete(id)
}

func (wi *WoWItem) Save() error {
	return itemPersistence.Save()
}

func (wi *WoWItem) Values() []Item {
	return itemPersistence.Values()
}

func (wi *WoWItem) Keys() []int64 {
	return itemPersistence.Keys()
}
