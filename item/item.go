package item

import (
	"fmt"
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/common"
	"time"
)

// Item holds values about a WoW item
type Item struct {
	// WARNING: Changing this struct invalidates the cache
	// The members have to be public to write to a gob file,
	// but only use the accessor functions!
	XId      int64
	XItem    map[string]interface{}
	XUpdated time.Time // Datetime when created or updated
}

// PetInfo contains the properties specific to a battle pet
type PetInfo struct {
	BreedId   int64
	Level     int64
	Name      string
	QualityId int64
	SpeciesId int64
}

// NewItem returns an Item populated with wowData
func NewItem(wowData map[string]interface{}) Item {
	return Item{
		XId:      web.ToInt64(wowData["id"]),
		XItem:    wowData,
		XUpdated: time.Now(),
	}
}

// Id returns the item ID
func (i Item) Id() int64 {
	return i.XId
}

// Equippable returns true if the item is equippable
func (i Item) Equippable() bool {
	equippable := i.XItem["is_equippable"].(bool)

	if equippable {
		return true
	}

	// Is this a special equippable?
	previewItem := i.XItem["preview_item"].(map[string]interface{})
	binding, ok := previewItem["binding"].(map[string]interface{})
	if ok {
		switch binding["type"].(string) {
		case "ON_EQUIP":
			equippable = true
		case "ON_USE":
			equippable = true
		case "HEALTH":
			equippable = true
		default:
			fmt.Println("LookupItem: Item had unknown binding_type:", i.Id(), binding["type"].(string))
			equippable = false
		}
	}

	return equippable
}

// Level returns the item level
func (i Item) Level() int64 {
	_, ok := i.XItem["preview_item"]
	if ok {
		previewItem := i.XItem["preview_item"].(map[string]interface{})
		_, ok = previewItem["level"]
		if ok {
			level := previewItem["level"].(map[string]interface{})
			_, ok = level["value"]
			if ok {
				return web.ToInt64(level["value"])
			}
		}
	}

	return 0
}

// Name returns the item name
func (i Item) Name() string {
	return i.XItem["name"].(string)
}

// SellPrice returns the item vendor sell price (zero if unsure)
func (i Item) SellPrice() int64 {
	if i.Equippable() {
		// Don't know how to price these
		return 0
	}

	switch i.Id() {
	case 194829:
		// Fated Fortune Card (only gets a price once read)
		return 10000
	}

	return web.ToInt64(i.XItem["sell_price"])
}

// Updated returns the last time this item was updated in the cache
func (i Item) Updated() time.Time {
	return i.XUpdated
}

// Format returns a formatted string representing the item
func (i Item) Format() string {
	equippable := "F"
	if i.Equippable() {
		equippable = "T"
	}
	return fmt.Sprintf("%7d  %s %11s   %3d   %s   %s", i.Id(), equippable, common.Gold(i.SellPrice()), i.Level(), i.Updated().Format("2006-01-02"), i.Name())
}
