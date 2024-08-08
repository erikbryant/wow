package common

import (
	"fmt"
	"time"
)

// Item contains the properties of a single auction house item
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
	Id         int64
	Name       string
	Equippable bool
	SellPrice  int64
	ItemLevel  int64
	Updated    time.Time // Datetime when created or updated
}

// Sample auction response. Some have more or fewer fields.
// map[buyout:1.1111011e+09 id:3.49632108e+08 item:map[id:142075] quantity:1 time_left:VERY_LONG]

// Commodity auction response. All have exactly these fields.
// map[id:3.44371058e+08 item:map[id:192672] quantity:1 time_left:SHORT unit_price:16800]

// Auction contains the properties of a single auction house auction
type Auction struct {
	Id       int64
	ItemId   int64
	Buyout   int64 // For commodity auctions this stores 'unit_price'
	Quantity int64
}

// Coins returns a single numeric value of the given denominations
func Coins(g, s, c int64) int64 {
	return g*100*100 + s*100 + c
}

// Gold returns a formatted string of the given numeric value
func Gold(price int64) string {
	copper := price % 100
	price /= 100
	silver := price % 100
	price /= 100
	gold := price
	return fmt.Sprintf("%d.%02d.%02d", gold, silver, copper)
}

func (item Item) Format() string {
	equippable := "F"
	if item.Equippable {
		equippable = "T"
	}
	return fmt.Sprintf("%7d  %s %11s   %3d   %s   %s", item.Id, equippable, Gold(item.SellPrice), item.ItemLevel, item.Updated.Format("2006-01-02"), item.Name)
}
