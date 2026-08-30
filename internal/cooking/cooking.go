package cooking

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/erikbryant/wow/internal/common"
	"github.com/erikbryant/wow/internal/userconfig"
	"github.com/erikbryant/wow/internal/wowapi"
)

type Recipe struct {
	href   string
	name   string
	itemID int64
	id     int64
}

type CookingRecipes struct {
	needed      []string
	neededByAlt map[string][]string
	neededCount map[string]int
}

func makeRecipe(r any) Recipe {
	recipe := Recipe{}

	href, _ := common.MsaValued(r, []string{"key", "href"}, nil)
	recipe.href = href.(string)
	name, _ := common.MsaValued(r, []string{"name"}, nil)
	recipe.name = name.(string)
	id, _ := common.MsaValued(r, []string{"id"}, nil)
	recipe.id = common.JSONInt64Panic(id)

	return recipe
}

func getProfession(realm, alt, professionName string) (any, error) {
	// Get all professions
	result, err := wowapi.Professions(realm, alt)
	if err != nil {
		return nil, fmt.Errorf("no professions found for %s, %s, %v", realm, alt, result)
	}

	// Find the desired profession
	s, _ := common.MsaValued(result, []string{"secondaries"}, nil)
	for _, prof := range s.([]any) {
		name, _ := common.MsaValued(prof, []string{"profession", "name"}, nil)
		if name == professionName {
			return prof, nil
		}
	}

	var zero any
	return zero, fmt.Errorf("profession not found %s, %s, %s", realm, alt, professionName)
}

// getTier returns the desired tier (Classic, Outland, etc.)
func getTier(prof any, tierName string) (any, error) {
	tiers, _ := common.MsaValued(prof, []string{"tiers"}, nil)
	for _, tier := range tiers.([]any) {
		t, _ := common.MsaValued(tier, []string{"tier", "name"}, nil)
		if t == tierName {
			return tier, nil
		}
	}

	var zero any
	return zero, fmt.Errorf("tier not found: %s", tierName)
}

func knownClassicCookingRecipes(realm, alt string) (map[int64]Recipe, error) {
	prof, err := getProfession(realm, alt, "Cooking")
	if err != nil {
		return nil, err
	}

	tier, err := getTier(prof, "Classic Cooking")
	if err != nil {
		return nil, err
	}

	kr, _ := common.MsaValued(tier, []string{"known_recipes"}, nil)
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

func key(alt userconfig.Alt) string {
	return alt.Realm + "-" + alt.Name
}

func scanAlts() (map[int64]Recipe, map[string]map[int64]Recipe, error) {
	allRecipes := map[int64]Recipe{}
	recipesByAlt := map[string]map[int64]Recipe{}

	// Find known recipes for each alt
	for _, alt := range userconfig.Alts {
		kr, err := knownClassicCookingRecipes(alt.Realm, alt.Name)
		if err != nil {
			return nil, nil, err
		}
		maps.Copy(allRecipes, kr)
		recipesByAlt[key(alt)] = kr
	}

	return allRecipes, recipesByAlt, nil
}

func generateReport(recipesNeededByAlt map[string][]string, recipesNeededCount map[string]int) string {
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
	slices.Sort(recipes)
	for _, recipe := range recipes {
		recipeOutputLog.WriteString(fmt.Sprintf("%-40s  %2d\n", recipe, recipesNeededCount[recipe]))
	}

	return recipeOutputLog.String()
}

func New() (*CookingRecipes, error) {
	c := CookingRecipes{
		neededCount: map[string]int{},
		neededByAlt: map[string][]string{},
	}

	allRecipes, recipesByAlt, err := scanAlts()
	if err != nil {
		return nil, err
	}

	// For each alt...
	for alt, altRecipes := range recipesByAlt {
		// For all known recipes, is this alt missing that recipe?
		for _, recipe := range allRecipes {
			_, ok := altRecipes[recipe.id]
			if !ok {
				recipeName := "Recipe: " + recipe.name
				c.neededCount[recipeName]++
				c.neededByAlt[alt] = append(c.neededByAlt[alt], recipeName)
			}
		}
	}

	c.needed = slices.Collect(maps.Keys(c.neededCount))
	slices.Sort(c.needed)

	return &c, nil
}

func (cr *CookingRecipes) RecipesNeeded() []string {
	return cr.needed
}

func (cr *CookingRecipes) Output() string {
	return generateReport(cr.neededByAlt, cr.neededCount)
}
