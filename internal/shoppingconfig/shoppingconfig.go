package shoppingconfig

import (
	"github.com/erikbryant/wow/internal/common"
	"github.com/erikbryant/wow/internal/cooking"
	"github.com/erikbryant/wow/internal/wowitem"
)

type UserConfig struct {
	AppearancePriceMax       int64
	AppearancePriceInSetMax  int64
	ArbitrageProfitMin       int64
	BattlePetPriceResellMax  int64
	BattlePetPriceUnownedMax int64
	ProfitToDisplayMin       int64
	RecipePriceMax           int64
	ToyPriceMax              int64

	UsefulGoods map[int64]int64
	SkipPets    map[int64]struct{}
}

func New(wi *wowitem.Persistence, cr *cooking.CookingRecipes) *UserConfig {
	//bagPriceMax := common.Coppers(100, 0, 0)
	//reagentBagPriceMax := common.Coppers(100, 0, 0)

	userConfig := UserConfig{
		AppearancePriceMax:       common.Coppers(50, 0, 0),
		AppearancePriceInSetMax:  common.Coppers(600, 0, 0),
		ArbitrageProfitMin:       common.Coppers(0, 50, 0),
		BattlePetPriceResellMax:  common.Coppers(80, 0, 0),
		BattlePetPriceUnownedMax: common.Coppers(500, 0, 0),
		ProfitToDisplayMin:       common.Coppers(15, 0, 0),
		RecipePriceMax:           common.Coppers(19, 0, 0),
		ToyPriceMax:              common.Coppers(400, 0, 0),

		// UsefulGoods are useful items I want, if the price is right
		// If the item name fails to match wi.Search will print a warning
		// the user can then come in here and fix the name.
		UsefulGoods: map[int64]int64{
			// Bags
			//wi.Search("Weavercloth Bag").ID():              bagPriceMax, // 34 slot
			//wi.Search("Azureweave Expedition Pack").ID():   bagPriceMax, // 34 slot
			//wi.Search("Imbued Bright Linen Backpack").ID(): bagPriceMax, // 36 slot
			//wi.Search("Duskweave Bag").ID():                bagPriceMax, // 36 slot
			//wi.Search("Sunfire Silk Backpack").ID():        bagPriceMax, // 38 slot

			// Reagent bags
			//wi.Search("Chronocloth Reagent Bag").ID():      reagentBagPriceMax, // 36 slot
			//wi.Search("Weavercloth Reagent Bag").ID():      reagentBagPriceMax, // 36 slot
			//wi.Search("Dawnweave Reagent Bag").ID():        reagentBagPriceMax, // 38 slot
			//wi.Search("Bright Linen Reagent Satchel").ID(): reagentBagPriceMax, // 38 slot
			//wi.Search("Arcanoweave Reagent Rucksack").ID(): reagentBagPriceMax, // 40 slot

			// Fun weapon appearances
			wi.Search("Blackfury").ID():                   common.Coppers(3000, 0, 0),
			wi.Search("Tyrhold Broadsword").ID():          common.Coppers(3000, 0, 0),
			wi.Search("Ameelton's Shot-Thrower").ID():     common.Coppers(3000, 0, 0),
			wi.Search("Kickback 5000").ID():               common.Coppers(3000, 0, 0),
			wi.Search("Extreme-Impact Hole Puncher").ID(): common.Coppers(3000, 0, 0),

			// Appearance set appearances
			wi.Search("Tyrhold Visage").ID():            common.Coppers(2000, 0, 0),
			wi.Search("Boots of the Black Flame").ID():  common.Coppers(2000, 0, 0),
			wi.Search("Helm of the Tranquil Path").ID(): common.Coppers(2000, 0, 0),
		},

		// SkipPets holds SpeciesID of pets that do not resell well
		SkipPets: map[int64]struct{}{
			1385: {}, // Albino Chimaeraling
			1706: {}, // Ashmaw Cub
			1150: {}, // Ashstone Core
			1934: {}, // Benax
			1964: {}, // Blood Boil
			1963: {}, // Boneshard
			4489: {}, // Bouncer
			4537: {}, // Chester
			1662: {}, // Cinder Pup
			2087: {}, // Cinderweb Recluse
			1149: {}, // Corefire Imp
			1205: {}, // Direhorn Runt
			119:  {}, // Father Winter's Helper
			1545: {}, // Firewing
			1442: {}, // Ghastly Kid
			1147: {}, // Harbinger of Flame
			2916: {}, // Hungry Burrower
			2089: {}, // Infernal Pyreclaw
			1687: {}, // Left Shark
			4647: {}, // Mr. DELVER
			1568: {}, // Puddle Terror
			1907: {}, // Pygmy Owl
			340:  {}, // Sea Pony
			162:  {}, // Sinister Squashling
			1628: {}, // Sister of Temptation
			200:  {}, // Spring Rabbit
			211:  {}, // Strand Crawler
			2088: {}, // Surger
			1434: {}, // Sun Sproutling
			1570: {}, // Sunfire Kaliri
			117:  {}, // Tiny Snowman
			251:  {}, // Toxic Wasteling
			118:  {}, // Winter Reindeer
			120:  {}, // Winter's Little Helper
			153:  {}, // Wolpertinger

			// We collect pets to sell to Stephen; limit how many of each we collect
			2842: {}, // Anomalus
			1965: {}, // Blightbreath
			191:  {}, // Clockwork Rocket Bot
			1802: {}, // Fetid Waveling
			1961: {}, // G0-R41-0N Ultratonk
			1233: {}, // Pocket Reaver
			3348: {}, // Primal Stormling
			1966: {}, // Soulbroken Whelpling
			3006: {}, // Stoneskin Dredwing Pup
			1151: {}, // Untamed Hatchling
			4506: {}, // Violet Sporbit
			1394: {}, // Weebomination
			4496: {}, // Wriggle
		},
	}

	// Add any recipes that the user needs
	for _, recipeName := range cr.RecipesNeeded() {
		userConfig.UsefulGoods[wi.Search(recipeName).ID()] = userConfig.RecipePriceMax
	}

	return &userConfig
}
