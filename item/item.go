package item

import (
	"fmt"
	"github.com/erikbryant/wow/common"
	"time"
)

// Item contains the properties of a single WoW item
type Item struct {
	// WARNING: Changing this struct invalidates the cache
	Id         int64
	Name       string
	Equippable bool
	SellPrice  int64
	ItemLevel  int64
	Updated    time.Time // Datetime when created or updated
}

type NewItem struct {
	// WARNING: Changing this struct invalidates the cache
	id      int64
	item    map[string]interface{}
	updated time.Time // Datetime when created or updated
}

// PetInfo contains the properties specific to a battle pet
type PetInfo struct {
	BreedId   int64
	Level     int64
	Name      string
	QualityId int64
	SpeciesId int64
}

func (i Item) Format() string {
	equippable := "F"
	if i.Equippable {
		equippable = "T"
	}
	return fmt.Sprintf("%7d  %s %11s   %3d   %s   %s", i.Id, equippable, common.Gold(i.SellPrice), i.ItemLevel, i.Updated.Format("2006-01-02"), i.Name)
}
