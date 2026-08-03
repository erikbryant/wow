package cooking

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/erikbryant/web"
	"github.com/erikbryant/wow/internal/userconfig"
	"github.com/erikbryant/wow/internal/wowapi"
)

type Recipe struct {
	href   string
	name   string
	itemID int64
	id     int64
}

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

func key(alt userconfig.Alt) string {
	return alt.Realm + "-" + alt.Name
}

func scanAlts() (map[int64]Recipe, map[string]map[int64]Recipe) {
	allRecipes := map[int64]Recipe{}
	recipesByAlt := map[string]map[int64]Recipe{}

	// Find known recipes for each alt
	for _, alt := range userconfig.Alts {
		kr := knownRecipes(alt.Realm, alt.Name, "Classic Cooking")
		maps.Copy(allRecipes, kr)
		recipesByAlt[key(alt)] = kr
	}

	// Merge all known recipes into one list
	for _, recipes := range recipesByAlt {
		for _, recipe := range recipes {
			allRecipes[recipe.id] = recipe
		}
	}

	return allRecipes, recipesByAlt
}

func getRecipesNeeded() (string, string, []string) {
	allRecipes, recipesByAlt := scanAlts()
	recipesNeeded := map[string]int{}
	var recipesNeededByAlt strings.Builder

	// Enumerate missing recipes
	for alt, recipes := range recipesByAlt {
		for _, recipe := range allRecipes {
			_, ok := recipes[recipe.id]
			if !ok {
				recipesNeeded[recipe.name]++
				recipesNeededByAlt.WriteString(fmt.Sprintf("%s %s\n", alt, recipe.name))
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

	return recipesNeededByAlt.String(), strings.Join(rnc, "\n"), rn
}

func logRecipes(recipesNeededByAlt string, recipesNeededCount string) error {
	// Ensure log file is empty
	f, err := os.Create(recipesNeededPath)
	if err != nil {
		return fmt.Errorf("could not create recipes needed file: %v", err)
	}
	defer f.Close()

	_, err = f.WriteString("Cooking recipes needed by alt:\n" + recipesNeededByAlt)
	if err != nil {
		return fmt.Errorf("failed to write cooking recipes needed: %s", err)
	}

	_, err = f.WriteString("\nCooking recipes needed by count:\n" + recipesNeededCount + "\n")
	if err != nil {
		return fmt.Errorf("failed to write cooking recipes count: %s", err)
	}

	return nil
}

// RecipesNeeded returns which recipes are needed and writes to the log file
func RecipesNeeded() []string {
	recipesNeededByAlt, recipesNeededCount, recipesNeeded := getRecipesNeeded()

	err := logRecipes(recipesNeededByAlt, recipesNeededCount)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	return recipesNeeded
}
