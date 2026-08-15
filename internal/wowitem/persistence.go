package wowitem

import (
	"fmt"
	"os"
	"strconv"

	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/wowapi"
)

type WoWItem struct {
	items persist.Persistence[int64, Item]
}

func NewEmpty(persistencePath string) *WoWItem {
	return &WoWItem{
		items: *persist.New[int64, Item](persistencePath + ".new"),
	}
}

func New(persistencePath string) (*WoWItem, error) {
	wi := WoWItem{
		items: *persist.New[int64, Item](persistencePath),
	}

	err := wi.items.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading items persist: %w", err)
	}

	fmt.Printf("-- #Items persisted  : %d\n", wi.items.Len())

	return &wi, nil
}

// Search returns the first item with name 's' (duplicates are very rare) or an empty item if not found
func (wi *WoWItem) Search(s string) Item {
	_, i, ok := wi.items.Search(func(v Item) bool {
		return v.Name() == s
	})
	if !ok {
		fmt.Fprintf(os.Stderr, "*** did not find item for search string: %s\n", s)
	}
	return i
}

// GetLive retrieves a single item from the WoW web API and persists it.
func (wi *WoWItem) GetLive(id int64) (Item, error) {
	result, err := wowapi.Item(strconv.FormatInt(id, 10))
	if err != nil {
		return Item{}, err
	}

	fmt.Println("Downloaded new item:", id)

	i := NewItem(result)
	wi.items.Set(i.ID(), i)

	return i, nil
}

// Get retrieves a single item. From persistence if present, web if not.
func (wi *WoWItem) Get(id int64) (Item, error) {
	i, ok := wi.items.Get(id)
	if ok {
		return i, nil
	}
	return wi.GetLive(id)
}

func (wi *WoWItem) Delete(id int64) {
	wi.items.Delete(id)
}

func (wi *WoWItem) Dirty() bool {
	return wi.items.Dirty()
}

func (wi *WoWItem) Save() error {
	return wi.items.Save()
}

func (wi *WoWItem) Values() []Item {
	return wi.items.Values()
}

func (wi *WoWItem) Keys() []int64 {
	return wi.items.Keys()
}

func (wi *WoWItem) Path() string {
	return wi.items.Path()
}
