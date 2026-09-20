package regions

import "strings"

type Realm struct {
	Name     string
	Slug     string
	Region   string
	Language string
}

type Realms map[string]Realm

const (
	USenRealms = "Aegwynn,Agamaggan,Aggramar,Akama,Alexstrasza,Alleria,Altar of Storms,Alterac Mountains,Andorhal,Anub'arak,Argent Dawn,Azgalor,Azjol-Nerub,Azralon,Azuremyst,Baelgun,Barthilas,Blackhand,Blackwing Lair,Bloodhoof,Bloodscalp,Bronzebeard,Caelestrasz,Cairne,Coilfang,Darrowmere,Dath'Remar,Deathwing,Dentarg,Draenor,Dragonblight,Drak'thul,Drakkari,Durotan,Eitrigg,Elune,Eredar,Farstriders,Feathermoon,Frostwolf,Gallywix,Ghostlands,Goldrinn,Greymane,Gundrak,Icecrown,Kilrogg,Kirin Tor,Kul Tiras,Lightninghoof,Llane,Misha,Nazgrel,Nemesis,Quel'Thalas,Ragnaros,Ravencrest,Runetotem,Sisters of Elune,Commodities"
	EUruRealms = "Вечная Песня"
)

func New() Realms {
	realms := Realms{}

	for realm := range strings.SplitSeq(USenRealms, ",") {
		realms.Add(realm, "US", "en_US")
	}

	for realm := range strings.SplitSeq(EUruRealms, ",") {
		realms.Add(realm, "EU", "ru_RU")
	}

	return realms
}

// RealmToSlug returns the slug form of a given realm name, based on WoW naming rules.
func RealmToSlug(realm string) string {
	slug := strings.ToLower(realm)
	slug = strings.ReplaceAll(slug, "-", "")
	slug = strings.ReplaceAll(slug, "'", "")
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}

func (r Realms) Add(realm, region, language string) {
	realm = strings.TrimSpace(realm)
	if realm == "" {
		return
	}

	slug := RealmToSlug(realm)
	config := Realm{
		Name:     realm,
		Slug:     slug,
		Region:   region,
		Language: language,
	}
	r[realm] = config
	r[slug] = config
}

func (r Realms) Config(realm string) (string, string) {
	config, ok := r[realm]
	if !ok {
		return "US", "en_US"
	}
	return config.Region, config.Language
}
