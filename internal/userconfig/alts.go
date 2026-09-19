package userconfig

import "strings"

type Realms struct {
	Realms   []string
	Language string
	Region   string
}

var (
	RealmsWithAltsUS = Realms{
		Realms:   strings.Split("Aegwynn,Agamaggan,Aggramar,Akama,Alexstrasza,Alleria,Altar of Storms,Alterac Mountains,Andorhal,Anub'arak,Argent Dawn,Azgalor,Azjol-Nerub,Azralon,Azuremyst,Baelgun,Barthilas,Blackhand,Blackwing Lair,Bloodhoof,Bloodscalp,Bronzebeard,Caelestrasz,Cairne,Coilfang,Darrowmere,Dath'Remar,Deathwing,Dentarg,Draenor,Dragonblight,Drak'thul,Drakkari,Durotan,Eitrigg,Elune,Eredar,Farstriders,Feathermoon,Frostwolf,Gallywix,Ghostlands,Goldrinn,Greymane,Gundrak,Icecrown,Kilrogg,Kirin Tor,Kul Tiras,Lightninghoof,Llane,Misha,Nazgrel,Nemesis,Quel'Thalas,Ragnaros,Ravencrest,Runetotem,Sisters of Elune,Commodities", ","),
		Language: "en_US",
		Region:   "US",
	}

	RealmsWithAltsEU = Realms{
		Realms:   strings.Split("Вечная Песня", ","),
		Language: "ru_RU",
		Region:   "EU",
	}
)
