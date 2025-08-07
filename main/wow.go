package main

import (
	"flag"
	"fmt"
	"github.com/erikbryant/wow/auction"
	"github.com/erikbryant/wow/battlePet"
	"github.com/erikbryant/wow/cache"
	"github.com/erikbryant/wow/common"
	"github.com/erikbryant/wow/toy"
	"github.com/erikbryant/wow/transmog"
	"github.com/erikbryant/wow/wowAPI"
	"github.com/fatih/color"
	"log"
	"os"
	"sort"
	"strings"
)

var (
	passPhrase     = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	realms         = flag.String("realms", "Aegwynn,Agamaggan,Aggramar,Akama,Alexstrasza,Alleria,Altar of Storms,Alterac Mountains,Andorhal,Anub'arak,Argent Dawn,Azgalor,Azjol-Nerub,Azralon,Azuremyst,Baelgun,Barthilas,Blackhand,Blackwing Lair,Bloodhoof,Bloodscalp,Bronzebeard,Caelestrasz,Cairne,Coilfang,Darrowmere,Dath'Remar,Deathwing,Dentarg,Draenor,Dragonblight,Drak'thul,Drakkari,Durotan,Eitrigg,Elune,Eredar,Farstriders,Feathermoon,Frostwolf,Gallywix,Ghostlands,Goldrinn,Greymane,Gundrak,IceCrown,Kilrogg,Kirin Tor,Kul Tiras,Lightninghoof,Llane,Misha,Nazgrel,Nemesis,Quel'Thalas,Ragnaros,Ravencrest,Runetotem,Sisters of Elune,Commodities", "WoW realm(s) to scan")
	oauthAvailable = flag.Bool("oauth", true, "Is OAuth authentication available?")
)

// usefulGoods are useful items I want
var usefulGoods = map[int64]int64{
	cache.Search("Flawless Battle-Stone").Id():        common.Coins(4000, 0, 0),
	cache.Search("Marked Flawless Battle-Stone").Id(): common.Coins(4000, 0, 0),
	cache.Search("Hexweave Bag").Id():                 common.Coins(120, 0, 0), // 30 slot
	cache.Search("Chronocloth Reagent Bag").Id():      common.Coins(90, 0, 0),  // 36 slot
	cache.Search("Dawnweave Reagent Bag").Id():        common.Coins(90, 0, 0),  // 38 slot
	cache.Search("Simply Stitched Reagent Bag").Id():  common.Coins(90, 0, 0),  // 32 slot
	cache.Search("Weavercloth Reagent Bag").Id():      common.Coins(90, 0, 0),  // 36 slot
}

// skipToys are toys I am not interested in
var skipToys = map[int64]bool{
	// Only usable by engineers
	cache.Search("Dimensional Ripper - Area 52").Id():     true,
	cache.Search("Dimensional Ripper - Everlook").Id():    true,
	cache.Search("Flying Machine").Id():                   true,
	cache.Search("Snowmaster 9000").Id():                  true,
	cache.Search("Turbo-Charged Flying Machine").Id():     true,
	cache.Search("Wormhole Centrifuge").Id():              true,
	cache.Search("Wormhole Generator: Argus").Id():        true,
	cache.Search("Wormhole Generator: Khaz Algar").Id():   true,
	cache.Search("Wormhole Generator: Kul Tiras").Id():    true,
	cache.Search("Wormhole Generator: Northrend").Id():    true,
	cache.Search("Wormhole Generator: Pandaria").Id():     true,
	cache.Search("Wormhole Generator: Shadowlands").Id():  true,
	cache.Search("Wormhole Generator: Zandalar").Id():     true,
	cache.Search("Wyrmhole Generator: Dragon Isles").Id(): true,

	// I am not interested in these
	cache.Search("Artisan's Sign").Id():         true,
	cache.Search("Cold Cushion").Id():           true,
	cache.Search("Cushion of Time Travel").Id(): true,
	cache.Search("Findle's Loot-A-Rang").Id():   true,
	cache.Search("Giggle Goggles").Id():         true,
	cache.Search("Moonfang Shroud").Id():        true,
	cache.Search("Safari Lounge Cushion").Id():  true,
	cache.Search("Winning Hand").Id():           true,
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
		if common.QualityId(i.Quality()) < common.QualityId("Rare") {
			continue
		}
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
		if petAuction.Buyout <= 0 {
			continue
		}
		if petAuction.Pet.QualityId < common.QualityId("Rare") {
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
		162: true, // Sinister Squashling
		//191:  true, // Clockwork Rocket Bot
		//251:  true, // Toxic Wasteling
		4489: true, // Bouncer
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

// findArbitrages returns auctions selling for lower than vendor prices
func findArbitrages(auctions map[int64][]auction.Auction) []string {
	arbitrages := map[string]int64{}

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
			arbitrages[i.Name()] += (i.SellPriceRealizable() - auc.Buyout) * auc.Quantity
		}
	}

	bargains := []string{}
	for name, profit := range arbitrages {
		if profit < common.Coins(3, 0, 0) {
			// Too small to bother with
			continue
		}
		str := fmt.Sprintf("%s   %s", name, common.Gold(profit))
		bargains = append(bargains, str)
	}

	return bargains
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

// findPetSpecialty returns specialty pets I am looking for (whether I own them or not)
func findPetSpecialty(auctions map[int64][]auction.Auction) []string {
	bargains := []string{}

	var specialtyPets = map[int64]int64{
		// Pets that make good gifts
		//1890: common.Coins(1000, 0, 0), // Corgi Pup
		//1929: common.Coins(1000, 0, 0), // Corgnelius
	}

	for _, petAuction := range auctions[battlePet.PetCageItemId] {
		if petAuction.Buyout <= 0 {
			continue
		}
		premiumPetPrice := specialtyPets[petAuction.Pet.SpeciesId]
		if petAuction.Buyout > premiumPetPrice {
			continue
		}

		namePrice := fmt.Sprintf("%s   %d   %s", battlePet.Name(petAuction.Pet.SpeciesId), petAuction.Pet.QualityId, common.Gold(premiumPetPrice))
		bargains = append(bargains, namePrice)
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

			maxPrice := common.Coins(25, 0, 0)
			suffix := ""
			if transmog.InAppearanceSet(i) {
				maxPrice = common.Coins(90, 0, 0)
				suffix = "    ---"
			}

			if auc.Buyout > maxPrice {
				continue
			}

			needed[i.Name()+suffix] = true
		}
	}

	bargains := []string{}
	for name, _ := range needed {
		bargains = append(bargains, name)
	}

	return bargains
}

// fmtShoppingList returns a formatted string of the given items or "" if none
func fmtShoppingList(label string, items []string, c *color.Color) string {
	if len(items) == 0 {
		return ""
	}
	return c.Sprintf("--- %s ---\n%s\n", label, strings.Join(common.SortUnique(items), "\n"))
}

// scanRealm retrieves auctions and prints suggestions for what to buy
func scanRealm(realm string, c chan<- string) {
	auctions, ok := auction.GetAuctions(realm)
	if !ok {
		c <- ""
		return
	}

	results := ""
	results += fmtShoppingList("Pets I Need", findPetNeeded(auctions), color.New(color.FgMagenta))
	results += fmtShoppingList("Pets to Resell", findPetBargains(auctions), color.New(color.FgGreen))
	results += fmtShoppingList("Arbitrages", findArbitrages(auctions), color.New(color.FgWhite))
	results += fmtShoppingList("Useful Item Bargains", findBargains(auctions), color.New(color.FgRed))
	results += fmtShoppingList("Specialty Pets", findPetSpecialty(auctions), color.New(color.FgRed))
	results += fmtShoppingList("Transmog Bargains", findTransmogBargains(auctions), color.New(color.FgBlue))

	if len(results) == 0 {
		// Nothing to buy
		c <- ""
		return
	}

	col := color.New(color.FgCyan)
	c <- col.Sprintf("\n===========>  %s (%d unique items)  <===========\n\n%s", realm, len(auctions), results)
}

func scanRealms(r string) {
	realms := strings.Split(r, ",")
	results := []string{}
	c := make(chan string)

	for _, realm := range realms {
		go scanRealm(realm, c)
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

	cache.Save()
}

// writeFile writes 'contents' to file
func writeFile(file, contents string) {
	f, err := os.Create(file)
	if err != nil {
		log.Fatal("Failed to create file:", file, err)
	}
	_, err = f.WriteString(contents)
	err = f.Close()
	if err != nil {
		fmt.Println("Failed to close file:", file, err)
	}
}

// generateLua writes the WoW 'Arbitrage' addon lua files
func generateLua() {
	writeFile("./generated/PriceCache.lua", cache.LuaVendorPrice())
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

	scanRealms(*realms)

	if *oauthAvailable {
		generateLua()
	}
}
