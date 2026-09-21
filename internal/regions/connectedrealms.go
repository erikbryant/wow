package regions

import (
	"fmt"
)

// connectedRealmIDs maps realm names to their ID.
var connectedRealmIDs = map[string]string{

	// ---------------------------- US region ----------------------------

	// US realms
	"Aegwynn":           "1136",
	"Agamaggan":         "1129",
	"Aggramar":          "106",
	"Akama":             "84",
	"Alexstrasza":       "1070",
	"Alleria":           "52",
	"Altar of Storms":   "78",
	"Alterac Mountains": "71",
	"Andorhal":          "96",
	"Anub'arak":         "1138",
	"Argent Dawn":       "75",
	"Azgalor":           "77",
	"Azjol-Nerub":       "121",
	"Azuremyst":         "160",
	"Baelgun":           "1190",
	"Blackhand":         "54",
	"Blackwing Lair":    "154",
	"Bloodhoof":         "64",
	"Bloodscalp":        "1185",
	"Bronzebeard":       "117",
	"Cairne":            "1168",
	"Coilfang":          "157",
	"Darrowmere":        "113",
	"Deathwing":         "155",
	"Dentarg":           "55",
	"Draenor":           "115",
	"Dragonblight":      "114",
	"Drak'thul":         "86",
	"Durotan":           "63",
	"Eitrigg":           "47",
	"Elune":             "67",
	"Eredar":            "53",
	"Farstriders":       "12",
	"Feathermoon":       "118",
	"Frostwolf":         "127",
	"Ghostlands":        "1175",
	"Greymane":          "158",
	"Icecrown":          "104",
	"Kilrogg":           "4",
	"Kirin Tor":         "1071",
	"Kul Tiras":         "1147",
	"Lightninghoof":     "163",
	"Llane":             "99",
	"Misha":             "1151",
	"Nazgrel":           "1184",
	"Ravencrest":        "1072",
	"Runetotem":         "151",
	"Sisters of Elune":  "125",

	// Remote realms: Oceanic
	"Aman'Thul":   "3726",
	"Barthilas":   "3723",
	"Caelestrasz": "3721",
	"Dath'Remar":  "3726",
	"Dreadmaul":   "3725",
	"Frostmourne": "3725",
	"Gundrak":     "3725",
	"Jubei'Thos":  "3725",
	"Khaz'goroth": "3726",
	"Nagrand":     "3721",
	"Saurfang":    "3721",
	"Thaurissan":  "3725",

	// Remote realms: Brazil
	"Azralon":   "3209",
	"Gallywix":  "3234",
	"Goldrinn":  "3207",
	"Nemesis":   "3208",
	"Tol Barad": "3208",

	// Remote realms: Latin America
	"Drakkari":    "1425",
	"Quel'Thalas": "1428",
	"Ragnaros":    "1427",

	// ---------------------------- EU region ----------------------------

	// Russian server
	"Вечная Песня": "1925",
}

// ConnectedRealmID returns the connected realm ID of the given realm.
func ConnectedRealmID(realm string) (string, error) {
	id, ok := connectedRealmIDs[realm]
	if ok {
		return id, nil
	}
	return "", fmt.Errorf("failed to find realm: %s", realm)
}
