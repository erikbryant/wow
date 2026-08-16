package auction

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/erikbryant/wow/internal/battlepet"
)

func jsonNumber(n int64) json.Number {
	s := strconv.FormatInt(n, 10)
	return json.Number(s)
}

func auction(id, item, buyout, qty int64) map[string]any {
	return map[string]any{"id": jsonNumber(id), "item": map[string]any{"id": jsonNumber(item)}, "buyout": jsonNumber(buyout), "quantity": jsonNumber(qty)}
}

func TestNewAuction(t *testing.T) {
	a := auction(11, 22, 333, 4)
	got := newAuction(a)
	if got.ID != 11 || got.ItemID != 22 || got.Buyout != 333 || got.Quantity != 4 {
		t.Fatalf("%+v", got)
	}
}

func TestNewAuctionCommodity(t *testing.T) {
	a := map[string]any{"id": jsonNumber(1), "item": map[string]any{"id": jsonNumber(2)}, "unit_price": jsonNumber(99), "quantity": jsonNumber(7)}
	got := newAuction(a)
	if got.Buyout != 99 || got.Quantity != 7 {
		t.Fatalf("%+v", got)
	}
}

func TestNewAuctionPet(t *testing.T) {
	a := auction(1, battlepet.PetCageItemID, 1000, 1)
	a["item"].(map[string]any)["pet_level"] = jsonNumber(25)
	a["item"].(map[string]any)["pet_quality_id"] = jsonNumber(3)
	a["item"].(map[string]any)["pet_species_id"] = jsonNumber(1446)
	got := newAuction(a)
	if got.Pet.Level != 25 || got.Pet.QualityID != 3 || got.Pet.SpeciesID != 1446 {
		t.Fatalf("%+v", got.Pet)
	}
}

func TestBuyoutMissing(t *testing.T) {
	for _, data := range []map[string]any{
		{"id": jsonNumber(1), "item": map[string]any{"id": jsonNumber(2)}, "quantity": jsonNumber(1)},
		{"id": jsonNumber(1), "item": map[string]any{"id": jsonNumber(2)}, "quantity": jsonNumber(1), "buyout": nil},
		{"id": jsonNumber(1), "item": map[string]any{"id": jsonNumber(2)}, "quantity": jsonNumber(1), "buyout": jsonNumber(0), "unit_price": jsonNumber(17)},
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
