package item

import (
	"fmt"
	"log"
	"time"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/common"
)

// Item holds values about a WoW item
type Item struct {
	// WARNING: Changing this struct invalidates the cache.
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
	// The key is only sometimes there; do not error if it is missing
	value, _ := web.MsiValued(i.XItem, []string{"preview_item", "binding", "type"}, "")
	return value.(string)
}

// InventoryType returns the slot this item equips to, or UNKNOWN
func (i Item) InventoryType() string {
	// The key is only sometimes there; do not error if it is missing
	value, _ := web.MsiValued(i.XItem, []string{"inventory_type"}, "UNKNOWN")
	return value.(string)
}

// EquipSlotTypes is a lookup set for valid gear slots
var EquipSlotTypes = map[string]struct{}{
	"HEAD":       {},
	"NECK":       {},
	"SHOULDER":   {},
	"CHEST":      {},
	"WAIST":      {},
	"LEGS":       {},
	"FEET":       {},
	"WRIST":      {},
	"HANDS":      {},
	"FINGER":     {},
	"TRINKET":    {},
	"CLOAK":      {},
	"ONE_HANDED": {},
	"TWO_HANDED": {},
	"MAIN_HAND":  {},
	"OFF_HAND":   {},
	"RANGED":     {},
	"SHIELD":     {},
}

// Equippable returns true if the item is equippable
func (i Item) Equippable() bool {
	// Preferred authoritative field
	if val, ok := i.XItem["is_equippable"]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}

	// Fallback: inventory_type
	_, isEquipSlot := EquipSlotTypes[i.InventoryType()]
	return isEquipSlot
}

// ItemLevel returns the item level
func (i Item) ItemLevel() int64 {
	value, err := web.MsiValue(i.XItem, []string{"level"})
	if err != nil {
		log.Fatalf("Level: %s in %v", err, i.XItem)
	}
	return web.ToInt64(value)
}

// VariableItemLevel returns true if the item can be enhanced, changing its iLevel
func (i Item) VariableItemLevel() bool {
	if i.Stackable() {
		return false
	}
	cn := i.ItemClassName()
	return cn == "Armor" || cn == "Gem" || cn == "Weapon"
}

// ItemSubclassName returns the item subclass name
func (i Item) ItemSubclassName() string {
	value, _ := web.MsiValued(i.XItem, []string{"item_subclass", "name"}, "")
	return value.(string)
}

// Cosmetic returns true if this item is a cosmetic
func (i Item) Cosmetic() bool {
	if i.ItemSubclassName() == "Cosmetic" {
		return true
	}
	if i.ItemLevel() == 1 && (i.Quality() == "Rare" || i.Quality() == "Epic") {
		return true
	}
	return false
}

// ItemClassName returns the item class name
func (i Item) ItemClassName() string {
	value, err := web.MsiValue(i.XItem, []string{"item_class", "name"})
	if err != nil {
		log.Fatalf("ItemClassName: %s in %v", err, i.XItem)
	}
	return value.(string)
}

// Stackable returns true if the item can be stacked in the inventory
func (i Item) Stackable() bool {
	value, err := web.MsiValue(i.XItem, []string{"is_stackable"})
	if err != nil {
		log.Fatalf("Stackable: %s in %v", err, i.XItem)
	}
	return value.(bool)
}

// RelicType returns the relic type
func (i Item) RelicType() string {
	// The key is only sometimes there; do not error if it is missing
	value, _ := web.MsiValued(i.XItem, []string{"preview_item", "gem_properties", "relic_type"}, "")
	return value.(string)
}

// Name returns the item name
func (i Item) Name() string {
	return i.XItem["name"].(string)
}

func (i Item) previewPrice() (int64, error) {
	value, err := web.MsiValue(i.XItem, []string{"preview_item", "sell_price", "value"})
	if err != nil {
		return 0, err
	}

	return web.ToInt64(value), nil
}

// SellPriceAdvertised returns the vendor sell price listed in the JSON
func (i Item) SellPriceAdvertised() int64 {
	pp, err := i.previewPrice()
	if err != nil {
		// Items with no preview price don't sell
		return 0
	}

	return pp
}

// SellPriceRealizable returns the actual price the vendor will offer for this specific item
func (i Item) SellPriceRealizable() int64 {
	if i.VariableItemLevel() {
		// I don't know how to price these
		return 0
	}
	return i.SellPriceAdvertised()
}

// Updated returns the last time this item was updated in the cache
func (i Item) Updated() time.Time {
	return i.XUpdated
}

func (i Item) Requirements() string {
	r, _ := web.MsiValued(i.XItem, []string{"preview_item", "requirements", "skill", "display_string"}, "")
	return web.ToString(r)
}

func (i Item) Quality() string {
	q, _ := web.MsiValued(i.XItem, []string{"preview_item", "quality", "name"}, "")
	return web.ToString(q)
}

func (i Item) Toy() bool {
	q, _ := web.MsiValued(i.XItem, []string{"preview_item", "toy"}, "")
	return web.ToString(q) == "Toy"
}

// Appearances returns the appearance IDs this item provides
func (i Item) Appearances() []int64 {
	appearanceIds := []int64{}

	q, _ := web.MsiValued(i.XItem, []string{"appearances"}, nil)
	if q == nil {
		// Most items do not have appearances
		return nil
	}

	appearances := q.([]interface{})
	for _, appearance := range appearances {
		id := appearance.(map[string]interface{})["id"]
		appearanceIds = append(appearanceIds, web.ToInt64(id))
	}

	return appearanceIds
}

// Format returns a formatted string representing the item
func (i Item) Format() string {
	equippable := "F"
	if i.Equippable() {
		equippable = "T"
	}
	stackable := "F"
	if i.Stackable() {
		stackable = "T"
	}
	return fmt.Sprintf("%7d  %s %s %11s   %3d   %-18s   %-8s   %s   %s", i.Id(), equippable, stackable, common.Gold(i.SellPriceAdvertised()), i.ItemLevel(), i.ItemClassName(), i.Quality(), i.Updated().Format("2006-01-02"), i.Name())
}
