package auction

import (
	"testing"

	"github.com/erikbryant/wow/internal/battlepet"
)

func auction(id, item, buyout, qty int64) map[string]any {
	return map[string]any{"id": float64(id), "item": map[string]any{"id": float64(item)}, "buyout": float64(buyout), "quantity": float64(qty)}
}

func TestNewAuction(t *testing.T) {
	a := auction(11, 22, 333, 4)
	got := newAuction(a)
	if got.ID != 11 || got.ItemID != 22 || got.Buyout != 333 || got.Quantity != 4 {
		t.Fatalf("%+v", got)
	}
}

func TestNewAuctionCommodity(t *testing.T) {
	a := map[string]any{"id": float64(1), "item": map[string]any{"id": float64(2)}, "unit_price": float64(99), "quantity": float64(7)}
	got := newAuction(a)
	if got.Buyout != 99 || got.Quantity != 7 {
		t.Fatalf("%+v", got)
	}
}

func TestNewAuctionPet(t *testing.T) {
	a := auction(1, battlepet.PetCageItemID, 1000, 1)
	a["item"].(map[string]any)["pet_level"] = float64(25)
	a["item"].(map[string]any)["pet_quality_id"] = float64(3)
	a["item"].(map[string]any)["pet_species_id"] = float64(1446)
	got := newAuction(a)
	if got.Pet.Level != 25 || got.Pet.QualityID != 3 || got.Pet.SpeciesID != 1446 {
		t.Fatalf("%+v", got.Pet)
	}
}

func TestBuyoutMissing(t *testing.T) {
	for _, data := range []map[string]any{
		{"id": float64(1), "item": map[string]any{"id": float64(2)}, "quantity": float64(1)},
		{"id": float64(1), "item": map[string]any{"id": float64(2)}, "quantity": float64(1), "buyout": nil},
		{"id": float64(1), "item": map[string]any{"id": float64(2)}, "quantity": float64(1), "buyout": float64(0), "unit_price": float64(17)},
	} {
		if got := newAuction(data).Buyout; got != 0 && got != 17 {
			t.Errorf("buyout=%d", got)
		}
	}
}

func TestBin(t *testing.T) {
	valid := auction(1, 100, 10, 1)
	badPrice := auction(2, 100, 0, 1)
	badID := auction(3, 23704, 10, 1)
	got := bin([]any{valid, badPrice, badID})
	if len(got) != 1 || len(got[100]) != 1 || got[100][0].ID != 1 {
		t.Fatalf("%+v", got)
	}
}
