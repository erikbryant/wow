package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/shopping"
	"github.com/erikbryant/wow/internal/wowapi"
)

var (
	passphrase = flag.String("passphrase", "", "Passphrase to unlock WoW API client ID/secret")
	realms     = flag.String("realms", "Aegwynn,Agamaggan,Aggramar,Akama,Alexstrasza,Alleria,Altar of Storms,Alterac Mountains,Andorhal,Anub'arak,Argent Dawn,Azgalor,Azjol-Nerub,Azralon,Azuremyst,Baelgun,Barthilas,Blackhand,Blackwing Lair,Bloodhoof,Bloodscalp,Bronzebeard,Caelestrasz,Cairne,Coilfang,Darrowmere,Dath'Remar,Deathwing,Dentarg,Draenor,Dragonblight,Drak'thul,Drakkari,Durotan,Eitrigg,Elune,Eredar,Farstriders,Feathermoon,Frostwolf,Gallywix,Ghostlands,Goldrinn,Greymane,Gundrak,Icecrown,Kilrogg,Kirin Tor,Kul Tiras,Lightninghoof,Llane,Misha,Nazgrel,Nemesis,Quel'Thalas,Ragnaros,Ravencrest,Runetotem,Sisters of Elune,Commodities", "WoW realm(s) to scan")
	summarize  = flag.Bool("summarize", true, "Summarize arbitrages?")
)

// usage prints a usage message and terminates the program with an error
func usage() {
	fmt.Println(`Usage:
  wow -passphrase <phrase> [-realms=<realm1,realm2,...>] [-summarize=true|false]`)
}

func main() {
	flag.Parse()

	if *passphrase == "" {
		usage()
		os.Exit(1)
	}

	err := wowapi.Init(*passphrase)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = shopping.Shop(*realms, *summarize)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
