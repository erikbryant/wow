package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/erikbryant/wow/internal/battlepet"
	"github.com/erikbryant/wow/internal/itemcache"
	"github.com/erikbryant/wow/internal/shopping"
	"github.com/erikbryant/wow/internal/toy"
	"github.com/erikbryant/wow/internal/transmog"
	"github.com/erikbryant/wow/internal/wowapi"
)

var (
	passPhrase     = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	realms         = flag.String("realms", "Aegwynn,Agamaggan,Aggramar,Akama,Alexstrasza,Alleria,Altar of Storms,Alterac Mountains,Andorhal,Anub'arak,Argent Dawn,Azgalor,Azjol-Nerub,Azralon,Azuremyst,Baelgun,Barthilas,Blackhand,Blackwing Lair,Bloodhoof,Bloodscalp,Bronzebeard,Caelestrasz,Cairne,Coilfang,Darrowmere,Dath'Remar,Deathwing,Dentarg,Draenor,Dragonblight,Drak'thul,Drakkari,Durotan,Eitrigg,Elune,Eredar,Farstriders,Feathermoon,Frostwolf,Gallywix,Ghostlands,Goldrinn,Greymane,Gundrak,Icecrown,Kilrogg,Kirin Tor,Kul Tiras,Lightninghoof,Llane,Misha,Nazgrel,Nemesis,Quel'Thalas,Ragnaros,Ravencrest,Runetotem,Sisters of Elune,Commodities", "WoW realm(s) to scan")
	oauthAvailable = flag.Bool("oauth", true, "Is OAuth authentication available?")
	summarize      = flag.Bool("summarize", true, "Summarize arbitrages?")
)

const (
	priceCachePath = "./generated/PriceCache.lua"
)

// usage prints a usage message and terminates the program with an error
func usage() {
	log.Fatal(`Usage:
  wow -passPhrase <phrase> [-realms=<realm1,realm2,...>] [-oauthAvailable=true|false] [-summarize=true|false]
`)
}

func main() {
	flag.Parse()

	if *passPhrase == "" {
		fmt.Println("You must specify -passPhrase to unlock the client Id/secret")
		usage()
	}

	wowapi.Init(*passPhrase, *oauthAvailable)
	battlepet.Init(*oauthAvailable)
	toy.Init(*oauthAvailable)
	transmog.Init(*oauthAvailable)
	shopping.Init(*oauthAvailable)

	if !*oauthAvailable {
		fmt.Printf("\n*** OAuth unavailable. Some features may be missing.\n")
	}

	shopping.ScanRealms(*realms, *summarize)

	if *oauthAvailable {
		// Write the prices file for the WoW 'wowMerchant' addon to consume
		err := os.WriteFile(priceCachePath, []byte(itemcache.Lua()), 0600)
		if err != nil {
			log.Fatal(err)
		}
	}
}
