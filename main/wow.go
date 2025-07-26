package main

import (
	"flag"
	"fmt"
	"github.com/erikbryant/wow/auction"
	"github.com/erikbryant/wow/battlePet"
	"github.com/erikbryant/wow/cache"
	"github.com/erikbryant/wow/common"
	"github.com/erikbryant/wow/item"
	"github.com/erikbryant/wow/toy"
	"github.com/erikbryant/wow/transmog"
	"github.com/erikbryant/wow/wowAPI"
	"github.com/fatih/color"
	"log"
	"os"
	"strings"
)

var (
	passPhrase = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	// Linked realms still to populate: Gallywix
	realms         = flag.String("realms", "Aegwynn,Agamaggan,Akama,Alexstrasza,Alleria,Altar of Storms,Andorhal,Anub'arak,Argent Dawn,Azgalor,Azjol-Nerub,Azuremyst,Baelgun,Blackhand,Blackwing Lair,Bloodhoof,Bloodscalp,Bronzebeard,Cairne,Coilfang,Darrowmere,Deathwing,Dentarg,Draenor,Dragonblight,Drak'thul,Durotan,Eitrigg,Elune,Farstriders,Feathermoon,Frostwolf,Ghostlands,Greymane,IceCrown,Kilrogg,Kirin Tor,Kul Tiras,Lightninghoof,Llane,Misha,Nazgrel,Ravencrest,Runetotem,Sisters of Elune,Commodities,Aggramar,Alterac Mountains,Azralon,Barthilas,Caelestrasz,Dath'Remar,Drakkari,Eredar,Goldrinn,Gundrak,Nemesis,Quel'Thalas,Ragnaros", "WoW realm(s) to scan")
	oauthAvailable = flag.Bool("oauth", true, "Is OAuth authentication available?")
)

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

// usefulGoods are generally-useful items to keep an eye out for
var usefulGoods = map[int64]int64{
	cache.Search("Flawless Battle-Stone").Id():        common.Coins(5000, 0, 0),
	cache.Search("Marked Flawless Battle-Stone").Id(): common.Coins(5000, 0, 0),

	cache.Search("Hexweave Bag").Id(): common.Coins(120, 0, 0), // 30 slot
	//cache.Search("Chronocloth Reagent Bag").Id():     common.Coins(90, 0, 0), // 36 slot
	//cache.Search("Dawnweave Reagent Bag").Id():       common.Coins(90, 0, 0), // 38 slot
	//cache.Search("Simply Stitched Reagent Bag").Id(): common.Coins(90, 0, 0), // 32 slot
	//cache.Search("Weavercloth Reagent Bag").Id():     common.Coins(90, 0, 0), // 36 slot
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
	cache.Search("Cold Cushion").Id():           true,
	cache.Search("Cushion of Time Travel").Id(): true,
	cache.Search("Findle's Loot-A-Rang").Id():   true,
	cache.Search("Giggle Goggles").Id():         true,
	cache.Search("Moonfang Shroud").Id():        true,
	cache.Search("Safari Lounge Cushion").Id():  true,
	cache.Search("Winning Hand").Id():           true,
}

// findBargains returns auctions for which the items are below our desired prices
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
			if ok && auc.Buyout < maxPrice {
				str := fmt.Sprintf("%s   %s", i.Name(), common.Gold(auc.Buyout))
				bargains = append(bargains, str)
			}
		}
	}

	return bargains
}

type Candidate struct {
	item            item.Item
	price           int64
	inAppearanceSet bool
}

// findTransmogBargains returns auctions for which the transmog is below our desired price
func findTransmogBargains(auctions map[int64][]auction.Auction) []string {
	if !*oauthAvailable {
		return nil
	}

	candidates := map[int64]Candidate{}

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

			maxPrice := common.Coins(30, 0, 0)
			if transmog.InAppearanceSet(i) {
				maxPrice = common.Coins(40, 0, 0)
			}
			if auc.Buyout > maxPrice {
				continue
			}

			t := i.Appearances()
			if t == nil {
				continue
			}
			transmogId := t[0] // There may be multiple, but we'll just look at the first
			previous, ok := candidates[transmogId]
			if ok && auc.Buyout >= previous.price {
				continue
			}
			candidates[transmogId] = Candidate{
				i,
				auc.Buyout,
				transmog.InAppearanceSet(i),
			}
		}
	}

	bargains := []string{}
	for _, candidate := range candidates {
		name := candidate.item.Name()
		if candidate.inAppearanceSet {
			name += "   " + common.Gold(candidate.price)
		}
		bargains = append(bargains, name)
	}

	return bargains
}

// findPetBargains returns a list of pets that are likely to sell for more than they are listed
func findPetBargains(auctions map[int64][]auction.Auction) []string {
	bargains := []string{}

	// SpeciesId of pets that do not resell well
	skipPets := map[int64]bool{
		162: true, // Sinister Squashling
		191: true, // Clockwork Rocket Bot
		251: true, // Toxic Wasteling
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

// findPetNeeded returns any pets for sale that I do not own
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

	return bargains
}

var specialtyPets = map[int64]int64{
	// Pets that make good gifts
	//1890: common.Coins(1000, 0, 0), // Corgi Pup
	//1929: common.Coins(1000, 0, 0), // Corgnelius
}

// findPetSpellNeeded returns any pet spells for sale that I do not own
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
			if specialtyPets[petId] > 0 {
				stats := fmt.Sprintf("%s %s %s (specialty)", battlePet.Name(petId), common.Gold(auc.Buyout), i.Quality())
				bargains = append(bargains, stats)
			}
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

// findPetSpecialty returns a list of specialty pets I am looking for (whether I own them or not)
func findPetSpecialty(auctions map[int64][]auction.Auction) []string {
	bargains := []string{}

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

// fmtShoppingList returns a formatted string of the given items or "" if none
func fmtShoppingList(label string, items []string, c *color.Color) string {
	if len(items) == 0 {
		return ""
	}
	return c.Sprintf("--- %s ---\n%s\n", label, strings.Join(common.SortUnique(items), "\n"))
}

// scanRealm retrieves auctions and prints any bargains/arbitrages to go buy
func scanRealm(realm string) {
	auctions, ok := auction.GetAuctions(realm)
	if !ok {
		return
	}

	cache.Save()

	results := ""
	results += fmtShoppingList("Pet Needed", findPetNeeded(auctions), color.New(color.FgMagenta))
	results += fmtShoppingList("Pet Bargains", findPetBargains(auctions), color.New(color.FgGreen))
	results += fmtShoppingList("Arbitrages", findArbitrages(auctions), color.New(color.FgWhite))
	results += fmtShoppingList("Bargains", findBargains(auctions), color.New(color.FgRed))
	results += fmtShoppingList("Pet Specialty", findPetSpecialty(auctions), color.New(color.FgRed))
	results += fmtShoppingList("Pet Spell", findPetSpellNeeded(auctions), color.New(color.FgBlue))
	results += fmtShoppingList("Transmog Bargains", findTransmogBargains(auctions), color.New(color.FgBlue))

	if len(results) == 0 {
		// Nothing to buy
		return
	}

	c := color.New(color.FgCyan)
	c.Printf("\n===========>  %s (%d unique items)  <===========\n\n%s", realm, len(auctions), results)
}

// writeFile writes contents to file
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

// generateLua writes the WoW addon lua files
func generateLua() {
	writeFile("./generated/PriceCache.lua", cache.LuaVendorPrice())
	writeFile("./generated/PetCache.lua", battlePet.LuaPetId())
}

// usage prints a usage message and terminates the program with an error
func usage() {
	log.Fatal(`Usage:
  wow -passPhrase <phrase> [-realmsUS|-realms <realms>]
`)
}

func main() {
	flag.Parse()

	fmt.Println(*oauthAvailable)

	if *passPhrase == "" {
		fmt.Println("ERROR: You must specify -passPhrase to unlock the client Id/secret")
		usage()
	}
	wowAPI.Init(*passPhrase, *oauthAvailable)

	battlePet.Init(*oauthAvailable)
	if *oauthAvailable {
		toy.Init()
		transmog.Init()
	}

	for _, realm := range strings.Split(*realms, ",") {
		scanRealm(realm)
	}

	if *oauthAvailable {
		generateLua()
	}
}
