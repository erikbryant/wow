package recipes

import (
	"fmt"
	"slices"
	"strings"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/wowAPI"
)

type Recipe struct {
	href   string
	name   string
	itemID int64
	id     int64
}

var (
	AllRecipes    = map[int64]Recipe{}
	NeededRecipes = map[int64]int64{}
)

func makeRecipe(r interface{}) Recipe {
	recipe := Recipe{}

	href, _ := web.MsiValued(r, []string{"key", "href"}, nil)
	recipe.href = href.(string)
	name, _ := web.MsiValued(r, []string{"name"}, nil)
	recipe.name = name.(string)
	id, _ := web.MsiValued(r, []string{"id"}, nil)
	recipe.id = int64(id.(float64))

	return recipe
}

func knownRecipes(realm, alt, tierName string) map[int64]Recipe {
	result, ok := wowAPI.Professions(realm, alt)
	if !ok {
		fmt.Println("No professions!", realm, alt, result)
	}

	s, _ := web.MsiValued(result, []string{"secondaries"}, nil)
	for _, prof := range s.([]interface{}) {
		name, _ := web.MsiValued(prof, []string{"profession", "name"}, nil)
		if name != "Cooking" {
			continue
		}
		tiers, _ := web.MsiValued(prof, []string{"tiers"}, nil)
		for _, tier := range tiers.([]interface{}) {
			t, _ := web.MsiValued(tier, []string{"tier", "name"}, nil)
			if t != tierName {
				continue
			}
			kr, _ := web.MsiValued(tier, []string{"known_recipes"}, nil)
			recipes := map[int64]Recipe{}
			for _, k := range kr.([]interface{}) {
				recipe := makeRecipe(k)
				recipes[recipe.id] = recipe
			}
			return recipes
		}
	}

	return nil
}

type Alt struct {
	realm string
	name  string
}

var Alts = []Alt{
	//{"Darrowmere", "Rrackette"},

	{"Aegwynn", "Rrynndelleh"},
	{"Agamaggan", "Rresgan"},
	{"Akama", "Rryzella"},
	{"Alexstrasza", "Rrunnestra"},
	{"Alleria", "Rrennie"},
	{"Altar of Storms", "Rrindella"},
	{"Andorhal", "Rrhosthka"},
	{"Anub'arak", "Rrendernna"},
	{"Argent Dawn", "Rreythlyn"},
	{"Azgalor", "Rrhooska"},
	{"Azjol-Nerub", "Rricci"},
	{"Azuremyst", "Rrynochk"},
	{"Baelgun", "Rrhustra"},
	{"Blackhand", "Rrysten"},
	{"Blackwing Lair", "Rrinky"},
	{"Bloodhoof", "Rrishtha"},
	{"Bloodscalp", "Rroushtra"},
	{"Bronzebeard", "Rrimminy"},
	{"Caelestrasz", "Rrousch"},
	{"Cairne", "Rrhazzash"},
	{"Coilfang", "Rrexanna"},
	{"Deathwing", "Rruthenny"},
	{"Dentarg", "Rrhoul"},
	{"Draenor", "Rrooh"},
	{"Dragonblight", "Rrizzy"},
	{"Drak'thul", "Rrkharst"},
	{"Durotan", "Rryjhen"},
	{"Eitrigg", "Rrhyn"},
	{"Elune", "Rrazyn"},
	{"Farstriders", "Rrhooshka"},
	{"Feathermoon", "Rrhannet"},
	{"Frostwolf", "Rrouscha"},
	//{"Ghostlands", "Rreezy"},
	//{"Greymane", "Rrznyth"},
	//{"Icecrown", "Pkhats"},
	//{"Kilrogg", "Rrinky"},
	//{"Kirin Tor", "Rruggles"},
	//{"Kul Tiras", "Rrkulth"},
	//{"Lightninghoof", "Rroovetta"},
	//{"Llane", "Rreebenna"},
	//{"Misha", "Rrazhneth"},
	//{"Nazgrel", "Rryxny"},
	//{"Ravencrest", "Rrothinna"},
	//{"Runetotem", "Rrygellna"},
	//{"Sisters of Elune", "Rrhette"},
}

func key(alt Alt) string {
	return alt.realm + "-" + alt.name
}

func scanAlts() map[string]map[int64]Recipe {
	recipesByAlt := map[string]map[int64]Recipe{}

	// Find known recipes for each alt
	for _, alt := range Alts {
		kr := knownRecipes(alt.realm, alt.name, "Classic Cooking")
		for id, r := range kr {
			AllRecipes[id] = r
		}
		recipesByAlt[key(alt)] = kr
	}

	// Merge all known recipes into one list
	for _, recipes := range recipesByAlt {
		for _, recipe := range recipes {
			AllRecipes[recipe.id] = recipe
		}
	}

	return recipesByAlt
}

func Needed() []string {
	recipesByAlt := scanAlts()
	recipesNeeded := map[string]int{}

	// Identify missing recipes
	for alt, recipes := range recipesByAlt {
		for _, recipe := range AllRecipes {
			_, ok := recipes[recipe.id]
			if !ok {
				fmt.Println(alt, recipe.name)
				recipesNeeded[recipe.name]++
			}
		}
	}

	rn := []string{}
	rnc := []string{}
	for recipe, count := range recipesNeeded {
		rn = append(rn, "Recipe: "+recipe)
		r := fmt.Sprintf("%-30s  %2d", recipe, count)
		rnc = append(rnc, r)
	}
	slices.Sort(rn)
	slices.Sort(rnc)

	fmt.Println()
	fmt.Println("Recipes needed:")
	fmt.Println(strings.Join(rnc, "\n"))
	fmt.Println()

	return rn
}
