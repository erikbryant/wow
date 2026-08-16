package toy

import (
	"testing"

	"github.com/erikbryant/wow/internal/wowitem"
)

func toyItem(name string) wowitem.Item {
	return wowitem.Item{XID: 1, XItem: map[string]any{"name": name}}
}
func TestOwned(test *testing.T) {
	t := &Toy{names: map[string]int64{"Toy A": 100, "Toy B": 200}, owned: map[int64]bool{100: true}}
	if !t.Owned(toyItem("Toy A")) {
		test.Error("owned toy reported false")
	}
	if t.Owned(toyItem("Toy B")) || t.Owned(toyItem("Not a toy")) {
		test.Error("unowned/non-toy reported true")
	}
}
