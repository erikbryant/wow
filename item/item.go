package item

import (
	"fmt"
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/common"
	"strings"
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

// Binding returns whether and when the item binds
func (i Item) Binding() string {
	value := common.MSIValue(i.XItem, []string{"preview_item", "binding", "type"})
	if value == nil {
		return ""
	}
	return value.(string)
}

// Equippable returns true if the item is equippable
func (i Item) Equippable() bool {
	// Is this a regular equippable?
	equippable := i.XItem["is_equippable"].(bool)
	if equippable {
		return true
	}

	// Is this a special equippable?
	binding := strings.ToUpper(i.Binding())
	switch binding {
	case "ON_EQUIP":
		equippable = true
	case "ON_USE":
		equippable = true
	case "HEALTH":
		equippable = false
	case "TO_ACCOUNT":
		equippable = false
	case "":
		equippable = false
	default:
		fmt.Println("LookupItem: Item had unknown binding_type:", i.Id(), binding)
		equippable = false
	}

	return equippable
}

// Level returns the item level
func (i Item) Level() int64 {
	return common.MSIValue(i.XItem, []string{"preview_item", "level", "value"}).(int64)
}

// ItemClassName returns the item class name
func (i Item) ItemClassName() string {
	return common.MSIValue(i.XItem, []string{"item_class", "name"}).(string)
}

// Name returns the item name
func (i Item) Name() string {
	return i.XItem["name"].(string)
}

// SellPrice returns the item vendor sell price (zero if unsure)
func (i Item) SellPrice() int64 {
	className := i.ItemClassName()
	if className == "Weapon" || className == "Armor" {
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
	return fmt.Sprintf("%7d  %s %11s   %3d   %-15s   %s   %s", i.Id(), equippable, common.Gold(i.SellPrice()), i.Level(), i.ItemClassName(), i.Updated().Format("2006-01-02"), i.Name())
}
