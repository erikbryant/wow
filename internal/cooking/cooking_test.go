package cooking

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/erikbryant/wow/internal/userconfig"
)

func TestMakeRecipe(t *testing.T) {
	r := makeRecipe(map[string]any{"key": map[string]any{"href": "/recipe/1"}, "name": "Recipe One", "id": json.Number("42")})
	if r.href != "/recipe/1" || r.name != "Recipe One" || r.id != 42 {
		t.Fatalf("%+v", r)
	}
}
func TestGetTier(t *testing.T) {
	prof := map[string]any{"tiers": []any{map[string]any{"tier": map[string]any{"name": "Classic Cooking"}}, map[string]any{"tier": map[string]any{"name": "Outland Cooking"}}}}
	if _, err := getTier(prof, "Outland Cooking"); err != nil {
		t.Fatal(err)
	}
	if _, err := getTier(prof, "Missing"); err == nil {
		t.Fatal("expected error")
	}
}
func TestKey(t *testing.T) {
	if got := key(userconfig.Alt{Realm: "A", Name: "B"}); got != "A-B" {
		t.Fail()
	}
}
func TestGenerateReport(t *testing.T) {
	s := generateReport(map[string][]string{"Zed-Z": {"Recipe: B", "Recipe: A"}, "A-A": {"Recipe: C"}}, map[string]int{"Recipe: B": 2, "Recipe: A": 1, "Recipe: C": 3})
	if !strings.Contains(s, "Cooking recipes needed by alt:") || strings.Index(s, "A-A") > strings.Index(s, "Zed-Z") {
		t.Error("alts not sorted")
	}
	if !strings.Contains(s, "Recipe: A") || !strings.Contains(s, "Recipe: C") {
		t.Error("recipes missing")
	}
}
