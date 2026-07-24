package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/erikbryant/wow/battlepet"
	"github.com/erikbryant/wow/common"
	"github.com/erikbryant/wow/itemcache"
	"github.com/erikbryant/wow/shopping"
	"github.com/erikbryant/wow/toy"
	"github.com/erikbryant/wow/transmog"
	"github.com/erikbryant/wow/wowapi"
)

var (
	passPhrase     = flag.String("passPhrase", "", "Passphrase to unlock WOW API client Id/secret")
	realms         = flag.String("realms", "Aegwynn,Agamaggan,Aggramar,Akama,Alexstrasza,Alleria,Altar of Storms,Alterac Mountains,Andorhal,Anub'arak,Argent Dawn,Azgalor,Azjol-Nerub,Azralon,Azuremyst,Baelgun,Barthilas,Blackhand,Blackwing Lair,Bloodhoof,Bloodscalp,Bronzebeard,Caelestrasz,Cairne,Coilfang,Darrowmere,Dath'Remar,Deathwing,Dentarg,Draenor,Dragonblight,Drak'thul,Drakkari,Durotan,Eitrigg,Elune,Eredar,Farstriders,Feathermoon,Frostwolf,Gallywix,Ghostlands,Goldrinn,Greymane,Gundrak,Icecrown,Kilrogg,Kirin Tor,Kul Tiras,Lightninghoof,Llane,Misha,Nazgrel,Nemesis,Quel'Thalas,Ragnaros,Ravencrest,Runetotem,Sisters of Elune,Commodities", "WoW realm(s) to scan")
	oauthAvailable = flag.Bool("oauth", true, "Is OAuth authentication available?")
	summarize      = flag.Bool("summarize", true, "Summarize arbitrages?")
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
	shopping.Init()

	if !*oauthAvailable {
		fmt.Printf("\n*** OAuth unavailable. Some features may be missing.\n")
	}

	shopping.ScanRealms(*realms, *summarize, *oauthAvailable)

	if *oauthAvailable {
		// Write the prices file for the WoW 'wowMerchant' addon to consume
		common.WriteFile("./generated/PriceCache.lua", itemcache.Lua())
	}
}
