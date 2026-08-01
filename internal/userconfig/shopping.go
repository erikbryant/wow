package userconfig

import (
	"github.com/erikbryant/wow/internal/common"
	"github.com/erikbryant/wow/internal/wowitem"
)

var (
	AppearancePriceMax       = common.Coppers(50, 0, 0)
	AppearancePriceInSetMax  = common.Coppers(600, 0, 0)
	ArbitrageProfitMin       = common.Coppers(0, 50, 0)
	BattlePetPriceResellMax  = common.Coppers(200, 0, 0)
	BattlePetPriceUnownedMax = common.Coppers(500, 0, 0)
	ProfitToDisplayMin       = common.Coppers(15, 0, 0)
	RecipePriceMax           = common.Coppers(20, 0, 0)
	ToyPriceMax              = common.Coppers(400, 0, 0)
)

// UsefulGoods are useful items I want, if the price is right
var UsefulGoods = map[int64]int64{
	// Bags
	//wowitem.Search("Hexweave Bag").ID(): common.Coppers(120, 0, 0), // 30 slot
	//wowitem.Search("Simply Stitched Reagent Bag").ID(): common.Coppers(90, 0, 0), // 32 slot
	//wowitem.Search("Chronocloth Reagent Bag").ID():     common.Coppers(90, 0, 0), // 36 slot
	//wowitem.Search("Weavercloth Reagent Bag").ID():     common.Coppers(90, 0, 0), // 36 slot
	//wowitem.Search("Dawnweave Reagent Bag").ID():       common.Coppers(90, 0, 0), // 38 slot

	// Fun weapon appearances
	wowitem.Search("Blackfury").ID():                   common.Coppers(3000, 0, 0),
	wowitem.Search("Tyrhold Broadsword").ID():          common.Coppers(3000, 0, 0),
	wowitem.Search("Ameelton's Shot-Thrower").ID():     common.Coppers(3000, 0, 0),
	wowitem.Search("Kickback 5000").ID():               common.Coppers(3000, 0, 0),
	wowitem.Search("Extreme-Impact Hole Puncher").ID(): common.Coppers(3000, 0, 0),

	// Appearance set appearances
	wowitem.Search("Tyrhold Visage").ID():            common.Coppers(2000, 0, 0),
	wowitem.Search("Tyrhold Slippers").ID():          common.Coppers(2000, 0, 0),
	wowitem.Search("Boots of the Black Flame").ID():  common.Coppers(2000, 0, 0),
	wowitem.Search("Helm of the Tranquil Path").ID(): common.Coppers(2000, 0, 0),
}

// SkipPets holds SpeciesID of pets that do not resell well
var SkipPets = map[int64]struct{}{
	1385: {}, // Albino Chimaeraling
	1706: {}, // Ashmaw Cub
	1150: {}, // Ashstone Core
	1934: {}, // Benax
	1964: {}, // Blood Boil
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

	// Pets that Stephen does not need right now
	1963: {}, // Boneshard
	191:  {}, // Clockwork Rocket Bot
	1961: {}, // G0-R41-0n Ultratonk
	2468: {}, // Laughing Stonekin
	1907: {}, // Pygmy Owl
	1721: {}, // Stormborne Whelpling
}
