package cooking

import (
	"fmt"
	"maps"
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

type CookingRecipe struct {
	recipesThatWeNeed []string
	recipeOutputLog   string
}

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

func knownRecipes(realm, alt, tierName string) (map[int64]Recipe, error) {
	result, err := wowapi.Professions(realm, alt)
	if err != nil {
		return nil, fmt.Errorf("no professions found for %s, %s, %v", realm, alt, result)
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
			return recipes, nil
		}
	}

	return nil, fmt.Errorf("should not have gotten here")
}

func key(alt userconfig.Alt) string {
	return alt.Realm + "-" + alt.Name
}

func scanAlts() (map[int64]Recipe, map[string]map[int64]Recipe, error) {
	allRecipes := map[int64]Recipe{}
	recipesByAlt := map[string]map[int64]Recipe{}

	// Find known recipes for each alt
	for _, alt := range userconfig.Alts {
		kr, err := knownRecipes(alt.Realm, alt.Name, "Classic Cooking")
		if err != nil {
			return nil, nil, err
		}
		maps.Copy(allRecipes, kr)
		recipesByAlt[key(alt)] = kr
	}

	return allRecipes, recipesByAlt, nil
}

func logRecipes(recipesNeededByAlt map[string][]string, recipesNeededCount map[string]int) string {
	var recipeOutputLog strings.Builder

	recipeOutputLog.WriteString("Cooking recipes needed by alt:\n")
	alts := slices.Collect(maps.Keys(recipesNeededByAlt))
	slices.Sort(alts)
	for _, alt := range alts {
		for _, recipe := range recipesNeededByAlt[alt] {
			recipeOutputLog.WriteString(fmt.Sprintf("%-40s %s\n", alt, recipe))
		}
	}

	recipeOutputLog.WriteString("\nCooking recipes needed by count:\n")
	recipes := slices.Collect(maps.Keys(recipesNeededCount))
	for _, recipe := range recipes {
		recipeOutputLog.WriteString(fmt.Sprintf("%-40s  %2d\n", recipe, recipesNeededCount[recipe]))

	}

	return recipeOutputLog.String()
}

func New() (*CookingRecipe, error) {
	allRecipes, recipesByAlt, err := scanAlts()
	if err != nil {
		return nil, err
	}
	recipesNeededCount := map[string]int{}
	recipesNeededByAlt := map[string][]string{}

	// For each alt...
	for alt, altRecipes := range recipesByAlt {
		// For all known recipes, is this alt missing that recipe?
		for _, recipe := range allRecipes {
			_, ok := altRecipes[recipe.id]
			if !ok {
				recipeName := "Recipe: " + recipe.name
				recipesNeededCount[recipeName]++
				recipesNeededByAlt[alt] = append(recipesNeededByAlt[alt], recipeName)
			}
		}
	}

	cr := CookingRecipe{
		recipesThatWeNeed: slices.Collect(maps.Keys(recipesNeededCount)),
		recipeOutputLog:   logRecipes(recipesNeededByAlt, recipesNeededCount),
	}

	slices.Sort(cr.recipesThatWeNeed)

	return &cr, nil
}

func (cr *CookingRecipe) RecipesNeeded() []string {
	return cr.recipesThatWeNeed
}

func (cr *CookingRecipe) Output() string {
	return cr.recipeOutputLog
}
