package wowitem

import "testing"

func TestILevels(t *testing.T) {
	if !Known(237946) || Known(999999999) {
		t.Fail()
	}
	got := ILevels(237946)
	want := []int64{180, 186, 192, 199, 206}
	if len(got) != len(want) {
		t.Fatalf("%v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%v", got)
		}
	}
	if got := ILevels(999999999); len(got) != 1 || got[0] != 0 {
		t.Fatalf("unknown=%v", got)
	}
}
