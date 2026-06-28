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

var usefulRecipes = map[int64]struct{}{
	// Outland cooking
	//itemCache.Search("Recipe: Blackened Trout").Id():     {}, // 1
	//itemCache.Search("Recipe: Buzzard Bites").Id():       {}, // 1
	//itemCache.Search("Recipe: Clam Bar").Id():            {}, // 1
	//itemCache.Search("Recipe: Blackened Sporefish").Id(): {}, // 10
	//itemCache.Search("Recipe: Blackened Basilisk").Id():  {}, // 15
	//itemCache.Search("Recipe: Grilled Mudfish").Id():     {}, // 20
	//itemCache.Search("Recipe: Poached Bluefish").Id():    {}, // 20
	//itemCache.Search("Recipe: Golden Fish Sticks").Id():  {}, // 25
	//itemCache.Search("Recipe: Roasted Clefthoof").Id():   {}, // 25
	//itemCache.Search("Recipe: Talbuk Steak").Id():        {}, // 25
	//itemCache.Search("Recipe: Warp Burger").Id():         {}, // 25
	//itemCache.Search("Recipe: Spicy Crawdad").Id():       {}, // 50

	// Stormwind Cooking Trainer
	//itemCache.Search("Recipe: Kaldorei Spider Kabob").Id():   {}, // 10
	//itemCache.Search("Recipe: Tasty Lion Steak").Id():        {}, // 150
	//itemCache.Search("Recipe: Barbecued Buzzard Wing").Id():  {}, // 175
	//itemCache.Search("Recipe: Soothing Turtle Bisque").Id():  {}, // 175
	//itemCache.Search("Recipe: Spider Sausage").Id():          {}, // 200
	//itemCache.Search("Recipe: Spotted Yellowtail").Id():      {}, // 225
	//itemCache.Search("Recipe: Grilled Squid").Id():           {}, // 240
	//itemCache.Search("Recipe: Charred Bear Kabobs").Id():     {}, // 250
	//itemCache.Search("Recipe: Juicy Bear Burger").Id():       {}, // 250
	//itemCache.Search("Recipe: Nightfin Soup").Id():           {}, // 250
	//itemCache.Search("Recipe: Poached Sunscale Salmon").Id(): {}, // 250

	// Stormwind Recipe Vendor: Kendor Kabonka
	//itemCache.Search("Recipe: Beer Basted Boar Ribs").Id():  {}, // 10
	//itemCache.Search("Recipe: Goretusk Liver Pie").Id():     {}, // 50
	//itemCache.Search("Recipe: Westfall Stew").Id():          {}, // 50
	//itemCache.Search("Recipe: Blood Sausage").Id():          {}, // 60
	//itemCache.Search("Recipe: Crocolisk Steak").Id():        {}, // 80
	//itemCache.Search("Recipe: Cooked Crab Claw").Id():       {}, // 85
	//itemCache.Search("Recipe: Murloc Fin Soup").Id():        {}, // 90
	//itemCache.Search("Recipe: Redridge Goulash").Id():       {}, // 100
	//itemCache.Search("Recipe: Seasoned Wolf Kabob").Id():    {}, // 100
	//itemCache.Search("Recipe: Gooey Spider Cake").Id():      {}, // 110
	//itemCache.Search("Recipe: Succulent Pork Ribs").Id():    {}, // 110
	//itemCache.Search("Recipe: Crocolisk Gumbo").Id():        {}, // 120
	//itemCache.Search("Recipe: Curiously Tasty Omelet").Id(): {}, // 130

	// Classic cooking
	//itemCache.Search("Recipe: Brilliant Smallfish").Id():          {}, // 1
	//itemCache.Search("Recipe: Crispy Bat Wing").Id():              {}, // 1
	//itemCache.Search("Recipe: Extra Lemony Herb Filet").Id():      {}, // 1
	//itemCache.Search("Recipe: Gingerbread Cookie").Id():           {}, // 1
	//itemCache.Search("Recipe: Lemon Herb Filet").Id():             {}, // 1
	//itemCache.Search("Recipe: Lynx Steak").Id():                   {}, // 1
	//itemCache.Search("Recipe: Roasted Moongraze Tenderloin").Id(): {}, // 1
	//itemCache.Search("Recipe: Slitherskin Mackerel").Id():         {}, // 1
	//itemCache.Search("Recipe: Scorpid Surprise").Id():             {}, // 20
	//itemCache.Search("Recipe: Roasted Kodo Meat").Id():            {}, // 35
	//itemCache.Search("Recipe: Smoked Bear Meat").Id():             {}, // 40
	//itemCache.Search("Recipe: Bat Bites").Id():                    {}, // 50
	//itemCache.Search("Recipe: Loch Frenzy Delight").Id():          {}, // 50
	//itemCache.Search("Recipe: Longjaw Mud Snapper").Id():          {}, // 50
	//itemCache.Search("Recipe: Rainbow Fin Albacore").Id():         {}, // 50
	//itemCache.Search("Recipe: Strider Stew").Id():                 {}, // 50
	//itemCache.Search("Recipe: Crunchy Spider Surprise").Id():      {}, // 60
	//itemCache.Search("Recipe: Thistle Tea").Id():                  {}, // 60
	//itemCache.Search("Recipe: Smoked Sagefish").Id():              {}, // 80
	//itemCache.Search("Recipe: Savory Deviate Delight").Id():       {}, // 85
	//itemCache.Search("Recipe: Clam Chowder").Id():                 {}, // 90
	//itemCache.Search("Recipe: Bristle Whisker Catfish").Id():      {}, // 100
	//itemCache.Search("Recipe: Crispy Lizard Tail").Id():           {}, // 100
	//itemCache.Search("Recipe: Big Bear Steak").Id():               {}, // 110
	//itemCache.Search("Recipe: Lean Venison").Id():                 {}, // 110
	//itemCache.Search("Recipe: Hot Lion Chops").Id():               {}, // 125
	//itemCache.Search("Recipe: Lean Wolf Steak").Id():              {}, // 125
	//itemCache.Search("Recipe: Heavy Crocolisk Stew").Id():         {}, // 150
	//itemCache.Search("Recipe: Goldthorn Tea").Id():                {}, // 160
	//itemCache.Search("Recipe: Carrion Surprise").Id():             {}, // 175
	//itemCache.Search("Recipe: Giant Clam Scorcho").Id():           {}, // 175
	//itemCache.Search("Recipe: Hot Wolf Ribs").Id():                {}, // 175
	//itemCache.Search("Recipe: Jungle Stew").Id():                  {}, // 175
	//itemCache.Search("Recipe: Mithril Head Trout").Id():           {}, // 175
	//itemCache.Search("Recipe: Mystery Stew").Id():                 {}, // 175
	//itemCache.Search("Recipe: Roast Raptor").Id():                 {}, // 175
	//itemCache.Search("Recipe: Rockscale Cod").Id():                {}, // 175
	//itemCache.Search("Recipe: Sagefish Delight").Id():             {}, // 175
	//itemCache.Search("Recipe: Dragonbreath Chili").Id():           {}, // 200
	//itemCache.Search("Recipe: Heavy Kodo Stew").Id():              {}, // 200
	//itemCache.Search("Recipe: Cooked Glossy Mightfish").Id():      {}, // 225
	//itemCache.Search("Recipe: Filet of Redgill").Id():             {}, // 225
	//itemCache.Search("Recipe: Monster Omelet").Id():               {}, // 225
	//itemCache.Search("Recipe: Spiced Chili Crab").Id():            {}, // 225
	//itemCache.Search("Recipe: Tender Wolf Steak").Id():            {}, // 225
	//itemCache.Search("Recipe: Undermine Clam Chowder").Id():       {}, // 225
	//itemCache.Search("Recipe: Hot Smoked Bass").Id():              {}, // 240
	//itemCache.Search("Recipe: Baked Salmon").Id():                 {}, // 275
	//itemCache.Search("Recipe: Lobster Stew").Id():                 {}, // 275
	//itemCache.Search("Recipe: Mightfish Steak").Id():              {}, // 275
}

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
	{"Darrowmere", "Rrackette"},
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
	{"Ghostlands", "Rreezy"},
	{"Greymane", "Rrznyth"},
	{"Icecrown", "Pkhats"},
	{"Kilrogg", "Rrinky"},
	{"Kirin Tor", "Rruggles"},
	{"Kul Tiras", "Rrkulth"},
	{"Lightninghoof", "Rroovetta"},
	{"Llane", "Rreebenna"},
	{"Misha", "Rrazhneth"},
	{"Nazgrel", "Rryxny"},
	{"Ravencrest", "Rrothinna"},
	{"Runetotem", "Rrygellna"},
	{"Sisters of Elune", "Rrhette"},
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
