package common

// Item contains the properties of a single auction house item
type Item struct {
	// WARNING: Changing this struct invalidates the cache
	Id         int64
	Name       string
	Equippable bool
	SellPrice  int64
	ItemLevel  int64
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
