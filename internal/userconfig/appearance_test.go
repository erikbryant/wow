package userconfig

import "testing"

func TestAppearancesNeed(t *testing.T) {
	a := &Appearances{owned: map[int64]bool{10: true, 20: false}}
	if a.Need([]int64{10}) {
		t.Error("owned appearance should not be needed")
	}
	if !a.Need([]int64{10, 20}) {
		t.Error("unowned appearance should be needed")
	}
	if a.Need([]int64{573}) {
		t.Error("excluded appearance should not be needed")
	}
	if a.Need(nil) {
		t.Error("empty input should not be needed")
	}
	if a.Len() != 2 {
		t.Fail()
	}
}
