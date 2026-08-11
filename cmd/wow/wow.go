package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erikbryant/wow/internal/shopping"
	"github.com/erikbryant/wow/internal/wowapi"
)

var (
	realms = flag.String("realms", "Aegwynn,Agamaggan,Aggramar,Akama,Alexstrasza,Alleria,Altar of Storms,Alterac Mountains,Andorhal,Anub'arak,Argent Dawn,Azgalor,Azjol-Nerub,Azralon,Azuremyst,Baelgun,Barthilas,Blackhand,Blackwing Lair,Bloodhoof,Bloodscalp,Bronzebeard,Caelestrasz,Cairne,Coilfang,Darrowmere,Dath'Remar,Deathwing,Dentarg,Draenor,Dragonblight,Drak'thul,Drakkari,Durotan,Eitrigg,Elune,Eredar,Farstriders,Feathermoon,Frostwolf,Gallywix,Ghostlands,Goldrinn,Greymane,Gundrak,Icecrown,Kilrogg,Kirin Tor,Kul Tiras,Lightninghoof,Llane,Misha,Nazgrel,Nemesis,Quel'Thalas,Ragnaros,Ravencrest,Runetotem,Sisters of Elune,Commodities", "WoW realm(s) to scan")
)

func main() {
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, "Usage: wow [-realms=realm1,realm2,...]")
		os.Exit(1)
	}

	flag.Parse()

	err := wowapi.Init()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	err = shopping.Shop(*realms)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
