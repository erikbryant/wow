package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/erikbryant/wow/internal/appearanceset"
	"github.com/erikbryant/wow/internal/wowitem"
)

func outputItem() wowitem.Item {
	return wowitem.Item{XID: 123, XUpdated: timeMust(), XItem: map[string]any{"name": "Widget", "level": json.Number("100"), "is_stackable": true, "is_equippable": false, "item_class": map[string]any{"name": "Consumable"}, "item_subclass": map[string]any{"name": "Potion"}, "preview_item": map[string]any{"quality": map[string]any{"name": "Rare"}, "sell_price": map[string]any{"value": json.Number("12345")}}}}
}

func timeMust() (t time.Time) { return time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC) }

func TestColorize(t *testing.T) {
	if got := Colorize("x", FgRed); got != "\x1b[31mx\x1b[0m" {
		t.Fatalf("%q", got)
	}
}

func TestJSON(t *testing.T) {
	var b bytes.Buffer
	JSON(&b, []wowitem.Item{outputItem()})
	if !strings.Contains(b.String(), `"XID": 123`) || !strings.Contains(b.String(), `"name": "Widget"`) {
		t.Fatalf("%s", b.String())
	}
}

func TestTable(t *testing.T) {
	var b bytes.Buffer
	as := appearanceset.NewEmpty(t.TempDir() + "/appearances")
	as.Set(123, true)
	Table(&b, []wowitem.Item{outputItem()}, as)
	s := b.String()
	for _, x := range []string{"ID", "Equips", "Stacks", "App Set", "Sell Price", "iLvl", "Class", "Quality", "Updated", "Name", "123", "Widget", "1.23.45"} {
		if !strings.Contains(s, x) {
			t.Errorf("table missing %q: %s", x, s)
		}
	}
}
