package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/erikbryant/wow/internal/shopping"
	"github.com/erikbryant/wow/internal/wowapi"
	"github.com/erikbryant/wow/internal/wowitem"
)

var (
	passphrase     = flag.String("passphrase", "", "Passphrase to unlock WoW API client ID/secret")
	realms         = flag.String("realms", "Aegwynn,Agamaggan,Aggramar,Akama,Alexstrasza,Alleria,Altar of Storms,Alterac Mountains,Andorhal,Anub'arak,Argent Dawn,Azgalor,Azjol-Nerub,Azralon,Azuremyst,Baelgun,Barthilas,Blackhand,Blackwing Lair,Bloodhoof,Bloodscalp,Bronzebeard,Caelestrasz,Cairne,Coilfang,Darrowmere,Dath'Remar,Deathwing,Dentarg,Draenor,Dragonblight,Drak'thul,Drakkari,Durotan,Eitrigg,Elune,Eredar,Farstriders,Feathermoon,Frostwolf,Gallywix,Ghostlands,Goldrinn,Greymane,Gundrak,Icecrown,Kilrogg,Kirin Tor,Kul Tiras,Lightninghoof,Llane,Misha,Nazgrel,Nemesis,Quel'Thalas,Ragnaros,Ravencrest,Runetotem,Sisters of Elune,Commodities", "WoW realm(s) to scan")
	oauthAvailable = flag.Bool("oauth", true, "Is OAuth authentication available?")
	summarize      = flag.Bool("summarize", true, "Summarize arbitrages?")
)

const (
	priceCachePath = "./exports/PriceCache.lua"
)

// usage prints a usage message and terminates the program with an error
func usage() {
	log.Fatal(`Usage:
  wow -passphrase <phrase> [-realms=<realm1,realm2,...>] [-oauthAvailable=true|false] [-summarize=true|false]
`)
}

func main() {
	flag.Parse()

	if *passphrase == "" {
		fmt.Println("You must specify -passphrase to unlock the client ID/secret")
		usage()
	}

	if !*oauthAvailable {
		fmt.Printf("\n*** OAuth unavailable. Some features may be missing.\n")
	}

	wowapi.Init(*passphrase, *oauthAvailable)

	shopping.Shop(*realms, *summarize, *oauthAvailable)

	if *oauthAvailable {
		// Write the prices file for the WoW 'wowMerchant' addon to consume
		err := os.WriteFile(priceCachePath, []byte(wowitem.Lua()), 0600)
		if err != nil {
			log.Fatal(err)
		}
	}
}
