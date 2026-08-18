package battlepet

import (
	"strings"
	"testing"

	"github.com/erikbryant/wow/internal/wowitem"
)

func bpItem(name, subclass string) wowitem.Item {
	return wowitem.Item{XID: 1, XItem: map[string]any{"name": name, "item_subclass": map[string]any{"name": subclass}}}
}
func TestBattlePetMethods(t *testing.T) {
	bp := &BattlePet{names: map[int64]string{10: "Cat", 20: "Dog"}, owned: map[int64]int64{10: 2}}
	if n, ok := bp.PetSpell(bpItem("Cat", "Companion Pets")); !ok || n != 10 {
		t.Fail()
	}
	if _, ok := bp.PetSpell(bpItem("Cat", "Other")); ok {
		t.Fail()
	}
	if _, ok := bp.PetSpell(bpItem("Bird", "Companion Pets")); ok {
		t.Fail()
	}
	if bp.Name(10) != "Cat" || bp.Name(99) != "" || !bp.Owned(10) || bp.Owned(20) {
		t.Fail()
	}
	if bp.LenNames() != 2 || bp.LenOwned() != 1 {
		t.Fail()
	}
	out := bp.Output()
	if !strings.Contains(out, "10: {}, // Cat") || !strings.Contains(out, "20: {}, // Dog") {
		t.Fatalf("%s", out)
	}
}
