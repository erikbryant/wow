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
	realms     = flag.String("realms", "Aegwynn,Agamaggan,Akama,Altar of Storms,Andorhal,Anub'arak,Argent Dawn,Azgalor,Azuremyst,Baelgun,Blackhand,Blackwing Lair,Bloodhoof,Bronzebeard,Cairne,Deathwing,Drak'thul,Eitrigg,Farstriders,Ghostlands,Greymane,IceCrown,Kul Tiras,Llane", "WoW realms")
	realmsUS   = flag.Bool("realmsUS", false, "Scan all other US realms")

	// restOfUS is the rest of the realms in the US
	restOfUS = []string{
		"Commodities",
		"Aggramar",
		"Alexstrasza",
		"Alleria",
		"Alterac Mountains",
		"Azjol-Nerub",
		"Bloodscalp",
		"Coilfang",
		"Darrowmere",
		"Dentarg",
		"Draenor",
		"Dragonblight",
		"Durotan",
		"Elune",
		"Eredar",
		"Feathermoon",
		"Frostwolf",
		"Kilrogg",
		"Kirin Tor",
		"Lightninghoof",
		"Misha",
		"Nazgrel",
		"Ravencrest",
		"Runetotem",
		"Sisters of Elune",
	}

	// Generally useful items to keep a watch on
	usefulGoods = map[int64]int64{
		65891: common.Coins(30000, 0, 0), // Vial of the Sands (2-person flying mount)
		92665: common.Coins(3000, 0, 0),  // Flawless Elemental Battle-Stone
		92675: common.Coins(3000, 0, 0),  // Flawless Beast Battle-Stone
		92676: common.Coins(3000, 0, 0),  // Flawless Critter Battle-Stone
		92677: common.Coins(3000, 0, 0),  // Flawless Flying Battle-Stone
		92678: common.Coins(3000, 0, 0),  // Flawless Magic Battle-Stone
		92679: common.Coins(3000, 0, 0),  // Flawless Aquatic Battle-Stone
		92680: common.Coins(3000, 0, 0),  // Flawless Mechanical Battle-Stone
		92681: common.Coins(3000, 0, 0),  // Flawless Undead Battle-Stone
		92682: common.Coins(3000, 0, 0),  // Flawless Humanoid Battle-Stone
		92683: common.Coins(3000, 0, 0),  // Flawless Dragonkin Battle-Stone
		98715: common.Coins(3000, 0, 0),  // Marked Flawless Battle-Stone
		92741: common.Coins(3000, 0, 0),  // Flawless Battle-Stone
	}
)

// findBargains returns auctions for which the goods are below our desired prices
func findBargains(goods map[int64]int64, auctions map[int64][]auction.Auction) []string {
	bargains := []string{}

	for itemId, maxPrice := range goods {
		item, ok := wowAPI.LookupItem(itemId, 0)
		if !ok {
			continue
		}
		if item.Equippable() {
			// Don't know how to price these
			continue
		}
		for _, auc := range auctions[itemId] {
			if auc.Buyout <= 0 {
				continue
			}
			if auc.Buyout < maxPrice {
				c := color.New(color.FgRed)
				str := c.Sprintf("%s %s", item.Name(), common.Gold(auc.Buyout))
				bargains = append(bargains, str)
			}
		}
	}

	return bargains
}

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
		str := fmt.Sprintf("%s   %s", name, common.Gold(profit))
		bargains = append(bargains, str)
	}

	return bargains
}

// printShoppingList prints a list of auctions the user should buy
func printShoppingList(label string, names []string) {
	if len(names) > 0 {
		fmt.Printf("--- %s ---\n", label)
		fmt.Println(strings.Join(common.SortUnique(names), "\n"))
		fmt.Println()
	}
}

// petValue returns the amount I'm willing to pay for a pet of a given level
func petValue(petLevel int64) int64 {
	level1Max := common.Coins(799, 0, 0)
	level25Max := common.Coins(900, 0, 0)
	extraPerLevel := (level25Max - level1Max) / 24
	return level1Max + extraPerLevel*(petLevel-1)
}

// printPetBargains prints a list of pets the user should buy
func printPetBargains(auctions map[int64][]auction.Auction) {
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

	// SpeciesId of pets that do not resell well
	skipPets := map[int64]bool{
		153:  true, // Wolpertinger
		162:  true, // Sinister Squashling
		211:  true, // Strand Crawler
		1687: true, // Left Shark
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
		petLevel := petAuction.Pet.Level
		if petLevel < 25 {
			continue
		}
		if petAuction.Buyout > common.Coins(100, 0, 0) {
			continue
		}
		note := fmt.Sprintf("%s (resell %d)", battlePet.Name(petAuction.Pet.SpeciesId), petAuction.Pet.SpeciesId)
		bargains = append(bargains, note)
	}

	if len(bargains) > 0 {
		fmt.Println("--- Pet auction bargains ---")
		c := color.New(color.FgGreen)
		c.Println(strings.Join(common.SortUnique(bargains), "\n"))
		fmt.Println()
	}
}

// printBargains prints the bargains found in the auction house
func printBargains(auctions map[int64][]auction.Auction) {
	toBuy := findBargains(usefulGoods, auctions)
	printShoppingList("Bargains", toBuy)
	toBuy = findArbitrages(auctions)
	printShoppingList("Arbitrages", toBuy)
}

// scanRealm downloads the available auctions and prints any bargains/arbitrages
func scanRealm(realm string) {
	auctions, ok := auction.GetAuctions(realm)
	if !ok {
		return
	}
	c := color.New(color.FgCyan)
	c.Printf("===========>  %s (%d unique items)  <===========\n\n", realm, len(auctions))
	printBargains(auctions)
	printPetBargains(auctions)
	cache.Save()
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

	var realmsToScan []string
	if *realmsUS {
		realmsToScan = restOfUS
	} else {
		realmsToScan = strings.Split(*realms, ",")
	}
	for _, realm := range realmsToScan {
		scanRealm(realm)
		fmt.Println()
	}

	generateLua()
}
