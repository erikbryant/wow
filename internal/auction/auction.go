package auction

import (
	"fmt"
	"log"
	"strings"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/battlepet"
	"github.com/erikbryant/wow/internal/wowapi"
)

// Sample 'commodity auction' response. All have exactly these fields.
// map[id:3.44371058e+08 item:map[id:192672] quantity:1 time_left:SHORT unit_price:16800]

// Sample 'auction' response. Some have more or fewer fields.
// map[buyout:1.1111011e+09 id:3.49632108e+08 item:map[id:142075] quantity:1 time_left:VERY_LONG]

// Sample 'auction' response for a pet auction. ItemID 82800 is a 'Pet Cage'. Pet cages have no sell value.
// map[buyout:9.99e+06 id:5.01784773e+08 item:map[id:82800 modifiers:[map[type:6 value:39130]] pet_breed_id:20 pet_level:1 pet_quality_id:2 pet_species_id:1446] quantity:1 time_left:VERY_LONG]

// PetInfo contains the properties specific to a battle pet
type PetInfo struct {
	BreedID   int64
	Level     int64
	Name      string
	QualityID int64
	SpeciesID int64
}

// Auction contains the properties of a single auction house auction
type Auction struct {
	ID       int64
	ItemID   int64
	Buyout   int64 // For commodity auctions this stores 'unit_price'
	Quantity int64
	Pet      PetInfo
}

func auctionID(msi any) int64 {
	value, err := web.MsiValue(msi, []string{"id"})
	if err != nil {
		log.Fatalf("ID: %s in %v", err, msi)
	}
	return web.ToInt64(value)
}

func itemID(msi any) int64 {
	value, err := web.MsiValue(msi, []string{"item", "id"})
	if err != nil {
		log.Fatalf("ItemID: %s in %v", err, msi)
	}
	return web.ToInt64(value)
}

func buyout(msi any) int64 {
	value, err := web.MsiValue(msi, []string{"buyout"})
	if value == nil || err != nil {
		value, err = web.MsiValued(msi, []string{"unit_price"}, 0)
		if err != nil {
			// Some auctions have neither 'buyout' nor 'unit_price'. Strange, but true.
			return 0
		}
	}
	return web.ToInt64(value)
}

func quantity(msi any) int64 {
	value, err := web.MsiValued(msi, []string{"quantity"}, 0)
	if err != nil {
		log.Fatalf("Quantity: %s in %v", err, msi)
	}
	return web.ToInt64(value)
}

func petBreedID(msi any) int64 {
	value, err := web.MsiValue(msi, []string{"item", "pet_breed_id"})
	if err != nil {
		log.Fatalf("PetBreedID: %s in %v", err, msi)
	}
	return web.ToInt64(value)
}

func petLevel(msi any) int64 {
	value, err := web.MsiValue(msi, []string{"item", "pet_level"})
	if err != nil {
		log.Fatalf("PetLevel: %s in %v", err, msi)
	}
	return web.ToInt64(value)
}

func petQualityID(msi any) int64 {
	value, err := web.MsiValue(msi, []string{"item", "pet_quality_id"})
	if err != nil {
		log.Fatalf("PetQualityID: %s in %v", err, msi)
	}
	return web.ToInt64(value)
}

func petSpeciesID(msi any) int64 {
	value, err := web.MsiValue(msi, []string{"item", "pet_species_id"})
	if err != nil {
		log.Fatalf("PetSpeciesID: %s in %v", err, msi)
	}
	return web.ToInt64(value)
}

// newAuction converts a single auction JSON string into a struct that is much easier to work with
func newAuction(auc any) Auction {
	var a Auction

	a.ID = auctionID(auc)
	a.ItemID = itemID(auc)
	a.Buyout = buyout(auc)
	a.Quantity = quantity(auc)

	// Is this a Pet Cage?
	if a.ItemID == battlepet.PetCageItemID {
		// A pet auction!
		a.Pet.BreedID = petBreedID(auc)
		a.Pet.Level = petLevel(auc)
		a.Pet.QualityID = petQualityID(auc)
		a.Pet.SpeciesID = petSpeciesID(auc)
	}

	return a
}

// unpackAuctions converts the []interface{} format we get from the web into structs
func unpackAuctions(auctions []any) map[int64][]Auction {
	a := map[int64][]Auction{}

	for _, auc := range auctions {
		aucStruct := newAuction(auc.(map[string]any))
		if wowapi.BadItemID(aucStruct.ItemID) {
			// There are some item IDs for which WoW has no data. Ignore those.
			continue
		}
		if aucStruct.Buyout <= 0 {
			// These accept bids, but not purchases. Ignore these.
			continue
		}
		a[aucStruct.ItemID] = append(a[aucStruct.ItemID], aucStruct)
	}

	return a
}

// Get returns the current auctions binned by item ID
func Get(realm string) (map[int64][]Auction, error) {
	var ok bool
	var auctions []any

	if strings.ToLower(realm) == "commodities" {
		auctions, ok = wowapi.Commodities()
	} else {
		auctions, ok = wowapi.Auctions(realm)
	}
	if !ok {
		return nil, fmt.Errorf("unable to obtain auctions for: %s", realm)
	}

	return unpackAuctions(auctions), nil
}
