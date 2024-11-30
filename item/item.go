package item

import (
	"fmt"
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/common"
	"log"
	"strings"
	"time"
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
	case "ON_ACQUIRE":
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
	// The key is only sometimes there; do not error if it is missing
	value, _ := web.MsiValued(i.XItem, []string{"preview_item", "level", "value"}, 0)
	return web.ToInt64(value)
}

// ItemClassName returns the item class name
func (i Item) ItemClassName() string {
	value, err := web.MsiValue(i.XItem, []string{"item_class", "name"})
	if err != nil {
		log.Fatalf("ItemClassName: %s in %v", err, i.XItem)
	}
	return value.(string)
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
	if i.Level() > 0 {
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

// Format returns a formatted string representing the item
func (i Item) Format() string {
	equippable := "F"
	if i.Equippable() {
		equippable = "T"
	}
	return fmt.Sprintf("%7d  %s %11s   %3d   %-18s   %-8s   %s   %s", i.Id(), equippable, common.Gold(i.SellPriceAdvertised()), i.Level(), i.ItemClassName(), i.Quality(), i.Updated().Format("2006-01-02"), i.Name())
}
