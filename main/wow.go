package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/erikbryant/wow/auction"
	"github.com/erikbryant/wow/battlePet"
	"github.com/erikbryant/wow/common"
	"github.com/erikbryant/wow/item"
	"github.com/erikbryant/wow/itemCache"
	"github.com/erikbryant/wow/toy"
	"github.com/erikbryant/wow/transmog"
	"github.com/erikbryant/wow/wowAPI"
	"github.com/fatih/color"
)

var (
	mu             sync.Mutex
	passPhrase     = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	realms         = flag.String("realms", "Aegwynn,Agamaggan,Aggramar,Akama,Alexstrasza,Alleria,Altar of Storms,Alterac Mountains,Andorhal,Anub'arak,Argent Dawn,Azgalor,Azjol-Nerub,Azralon,Azuremyst,Baelgun,Barthilas,Blackhand,Blackwing Lair,Bloodhoof,Bloodscalp,Bronzebeard,Caelestrasz,Cairne,Coilfang,Darrowmere,Dath'Remar,Deathwing,Dentarg,Draenor,Dragonblight,Drak'thul,Drakkari,Durotan,Eitrigg,Elune,Eredar,Farstriders,Feathermoon,Frostwolf,Gallywix,Ghostlands,Goldrinn,Greymane,Gundrak,IceCrown,Kilrogg,Kirin Tor,Kul Tiras,Lightninghoof,Llane,Misha,Nazgrel,Nemesis,Quel'Thalas,Ragnaros,Ravencrest,Runetotem,Sisters of Elune,Commodities", "WoW realm(s) to scan")
	oauthAvailable = flag.Bool("oauth", true, "Is OAuth authentication available?")
	petResell      = flag.Bool("petResell", true, "Suggest pets to resell?")
	summarize      = flag.Bool("summarize", true, "Summarize arbitrages?")
)

// usefulGoods are useful items I want
var usefulGoods = map[int64]int64{
	itemCache.Search("Flawless Battle-Stone").Id(): common.Coins(300, 0, 0),

	//cache.Search("Hexweave Bag").Id():                 common.Coins(120, 0, 0), // 30 slot
	//cache.Search("Chronocloth Reagent Bag").Id():      common.Coins(90, 0, 0),  // 36 slot
	//cache.Search("Dawnweave Reagent Bag").Id():        common.Coins(90, 0, 0),  // 38 slot
	//cache.Search("Simply Stitched Reagent Bag").Id():  common.Coins(90, 0, 0),  // 32 slot
	//cache.Search("Weavercloth Reagent Bag").Id():      common.Coins(90, 0, 0),  // 36 slot

	//itemCache.Search("Xiwyllag ATV").Id(): common.Coins(3999, 0, 0),

	// Gun appearances
	itemCache.Search("Ameelton's Shot-Thrower").Id():  common.Coins(1000, 0, 0),
	itemCache.Search("Converted Carrot Carbine").Id(): common.Coins(1000, 0, 0),
	itemCache.Search("Kickback 5000").Id():            common.Coins(1000, 0, 0),

	// Appearance set appearances
	itemCache.Search("Tyrhold Epaulets").Id(): common.Coins(1000, 0, 0),
	itemCache.Search("Tyrhold Robe").Id():     common.Coins(1000, 0, 0),
	itemCache.Search("Tyrhold Slippers").Id(): common.Coins(1000, 0, 0),
	itemCache.Search("Tyrhold Visage").Id():   common.Coins(1000, 0, 0),

	itemCache.Search("Starless Mantle").Id(): common.Coins(5000, 0, 0),

	itemCache.Search("Shadowghast Breastplate").Id(): common.Coins(1000, 0, 0),
	itemCache.Search("Shadowghast Helm").Id():        common.Coins(1000, 0, 0),
	itemCache.Search("Shadowghast Pauldrons").Id():   common.Coins(1000, 0, 0),
}

// skipToys are toys I am not interested in
var skipToys = map[int64]bool{
	// Only usable by engineers
	itemCache.Search("Dimensional Ripper - Area 52").Id():     true,
	itemCache.Search("Dimensional Ripper - Everlook").Id():    true,
	itemCache.Search("Flying Machine").Id():                   true,
	itemCache.Search("Snowmaster 9000").Id():                  true,
	itemCache.Search("Turbo-Charged Flying Machine").Id():     true,
	itemCache.Search("Wormhole Centrifuge").Id():              true,
	itemCache.Search("Wormhole Generator: Argus").Id():        true,
	itemCache.Search("Wormhole Generator: Khaz Algar").Id():   true,
	itemCache.Search("Wormhole Generator: Kul Tiras").Id():    true,
	itemCache.Search("Wormhole Generator: Northrend").Id():    true,
	itemCache.Search("Wormhole Generator: Pandaria").Id():     true,
	itemCache.Search("Wormhole Generator: Quel'Thalas").Id():  true,
	itemCache.Search("Wormhole Generator: Shadowlands").Id():  true,
	itemCache.Search("Wormhole Generator: Zandalar").Id():     true,
	itemCache.Search("Wyrmhole Generator: Dragon Isles").Id(): true,

	// I am not interested in these
	itemCache.Search("Artisan's Sign").Id():         true,
	itemCache.Search("Cold Cushion").Id():           true,
	itemCache.Search("Cushion of Time Travel").Id(): true,
	itemCache.Search("Giggle Goggles").Id():         true,
	itemCache.Search("Leather Pet Bed").Id():        true,
	itemCache.Search("Leather Pet Leash").Id():      true,
	itemCache.Search("Moonfang Shroud").Id():        true,
	itemCache.Search("Safari Lounge Cushion").Id():  true,
	itemCache.Search("Winning Hand").Id():           true,
}

// findPetSpellNeeded returns pet spells for sale that I do not own
func findPetSpellNeeded(auctions map[int64][]auction.Auction) []string {
	if !*oauthAvailable {
		return nil
	}

	bargains := []string{}

	for itemId, itemAuctions := range auctions {
		i, ok := wowAPI.LookupItem(itemId, 0)
		if !ok {
			continue
		}
		petId, ok := battlePet.IsPetSpell(i)
		if !ok {
			continue
		}
		//if common.QualityId(i.Quality()) < common.QualityId("Rare") {
		//	continue
		//}
		if battlePet.Own(petId) {
			continue
		}

		for _, auc := range itemAuctions {
			if auc.Buyout <= 0 {
				continue
			}
			if auc.Buyout >= common.Coins(1000, 0, 0) {
				continue
			}
			stats := fmt.Sprintf("%s %s %s", battlePet.Name(petId), common.Gold(auc.Buyout), i.Quality())
			bargains = append(bargains, stats)
		}
	}

	return bargains
}

// findPetNeeded returns pets for sale that I do not own
func findPetNeeded(auctions map[int64][]auction.Auction) []string {
	if !*oauthAvailable {
		return nil
	}

	bargains := []string{}

	for _, petAuction := range auctions[battlePet.PetCageItemId] {
		if battlePet.Own(petAuction.Pet.SpeciesId) {
			continue
		}
		if petAuction.Pet.SpeciesId == 3302 {
			// Pilot - we own this, but he is in the 'penalty box' for being so noisy
			continue
		}
		if petAuction.Buyout <= 0 {
			continue
		}
		if petAuction.Buyout > common.Coins(1000, 0, 0) {
			continue
		}
		bargains = append(bargains, battlePet.Name(petAuction.Pet.SpeciesId))
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
	skipPets := map[int64]bool{
		1385: true, // Albino Chimaeraling
		1150: true, // Ashstone Core
		1934: true, // Benax
		1964: true, // Blood Boil
		4489: true, // Bouncer
		4537: true, // Chester
		1662: true, // Cinder Pup
		2087: true, // Cinderweb Recluse
		1149: true, // Corefire Imp
		1205: true, // Direhorn Runt
		119:  true, // Father Winter's Helper
		1545: true, // Firewing
		1442: true, // Ghastly Kid
		1147: true, // Harbinger of Flame
		2916: true, // Hungry Burrower
		1687: true, // Left Shark
		4647: true, // Mr. DELVER
		1568: true, // Puddle Terror
		162:  true, // Sinister Squashling
		200:  true, // Spring Rabbit
		211:  true, // Strand Crawler
		2088: true, // Surger
		1434: true, // Sun Sproutling
		1570: true, // Sunfire Kaliri
		117:  true, // Tiny Snowman
		251:  true, // Toxic Wasteling
		118:  true, // Winter Reindeer
		120:  true, // Winter's Little Helper
	}

	for _, petAuction := range auctions[battlePet.PetCageItemId] {
		if skipPets[petAuction.Pet.SpeciesId] {
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

		bargains = append(bargains, battlePet.Name(petAuction.Pet.SpeciesId))
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
		i, ok := wowAPI.LookupItem(itemId, 0)
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

			if i.VariableItemLevel() && !item.Known(i.Id()) {
				// We have not seen this arbitrage before. Add iLevels for it in iLevel.go.
				msg := fmt.Sprintf("%d: {}, // %s (%s)  iLvl: %d\n", i.Id(), i.Name(), i.ItemClassName(), i.ItemLevel())
				appendFile("./generated/arbitrageWithiLvl.log", msg)
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
			appendFile("./generated/arbitrageLatest.log", logEntry)
		}
	}

	return bargains, totalProfit
}

// findBargains returns auctions selling below our desired prices
func findBargains(auctions map[int64][]auction.Auction) []string {
	bargains := []string{}

	for itemId, itemAuctions := range auctions {
		i, ok := wowAPI.LookupItem(itemId, 0)
		if !ok {
			continue
		}
		for _, auc := range itemAuctions {
			if auc.Buyout <= 0 {
				continue
			}

			// Bargains on toys
			if *oauthAvailable {
				maxPrice := common.Coins(400, 0, 0)
				if i.Toy() && !toy.Own(i) && !skipToys[i.Id()] && auc.Buyout <= maxPrice {
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
		}
	}

	return bargains
}

// findTransmogBargains returns transmog auctions selling below our desired price
func findTransmogBargains(auctions map[int64][]auction.Auction) []string {
	if !*oauthAvailable {
		return nil
	}

	needed := map[string]bool{}

	for itemId, itemAuctions := range auctions {
		i, ok := wowAPI.LookupItem(itemId, 0)
		if !ok {
			continue
		}
		for _, auc := range itemAuctions {
			if auc.Buyout <= 0 {
				continue
			}

			if !transmog.NeedItem(i) {
				continue
			}

			maxPrice := common.Coins(20, 0, 0)
			appearanceSetSuffix := ""
			if transmog.InAppearanceSet(i) {
				maxPrice = common.Coins(300, 0, 0)
				appearanceSetSuffix = "    ---"
			}

			if auc.Buyout > maxPrice {
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
func scanRealm(realm string, c chan<- string, summarize, includePets bool) {
	auctions, ok := auction.GetAuctions(realm)
	if !ok {
		c <- ""
		return
	}

	shoppingList := ""
	shoppingList += fmtShoppingList("Pets I Need", findPetNeeded(auctions), color.New(color.FgMagenta), summarize)
	if includePets {
		shoppingList += fmtShoppingList("Pets to Resell", findPetBargains(auctions), color.New(color.FgGreen), summarize)
	}
	shoppingList += fmtShoppingList("Useful Item Bargains", findBargains(auctions), color.New(color.FgRed), summarize)
	shoppingList += fmtShoppingList("Transmog Bargains", findTransmogBargains(auctions), color.New(color.FgBlue), summarize)

	arbitrages, profit := findArbitrages(auctions, realm)

	if summarize {
		if profit > common.Coins(20, 0, 0) {
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

// scanRealms processes auctions on all realms in 'r'
func scanRealms(r string, summarize, includePets bool) {
	realms := strings.Split(r, ",")
	results := []string{}
	c := make(chan string)

	for _, realm := range realms {
		go scanRealm(realm, c, summarize, includePets)
	}

	err := os.Remove("./generated/arbitrageLatest.log")
	if err != nil {
		fmt.Println(err)
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

	itemCache.Save()
}

// appendFile appends 'contents' to a file
func appendFile(file, contents string) {
	mu.Lock()
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal("Failed to open file:", file, err)
	}
	defer f.Close()

	_, err = f.WriteString(contents)
	if err != nil {
		log.Fatal("Failed to append file:", file, err)
	}

	err = f.Close()
	if err != nil {
		log.Fatal("Failed to close file:", file, err)
	}
	mu.Unlock()
}

// writeFile writes 'contents' to a new file
func writeFile(file, contents string) {
	f, err := os.Create(file)
	if err != nil {
		log.Fatal("Failed to create file:", file, err)
	}
	defer f.Close()

	_, err = f.WriteString(contents)
	if err != nil {
		log.Fatal("Failed to write file:", file, err)
	}

	err = f.Close()
	if err != nil {
		log.Fatal("Failed to close file:", file, err)
	}
}

// generateLua writes the WoW 'Arbitrage' addon Lua files
func generateLua() {
	writeFile("./generated/PriceCache.lua", itemCache.LuaVendorPrice())
}

// usage prints a usage message and terminates the program with an error
func usage() {
	log.Fatal(`Usage:
  wow -passPhrase <phrase> [-realmsUS|-realms <realms>]
`)
}

func main() {
	flag.Parse()

	if *passPhrase == "" {
		fmt.Println("ERROR: You must specify -passPhrase to unlock the client Id/secret")
		usage()
	}

	wowAPI.Init(*passPhrase, *oauthAvailable)
	battlePet.Init(*oauthAvailable)
	toy.Init(*oauthAvailable)
	transmog.Init(*oauthAvailable)

	if !*oauthAvailable {
		fmt.Printf("\n*** OAuth unavailable. Some features may be missing.\n")
	}

	scanRealms(*realms, *summarize, *petResell)

	if *oauthAvailable {
		generateLua()
	}
}
