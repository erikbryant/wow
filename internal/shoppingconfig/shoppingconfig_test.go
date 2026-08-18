package shoppingconfig

import (
	"testing"

	"github.com/erikbryant/wow/internal/cooking"
	"github.com/erikbryant/wow/internal/wowitem"
)

func TestNewDefaults(t *testing.T) {
	wi := wowitem.NewEmpty(t.TempDir() + "/items")
	cr := &cooking.CookingRecipes{}
	c := New(wi, cr)
	if c.AppearancePriceMax <= 0 || c.AppearancePriceInSetMax <= c.AppearancePriceMax || c.ProfitToDisplayMin <= 0 {
		t.Fatal("unexpected defaults")
	}
	for _, id := range []int64{1385, 1706, 1150} {
		if _, ok := c.SkipPets[id]; !ok {
			t.Errorf("missing skip pet %d", id)
		}
	}
	if len(c.UsefulGoods) == 0 {
		t.Error("expected useful goods")
	}
}
