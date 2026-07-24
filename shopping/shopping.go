package shopping

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/erikbryant/wow/auction"
	"github.com/erikbryant/wow/battlepet"
	"github.com/erikbryant/wow/common"
	"github.com/erikbryant/wow/item"
	"github.com/erikbryant/wow/itemcache"
	"github.com/erikbryant/wow/recipes"
	"github.com/erikbryant/wow/toy"
	"github.com/erikbryant/wow/transmog"
	"github.com/fatih/color"
)

var (
	mu             sync.Mutex
	usefulRecipes  = map[int64]struct{}{}
	oauthAvailable = true
)

const (
	usefulRecipesMaxPrice = 100000
)

// usefulGoods are useful items I want
var usefulGoods = map[int64]int64{
	//cache.Search("Hexweave Bag").Id():                 common.Coins(120, 0, 0), // 30 slot
	//cache.Search("Chronocloth Reagent Bag").Id():      common.Coins(90, 0, 0),  // 36 slot
	//cache.Search("Dawnweave Reagent Bag").Id():        common.Coins(90, 0, 0),  // 38 slot
	//cache.Search("Simply Stitched Reagent Bag").Id():  common.Coins(90, 0, 0),  // 32 slot
	//cache.Search("Weavercloth Reagent Bag").Id():      common.Coins(90, 0, 0),  // 36 slot

	// Fun weapon transmogs
	itemcache.Search("Blackfury").Id():          common.Coins(3000, 0, 0),
	itemcache.Search("Tyrhold Broadsword").Id(): common.Coins(3000, 0, 0),

	// Appearance set transmogs
	itemcache.Search("Tyrhold Visage").Id():            common.Coins(2000, 0, 0),
	itemcache.Search("Tyrhold Epaulets").Id():          common.Coins(2000, 0, 0),
	itemcache.Search("Tyrhold Robe").Id():              common.Coins(2000, 0, 0),
	itemcache.Search("Tyrhold Slippers").Id():          common.Coins(2000, 0, 0),
	itemcache.Search("Boots of the Black Flame").Id():  common.Coins(2000, 0, 0),
	itemcache.Search("Helm of the Tranquil Path").Id(): common.Coins(2000, 0, 0),
	itemcache.Search("Vest of the Tranquil Path").Id(): common.Coins(2000, 0, 0),

	// Gun appearances
	itemcache.Search("Ameelton's Shot-Thrower").Id():     common.Coins(3000, 0, 0),
	itemcache.Search("Kickback 5000").Id():               common.Coins(3000, 0, 0),
	itemcache.Search("Extreme-Impact Hole Puncher").Id(): common.Coins(3000, 0, 0),
}

// Init determines which classic cooking recipes are needed
func Init() {
	recipeNames := recipes.Needed()
	for _, recipeName := range recipeNames {
		usefulRecipes[itemcache.Search(recipeName).Id()] = struct{}{}
	}
}

// findPetSpellNeeded returns pet spells for sale that I do not own
func findPetSpellNeeded(auctions map[int64][]auction.Auction) []string {
	if !oauthAvailable {
		return nil
	}

	bargains := []string{}

	for itemId, itemAuctions := range auctions {
		i, ok := itemcache.LookupItem(itemId, 0)
		if !ok {
			continue
		}
		petId, ok := battlepet.IsPetSpell(i)
		if !ok {
			continue
		}
		//if common.QualityId(i.Quality()) < common.QualityId("Rare") {
		//	continue
		//}
		if battlepet.Own(petId) {
			continue
		}

		for _, auc := range itemAuctions {
			if auc.Buyout <= 0 {
				continue
			}
			if auc.Buyout >= common.Coins(800, 0, 0) {
				continue
			}
			stats := fmt.Sprintf("%s %s %s", battlepet.Name(petId), common.Gold(auc.Buyout), i.Quality())
			bargains = append(bargains, stats)
		}
	}

	return bargains
}

// findPetNeeded returns pets for sale that I do not own
func findPetNeeded(auctions map[int64][]auction.Auction) []string {
	if !oauthAvailable {
		return nil
	}

	bargains := []string{}

	for _, petAuction := range auctions[battlepet.PetCageItemId] {
		if battlepet.Own(petAuction.Pet.SpeciesId) {
			continue
		}
		if petAuction.Pet.SpeciesId == 3302 {
			// Pilot - we own this, but he is in the 'penalty box' for being so noisy
			continue
		}
		if petAuction.Buyout <= 0 {
			continue
		}
		if petAuction.Buyout > common.Coins(800, 0, 0) {
			continue
		}
		bargains = append(bargains, battlepet.Name(petAuction.Pet.SpeciesId))
	}

	// Include any pets available via spells
	spellBargains := findPetSpellNeeded(auctions)
	bargains = append(bargains, spellBargains...)

	return bargains
}

// findPetBargains returns pets that are likely to sell for more than they are listed
func findPetBargains(auctions map[int64][]auction.Auction) []string {
	bargains := []string{}

	// SpeciesId of pets that do not resell well
	skipPets := map[int64]struct{}{
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
	}

	for _, petAuction := range auctions[battlepet.PetCageItemId] {
		_, ok := skipPets[petAuction.Pet.SpeciesId]
		if ok {
			continue
		}
		if petAuction.Buyout <= 0 {
			continue
		}
		if petAuction.Pet.QualityId < common.QualityId("Rare") {
			continue
		}
		if petAuction.Pet.Level < 25 {
			continue
		}
		if petAuction.Buyout > common.Coins(100, 0, 0) {
			continue
		}

		bargains = append(bargains, battlepet.Name(petAuction.Pet.SpeciesId))
	}

	return bargains
}

type Arbitrage struct {
	item   item.Item
	profit int64
}

// findArbitrages returns auctions selling for lower than vendor prices
func findArbitrages(auctions map[int64][]auction.Auction, realm string) ([]string, int64) {
	arbitrages := []Arbitrage{}
	totalProfit := int64(0)

	for itemId, itemAuctions := range auctions {
		i, ok := itemcache.LookupItem(itemId, 0)
		if !ok {
			continue
		}
		for _, auc := range itemAuctions {
			if auc.Buyout <= 0 {
				continue
			}
			if auc.Buyout >= i.SellPriceRealizable() {
				continue
			}
			profit := (i.SellPriceRealizable() - auc.Buyout) * auc.Quantity
			if profit < common.Coins(0, 50, 0) {
				// Not enough profit to make it worth the WoW runtime it takes to scan the AH
				continue
			}

			arbitrages = append(arbitrages, Arbitrage{i, profit})

			if i.ItemClassName() == "Profession" && !item.Known(i.Id()) {
				// We have not seen this arbitrage before. Add iLevels for it in ilevel.go.
				msg := fmt.Sprintf("%d: {}, // %s (%s)  iLvl: %d\n", i.Id(), i.Name(), i.ItemClassName(), i.ItemLevel())
				common.AppendFile("./generated/arbitrageWithiLvl.log", msg, &mu)
				fmt.Println(msg)
			}
		}
	}

	bargains := []string{}
	for _, arbitrage := range arbitrages {
		totalProfit += arbitrage.profit

		str := fmt.Sprintf("%s   %s", arbitrage.item.Name(), common.Gold(arbitrage.profit))
		bargains = append(bargains, str)

		if realm == "Commodities" {
			// Commodities are not worth logging; their prices are too volatile
			continue
		}

		iLevels := item.ILevels(arbitrage.item.Id(), arbitrage.item.ItemLevel())
		for _, iLevel := range iLevels {
			logEntry := fmt.Sprintf("    {%d, %d}, -- %s\n", arbitrage.item.Id(), iLevel, arbitrage.item.Name())
			common.AppendFile("./generated/arbitrageLatest.log", logEntry, &mu)
		}
	}

	slices.Sort(bargains)
	bargains = slices.Compact(bargains)

	return bargains, totalProfit
}

// findBargains returns auctions selling below our desired prices
func findBargains(auctions map[int64][]auction.Auction) []string {
	bargains := []string{}

	for itemId, itemAuctions := range auctions {
		i, ok := itemcache.LookupItem(itemId, 0)
		if !ok {
			continue
		}
		for _, auc := range itemAuctions {
			if auc.Buyout <= 0 {
				continue
			}

			// Bargains on toys
			if oauthAvailable {
				maxPrice := common.Coins(400, 0, 0)
				if i.Toy() && !toy.Own(i) && auc.Buyout <= maxPrice {
					str := fmt.Sprintf("%s   %s", i.Name(), common.Gold(auc.Buyout))
					bargains = append(bargains, str)
				}
			}

			// Bargains on specific items
			maxPrice, ok := usefulGoods[itemId]
			if ok && auc.Buyout <= maxPrice {
				str := fmt.Sprintf("%s   %s", i.Name(), common.Gold(auc.Buyout))
				bargains = append(bargains, str)
			}

			// Bargains on recipes
			_, ok = usefulRecipes[itemId]
			if ok && auc.Buyout <= usefulRecipesMaxPrice {
				str := fmt.Sprintf("%s   %s", i.Name(), common.Gold(auc.Buyout))
				bargains = append(bargains, str)
			}
		}
	}

	return bargains
}

// findTransmogBargains returns transmog auctions selling below our desired price
func findTransmogBargains(auctions map[int64][]auction.Auction) []string {
	if !oauthAvailable {
		return nil
	}

	needed := map[string]bool{}

	for itemId, itemAuctions := range auctions {
		i, ok := itemcache.LookupItem(itemId, 0)
		if !ok {
			continue
		}
		for _, auc := range itemAuctions {
			if auc.Buyout <= 0 {
				continue
			}

			maxPrice := common.Coins(80, 0, 0)
			appearanceSetSuffix := ""
			if transmog.InAppearanceSet(i.Appearances()) {
				maxPrice = common.Coins(600, 0, 0)
				appearanceSetSuffix = "    ---"
			}

			if auc.Buyout > maxPrice {
				continue
			}

			if !transmog.NeedAppearance(i.Appearances()) {
				continue
			}

			needed[i.Name()+appearanceSetSuffix] = true
		}
	}

	bargains := []string{}
	for name := range needed {
		bargains = append(bargains, name)
	}

	return bargains
}

// fmtShoppingList returns a formatted string of the given items or "" if none
func fmtShoppingList(label string, items []string, c *color.Color, summarize bool) string {
	if len(items) == 0 {
		return ""
	}
	header := ""
	if !summarize {
		header = fmt.Sprintf("--- %s ---\n", label)
	}
	return c.Sprintf("%s%s\n", header, strings.Join(common.SortUnique(items), "\n"))
}

// scanRealm retrieves auctions and prints suggestions for what to buy for a single realm
func scanRealm(realm string, c chan<- string, summarize bool) {
	auctions, ok := auction.GetAuctions(realm)
	if !ok {
		c <- ""
		return
	}

	shoppingList := ""
	shoppingList += fmtShoppingList("Pets I Need", findPetNeeded(auctions), color.New(color.FgMagenta), summarize)
	shoppingList += fmtShoppingList("Pets to Resell", findPetBargains(auctions), color.New(color.FgGreen), summarize)
	shoppingList += fmtShoppingList("Useful Item Bargains", findBargains(auctions), color.New(color.FgRed), summarize)
	shoppingList += fmtShoppingList("Transmog Bargains", findTransmogBargains(auctions), color.New(color.FgBlue), summarize)

	arbitrages, profit := findArbitrages(auctions, realm)

	if summarize {
		if profit > common.Coins(15, 0, 0) {
			// Only show arbitrages if there is some actual amount of money
			// If the arbitrages are the only things on this realm, only show if worthwhile to visit
			c := color.New(color.FgWhite)
			shoppingList += c.Sprintf("Arbitrages: %s\n", common.Gold(profit))
		}
	} else {
		shoppingList += fmtShoppingList("Arbitrages", arbitrages, color.New(color.FgWhite), summarize)
	}

	if len(shoppingList) == 0 {
		// Nothing to buy
		c <- ""
		return
	}

	col := color.New(color.FgCyan)
	c <- col.Sprintf("\n===========>  %s (%d unique items)  <===========\n%s", realm, len(auctions), shoppingList)
}

// ScanRealms processes auctions on all realms in 'r'
func ScanRealms(r string, summarize, oauth bool) {
	oauthAvailable = oauth
	realms := strings.Split(r, ",")
	results := []string{}
	c := make(chan string)

	common.CreateFile("./generated/arbitrageLatest.log")

	for _, realm := range realms {
		go scanRealm(realm, c, summarize)
	}

	for range len(realms) {
		s := <-c
		if s == "" {
			continue
		}
		// Hack to get Commodities to sort to end of slice
		s = strings.Replace(s, " Commodities ", " _Commodities_ ", 1)
		results = append(results, s)
	}

	sort.Strings(results)
	fmt.Println(results)

	itemcache.Save()
}
