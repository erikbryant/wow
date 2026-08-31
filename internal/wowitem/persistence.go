package wowitem

import (
	"fmt"
	"os"
	"slices"
	"strconv"

	"github.com/erikbryant/wow/internal/persist"
	"github.com/erikbryant/wow/internal/wowapi"
)

type Persistence struct {
	*persist.Persistence[int64, Item]
}

// NewEmpty creates a new Persistence with no items in it.
func NewEmpty(persistencePath string) *Persistence {
	return &Persistence{
		Persistence: persist.New[int64, Item](persistencePath + ".new"),
	}
}

// New creates a new Persistence, populated with data from its persistence store.
func New(persistencePath string) (*Persistence, error) {
	p := &Persistence{
		Persistence: persist.New[int64, Item](persistencePath),
	}

	if err := p.Load(); err != nil {
		return nil, fmt.Errorf("error loading items persist: %w", err)
	}

	return p, nil
}

// Search returns the first item with name s.
// Duplicates are very rare.
func (p *Persistence) Search(s string) *Item {
	_, item, ok := p.Persistence.Search(func(v Item) bool {
		return v.Name() == s
	})
	if !ok {
		fmt.Fprintf(os.Stderr, "*** did not find item for search string: %s\n", s)
	}

	return &item
}

// GetLive retrieves a single item from the WoW web API and persists it.
func (p *Persistence) GetLive(id int64, client *wowapi.Client) (Item, error) {
	result, err := client.Item(strconv.FormatInt(id, 10))
	if err != nil {
		return Item{}, err
	}

	fmt.Println("Downloaded new item:", id)

	item := NewItem(result)
	p.Set(item.ID(), *item)

	return *item, nil
}

// Get retrieves a single item. From persistence if present, web if not.
func (p *Persistence) Get(id int64, client *wowapi.Client) (Item, error) {
	item, ok := p.Persistence.Get(id)
	if ok {
		return item, nil
	}

	return p.GetLive(id, client)
}

func (p *Persistence) Keys() []int64 {
	keys := p.Persistence.Keys()
	slices.Sort(keys)
	return keys
}
