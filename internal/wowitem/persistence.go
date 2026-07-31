package wowitem

import (
	"encoding/gob"
	"fmt"
	"slices"
	"time"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/wowapi"
)

const (
	persistName = "items"
)

var (
	Items = persist.New[int64, Item](persistName)
)

func init() {
	gob.Register(map[string]any{})
	gob.Register([]any{})
	err := Items.Load()
	if err != nil {
		fmt.Printf("*** error opening items persist, creating new one: %v\n", err)
	}
	fmt.Printf("-- #Items in cache: %d\n", Items.Len())
}

// IDs returns the sorted list of keys from the item cache file
func IDs() []int64 {
	keys := Items.Keys()
	slices.Sort(keys)
	return keys
}

// Search returns the item with name 's' or an empty item if not found
func Search(s string) Item {
	_, i, ok := Items.Search(func(v Item) bool {
		return v.Name() == s
	})
	if !ok {
		fmt.Println("Did not find item for search string: ", s)
	}
	return i
}

// LookupItem retrieves a single item. From persistence if present, web if not. If from the web, also store it.
func LookupItem(id int64, age time.Duration) (Item, bool) {
	// Use the persisted value if exists and not stale
	i, ok := Items.Get(id)
	if ok {
		if !i.Stale(age) {
			return i, true
		}
		fmt.Println("Refreshing stale item:", i.Updated().Format("2006-01-02"), i.ID(), i.Name())
	}

	result, ok := wowapi.Item(web.ToString(id))
	if !ok {
		return Item{}, false
	}

	i = NewItem(result)
	Items.Set(i.ID(), i)

	return i, true
}
