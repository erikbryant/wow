package main

import (
	"flag"
	"fmt"
	"github.com/erikbryant/wow/auction"
	"github.com/erikbryant/wow/battlePet"
	"github.com/erikbryant/wow/cache"
	"github.com/erikbryant/wow/common"
	"github.com/erikbryant/wow/wowAPI"
	"github.com/fatih/color"
	"log"
	"os"
	"strings"
)

var (
	passPhrase = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	// These realms are distinct from the others, but are currently full: Aggramar,Alterac Mountains,Eredar
	realms = flag.String("realms", "Aegwynn,Agamaggan,Akama,Alexstrasza,Altar of Storms,Andorhal,Anub'arak,Argent Dawn,Azgalor,Azjol-Nerub,Azuremyst,Baelgun,Blackhand,Blackwing Lair,Bloodhoof,Bloodscalp,Bronzebeard,Cairne,Coilfang,Darrowmere,Deathwing,Dentarg,Draenor,Dragonblight,Drak'thul,Durotan,Eitrigg,Elune,Farstriders,Feathermoon,Frostwolf,Ghostlands,Greymane,IceCrown,Kilrogg,Kul Tiras,Llane,Misha,Nazgrel,Ravencrest,Runetotem,Sisters of Elune,Commodities,Alleria,Kirin Tor,Lightninghoof", "WoW realms to scan")
)

// findArbitrages returns auctions selling for lower than vendor prices
func findArbitrages(auctions map[int64][]auction.Auction) []string {
	arbitrages := map[string]int64{}

	for itemId, itemAuctions := range auctions {
		item, ok := wowAPI.LookupItem(itemId, 0)
		if !ok {
			continue
		}
		for _, auc := range itemAuctions {
			if auc.Buyout <= 0 {
				continue
			}
			if auc.Buyout >= item.SellPriceRealizable() {
				continue
			}
			arbitrages[item.Name()] += (item.SellPriceRealizable() - auc.Buyout) * auc.Quantity
		}
	}

	bargains := []string{}
	for name, profit := range arbitrages {
		if profit < common.Coins(1, 0, 0) {
			// Too small to bother with, would just clutter the output
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

	// Generally useful items to keep a watch on
	goods := map[int64]int64{
		65891: common.Coins(30000, 0, 0), // Vial of the Sands (2-person flying mount)
		98715: common.Coins(8000, 0, 0),  // Marked Flawless Battle-Stone
		92741: common.Coins(8000, 0, 0),  // Flawless Battle-Stone

		114821: common.Coins(130, 0, 0), // Hexweave Bag (30 slot)

		//194019: common.Coins(110, 0, 0), // Simply Stitched Reagent Bag (32 slot)
		194020: common.Coins(110, 0, 0), // Chronocloth Reagent Bag (36 slot)
		222855: common.Coins(110, 0, 0), // Weavercloth Reagent Bag (36 slot)
		222854: common.Coins(110, 0, 0), // Dawnweave Reagent Bag (38 slot)

		// Cats I need for "Crazy Cat Lady" title
		8491:  common.Coins(10000, 0, 0), // Black Tabby
		72068: common.Coins(10000, 0, 0), // Guardian Cub
	}

	for itemId, maxPrice := range goods {
		item, ok := wowAPI.LookupItem(itemId, 0)
		if !ok {
			continue
		}
		for _, auc := range auctions[itemId] {
			if auc.Buyout <= 0 {
				continue
			}
			if auc.Buyout < maxPrice {
				str := fmt.Sprintf("%s   %s", item.Name(), common.Gold(auc.Buyout))
				bargains = append(bargains, str)
			}
		}
	}

	return bargains
}

// petValue returns the amount I'm willing to pay for a pet of a given level
func petValue(petLevel int64) int64 {
	level1Max := common.Coins(799, 0, 0)
	level25Max := common.Coins(900, 0, 0)
	extraPerLevel := (level25Max - level1Max) / 24
	return level1Max + extraPerLevel*(petLevel-1)
}

// findPetBargains returns a list of pets that sell for more than they are listed
func findPetBargains(auctions map[int64][]auction.Auction) []string {
	bargains := []string{}

	// SpeciesId of pets that do not resell well
	skipPets := map[int64]bool{
		162: true, // Sinister Squashling
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
		petLevel := petAuction.Pet.Level
		if petAuction.Buyout > petValue(petLevel) {
			continue
		}
		bargains = append(bargains, battlePet.Name(petAuction.Pet.SpeciesId))
	}

	return bargains
}

// findPetSpecialty returns a list of pets I am looking for
func findPetSpecialty(auctions map[int64][]auction.Auction) []string {
	bargains := []string{}

	var specialtyPets = map[int64]int64{
		// Needed for "Crazy Cat Lady" title
		42:  common.Coins(5000, 0, 0), // Black Tabby Cat
		242: common.Coins(5000, 0, 0), // Spectral Tiger Cub
		303: common.Coins(5000, 0, 0), // Nightsaber Cub
		311: common.Coins(5000, 0, 0), // Guardian Cub

		// Collecting it earns a cat battle pet
		93039: common.Coins(5000, 0, 0), // Viscidus Globule (not a cat, but gets there...)

		// Pets that make good gifts
		1890: common.Coins(1000, 0, 0), // Corgi Pup
		1929: common.Coins(1000, 0, 0), // Corgnelius
	}

	for _, petAuction := range auctions[battlePet.PetCageItemId] {
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

	cache.Save()

	if len(arbitrages) == 0 && len(bargains) == 0 && len(petBargains) == 0 && len(petNeeded) == 0 && len(petSpecialty) == 0 {
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
	wowAPI.Init(*passPhrase)

	profileAccessToken, ok := wowAPI.ProfileAccessToken()
	if !ok {
		log.Fatal("ERROR: Unable to obtain profile access token.")
	}
	battlePet.Init(profileAccessToken)

	for _, realm := range strings.Split(*realms, ",") {
		scanRealm(realm)
	}

	generateLua()
}
