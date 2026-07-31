package shopping

import (
	"github.com/erikbryant/wow/internal/common"
	"github.com/erikbryant/wow/internal/wowitem"
)

var (
	appearancePriceMax       = common.Coppers(50, 0, 0)
	appearancePriceInSetMax  = common.Coppers(600, 0, 0)
	arbitrageProfitMin       = common.Coppers(0, 50, 0)
	battlePetPriceResellMax  = common.Coppers(200, 0, 0)
	battlePetPriceUnownedMax = common.Coppers(500, 0, 0)
	profitToDisplayMin       = common.Coppers(15, 0, 0)
	recipePriceMax           = common.Coppers(10, 0, 0)
	toyPriceMax              = common.Coppers(400, 0, 0)
)

// usefulGoods are useful items I want, if the price is right
var usefulGoods = map[int64]int64{
	// Bags
	//itempersistence.Search("Hexweave Bag").ID(): common.Coppers(120, 0, 0), // 30 slot
	//itempersistence.Search("Simply Stitched Reagent Bag").ID(): common.Coppers(90, 0, 0), // 32 slot
	//itempersistence.Search("Chronocloth Reagent Bag").ID():     common.Coppers(90, 0, 0), // 36 slot
	//itempersistence.Search("Weavercloth Reagent Bag").ID():     common.Coppers(90, 0, 0), // 36 slot
	//itempersistence.Search("Dawnweave Reagent Bag").ID():       common.Coppers(90, 0, 0), // 38 slot

	// Fun weapon transmogs
	wowitem.Search("Blackfury").ID():                   common.Coppers(3000, 0, 0),
	wowitem.Search("Tyrhold Broadsword").ID():          common.Coppers(3000, 0, 0),
	wowitem.Search("Ameelton's Shot-Thrower").ID():     common.Coppers(3000, 0, 0),
	wowitem.Search("Kickback 5000").ID():               common.Coppers(3000, 0, 0),
	wowitem.Search("Extreme-Impact Hole Puncher").ID(): common.Coppers(3000, 0, 0),

	// Appearance set transmogs
	wowitem.Search("Tyrhold Visage").ID():            common.Coppers(2000, 0, 0),
	wowitem.Search("Tyrhold Epaulets").ID():          common.Coppers(2000, 0, 0),
	wowitem.Search("Tyrhold Robe").ID():              common.Coppers(2000, 0, 0),
	wowitem.Search("Tyrhold Slippers").ID():          common.Coppers(2000, 0, 0),
	wowitem.Search("Boots of the Black Flame").ID():  common.Coppers(2000, 0, 0),
	wowitem.Search("Helm of the Tranquil Path").ID(): common.Coppers(2000, 0, 0),
}

// SpeciesID of pets that do not resell well
var skipPets = map[int64]struct{}{
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
