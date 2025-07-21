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
	passPhrase      = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	realms          = flag.String("realms", "Aegwynn,Agamaggan,Akama,Alexstrasza,Alleria,Altar of Storms,Andorhal,Anub'arak,Argent Dawn,Azgalor,Azjol-Nerub,Azuremyst,Baelgun,Blackhand,Blackwing Lair,Bloodhoof,Bloodscalp,Bronzebeard,Cairne,Coilfang,Darrowmere,Deathwing,Dentarg,Draenor,Dragonblight,Drak'thul,Durotan,Eitrigg,Elune,Farstriders,Feathermoon,Frostwolf,Ghostlands,Greymane,IceCrown,Kilrogg,Kirin Tor,Kul Tiras,Lightninghoof,Llane,Misha,Nazgrel,Ravencrest,Runetotem,Sisters of Elune,Commodities,Aggramar,Alterac Mountains,Azralon,Barthilas,Caelestrasz,Dath'Remar,Drakkari,Eredar,Goldrinn,Gundrak,Nemesis,Quel'Thalas,Ragnaros", "WoW realms to scan")
	untracked       = flag.Bool("untracked", false, "Scan all untracked items")
	untrackedRealms = "Gallywix"
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

// findBargains returns auctions for which the goods are below our desired prices
func findBargains(auctions map[int64][]auction.Auction) []string {
	bargains := []string{}

	// Generally useful items to keep a watch for
	goods := map[int64]int64{
		98715: common.Coins(5000, 0, 0), // Marked Flawless Battle-Stone
		92741: common.Coins(5000, 0, 0), // Flawless Battle-Stone

		114821: common.Coins(120, 0, 0), // Hexweave Bag (30 slot)

		//194019: common.Coins(90, 0, 0), // Simply Stitched Reagent Bag (32 slot)
		//194020: common.Coins(90, 0, 0), // Chronocloth Reagent Bag (36 slot)
		//222855: common.Coins(90, 0, 0), // Weavercloth Reagent Bag (36 slot)
		//222854: common.Coins(90, 0, 0), // Dawnweave Reagent Bag (38 slot)
	}

	// Toys I am not interested in
	skipToys := map[int64]bool{
		// Creepy
		101571: true, // Moonfang Shroud (creepy)

		// Too fiddly
		109167: true, // Findle's Loot-A-Rang

		// Only usable by engineers
		17716:  true, // Snowmaster 9000
		18984:  true, // Dimensional Ripper - Everlook
		30542:  true, // Dimensional Ripper - Area 52
		34060:  true, // Flying Machine
		34061:  true, // Turbo-Charged Flying Machine
		48933:  true, // Wormhole Generator: Northrend
		87215:  true, // Wormhole Generator: Pandaria
		112059: true, // Wormhole Centrifuge
		151652: true, // Wormhole Generator: Argus
		168807: true, // Wormhole Generator: Kul Tiras
		168808: true, // Wormhole Generator: Zandalar
		172924: true, // Wormhole Generator: Shadowlands
		198156: true, // Wyrmhole Generator: Dragon Isles
		221966: true, // Wormhole Generator: Khaz Algar

		// Just not interested
		116690: true, // Safari Lounge Cushion
		119212: true, // Winning Hand
		194057: true, // Cushion of Time Travel
		194058: true, // Cold Cushion
		198227: true, // Giggle Goggles
	}

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
			maxPrice := common.Coins(400, 0, 0)
			if i.Toy() && !toy.Own(i) && !skipToys[i.Id()] && auc.Buyout <= maxPrice {
				str := fmt.Sprintf("%s   %s", i.Name(), common.Gold(auc.Buyout))
				bargains = append(bargains, str)
			}

			// Bargains on specific goods
			maxPrice, ok := goods[itemId]
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

// findTransmogBargains returns auctions for which the goods are below our desired prices
func findTransmogBargains(auctions map[int64][]auction.Auction) []string {
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

	//foundAppearanceSet := false
	bargains := []string{}
	for _, candidate := range candidates {
		name := candidate.item.Name()
		if candidate.inAppearanceSet {
			name += "   " + common.Gold(candidate.price)
			//foundAppearanceSet = true
		}
		bargains = append(bargains, name)
	}

	//if len(bargains) < 5 && !foundAppearanceSet {
	//	return nil
	//}

	return bargains
}

// findPetBargains returns a list of pets that sell for more than they are listed
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

// findPetNeeded returns a list of pets I do not have
func findPetNeeded(auctions map[int64][]auction.Auction) []string {
	bargains := []string{}

	// Pets I do not own yet
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
		if petAuction.Buyout > common.Coins(2000, 0, 0) {
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

// findPetSpell returns a list of pet spells for sale
func findPetSpell(auctions map[int64][]auction.Auction) []string {
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

// findPetSpecialty returns a list of pets I am looking for
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

// printShoppingList prints a list of auctions the user may want to buy
func printShoppingList(label string, names []string, c *color.Color) {
	if len(names) == 0 {
		return
	}
	fmt.Printf("--- %s ---\n", label)
	c.Println(strings.Join(common.SortUnique(names), "\n"))
}

// scanRealm downloads the available auctions and prints any bargains/arbitrages
func scanRealm(realm string) {
	auctions, ok := auction.GetAuctions(realm)
	if !ok {
		return
	}

	arbitrages := findArbitrages(auctions)
	bargains := findBargains(auctions)
	petBargains := findPetBargains(auctions)
	petNeeded := findPetNeeded(auctions)
	petSpecialty := findPetSpecialty(auctions)
	petSpell := findPetSpell(auctions)
	transmogBargains := findTransmogBargains(auctions)

	cache.Save()

	if len(arbitrages) == 0 && len(bargains) == 0 && len(petBargains) == 0 && len(petNeeded) == 0 && len(petSpecialty) == 0 && len(petSpell) == 0 && len(transmogBargains) == 0 {
		// Skip realms that have nothing to buy
		return
	}

	c := color.New(color.FgCyan)
	c.Printf("\n===========>  %s (%d unique items)  <===========\n\n", realm, len(auctions))

	printShoppingList("Pet Needed", petNeeded, color.New(color.FgMagenta))
	printShoppingList("Pet Bargains", petBargains, color.New(color.FgGreen))
	printShoppingList("Arbitrages", arbitrages, color.New(color.FgWhite))
	printShoppingList("Bargains", bargains, color.New(color.FgRed))
	printShoppingList("Pet Specialty", petSpecialty, color.New(color.FgRed))
	printShoppingList("Pet Spell", petSpell, color.New(color.FgBlue))
	printShoppingList("Transmog Bargains", transmogBargains, color.New(color.FgBlue))
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

	if *passPhrase == "" {
		fmt.Println("ERROR: You must specify -passPhrase to unlock the client Id/secret")
		usage()
	}
	wowAPI.Init(*passPhrase, true)

	battlePet.Init()
	toy.Init()
	transmog.Init()

	realmsToScan := *realms
	if *untracked {
		realmsToScan = untrackedRealms
	}
	for _, realm := range strings.Split(realmsToScan, ",") {
		scanRealm(realm)
	}

	generateLua()
}
