package auction

import (
	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/common"
	"github.com/erikbryant/wow/item"
	"github.com/erikbryant/wow/wowAPI"
	"log"
	"strings"
)

// Sample 'commodity auction' response. All have exactly these fields.
// map[id:3.44371058e+08 item:map[id:192672] quantity:1 time_left:SHORT unit_price:16800]

// Sample 'auction' response. Some have more or fewer fields.
// map[buyout:1.1111011e+09 id:3.49632108e+08 item:map[id:142075] quantity:1 time_left:VERY_LONG]

// Sample 'auction' response for a pet auction. ItemId 82800 is a 'Pet Cage'. Pet cages have no sell value.
// map[buyout:9.99e+06 id:5.01784773e+08 item:map[id:82800 modifiers:[map[type:6 value:39130]] pet_breed_id:20 pet_level:1 pet_quality_id:2 pet_species_id:1446] quantity:1 time_left:VERY_LONG]

// Auction contains the properties of a single auction house auction
type Auction struct {
	Id       int64
	ItemId   int64
	Buyout   int64 // For commodity auctions this stores 'unit_price'
	Quantity int64
	Pet      item.PetInfo
}

func Id(raw interface{}) int64 {
	return web.ToInt64(common.MSIValue(raw, []string{"id"}))
}

func ItemId(raw interface{}) int64 {
	return web.ToInt64(common.MSIValue(raw, []string{"item", "id"}))
}

func Buyout(raw interface{}) int64 {
	value := common.MSIValue(raw, []string{"buyout"})
	if value == nil {
		value = common.MSIValue(raw, []string{"unit_price"})
		if value == nil {
			log.Println("could not find 'buyout' nor 'unit_price'")
			return 0
		}
	}
	return web.ToInt64(value)
}

func Quantity(raw interface{}) int64 {
	return web.ToInt64(common.MSIValue(raw, []string{"quantity"}))
}

func PetBreedId(raw interface{}) int64 {
	return web.ToInt64(common.MSIValue(raw, []string{"item", "pet_breed_id"}))
}

func PetLevel(raw interface{}) int64 {
	return web.ToInt64(common.MSIValue(raw, []string{"item", "pet_level"}))
}

func PetQualityId(raw interface{}) int64 {
	return web.ToInt64(common.MSIValue(raw, []string{"item", "pet_quality_id"}))
}

func PetSpeciesId(raw interface{}) int64 {
	return web.ToInt64(common.MSIValue(raw, []string{"item", "pet_species_id"}))
}

// JsonToStruct converts a single auction json string into a struct that is much easier to work with
func JsonToStruct(auc interface{}) Auction {
	var a Auction

	a.Id = Id(auc)
	a.ItemId = ItemId(auc)
	a.Buyout = Buyout(auc)
	a.Quantity = Quantity(auc)

	// Is this a Pet Cage?
	if a.ItemId == 82800 {
		// A pet auction!
		a.Pet.BreedId = PetBreedId(auc)
		a.Pet.Level = PetLevel(auc)
		a.Pet.QualityId = PetQualityId(auc)
		a.Pet.SpeciesId = PetSpeciesId(auc)
	}

	return a
}

// UnpackAuctions converts the []interface{} format we get from the web into structs
func UnpackAuctions(auctions []interface{}) map[int64][]Auction {
	a := map[int64][]Auction{}

	for _, auc := range auctions {
		aucStruct := JsonToStruct(auc.(map[string]interface{}))
		if wowAPI.SkipItem(aucStruct.ItemId) {
			continue
		}
		a[aucStruct.ItemId] = append(a[aucStruct.ItemId], aucStruct)
	}

	return a
}

// GetAuctions returns the current auctions and their hash
func GetAuctions(realm, accessToken string) (map[int64][]Auction, bool) {
	var ok bool
	var auctions []interface{}

	if strings.ToLower(realm) == "commodities" {
		auctions, ok = wowAPI.Commodities(accessToken)
	} else {
		auctions, ok = wowAPI.Auctions(realm, accessToken)
	}
	if !ok {
		log.Println("ERROR: Unable to obtain auctions for", realm)
		return nil, false
	}

	return UnpackAuctions(auctions), true
}
