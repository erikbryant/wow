package recipes

import (
	"fmt"
	"log"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/common"
	"github.com/erikbryant/wow/internal/wowapi"
)

type Recipe struct {
	href   string
	name   string
	itemID int64
	id     int64
}

var (
	AllRecipes = map[int64]Recipe{}
)

const (
	recipesNeededPath = "./reports/recipesNeeded"
)

func makeRecipe(r any) Recipe {
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
	result, ok := wowapi.Professions(realm, alt)
	if !ok {
		fmt.Println("No professions!", realm, alt, result)
	}

	s, _ := web.MsiValued(result, []string{"secondaries"}, nil)
	for _, prof := range s.([]any) {
		name, _ := web.MsiValued(prof, []string{"profession", "name"}, nil)
		if name != "Cooking" {
			continue
		}
		tiers, _ := web.MsiValued(prof, []string{"tiers"}, nil)
		for _, tier := range tiers.([]any) {
			t, _ := web.MsiValued(tier, []string{"tier", "name"}, nil)
			if t != tierName {
				continue
			}
			kr, _ := web.MsiValued(tier, []string{"known_recipes"}, nil)
			recipes := map[int64]Recipe{}
			for _, k := range kr.([]any) {
				recipe := makeRecipe(k)
				if recipe.name == "Captain Rumsey's Lager" {
					// This is a quest reward or something; won't be found in the AH
					continue
				}
				recipes[recipe.id] = recipe
			}
			return recipes
		}
	}

	return nil
}

func key(alt common.Alt) string {
	return alt.Realm + "-" + alt.Name
}

func scanAlts() map[string]map[int64]Recipe {
	recipesByAlt := map[string]map[int64]Recipe{}

	// Find known recipes for each alt
	for _, alt := range common.Alts {
		kr := knownRecipes(alt.Realm, alt.Name, "Classic Cooking")
		maps.Copy(AllRecipes, kr)
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

	// Ensure log file is empty
	f, err := os.Create(recipesNeededPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Identify missing recipes
	for alt, recipes := range recipesByAlt {
		for _, recipe := range AllRecipes {
			_, ok := recipes[recipe.id]
			if !ok {
				_, err = f.WriteString(alt + " " + recipe.name + "\n")
				if err != nil {
					log.Fatal("Failed to write recipe needed:", recipesNeededPath, err)
				}
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

	_, err = f.WriteString("\nRecipes needed:\n" + strings.Join(rnc, "\n"))
	if err != nil {
		log.Fatal("Failed to write table:", recipesNeededPath, err)
	}

	return rn
}
