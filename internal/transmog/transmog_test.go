package transmog

import (
	"testing"
)

func TestNeedAppearance(t *testing.T) {
	oldFlaky := flakyAppearanceID
	defer func() {
		flakyAppearanceID = oldFlaky
	}()

	flakyAppearanceID = func(int64) bool {
		return false
	}

	tests := []struct {
		name       string
		owned      map[int64]bool
		appearance []int64
		wantNeed   bool
	}{
		{
			name:       "empty",
			owned:      map[int64]bool{},
			appearance: nil,
			wantNeed:   false,
		},
		{
			name:       "owned",
			owned:      map[int64]bool{1: true},
			appearance: []int64{1},
			wantNeed:   false,
		},
		{
			name:       "missing",
			owned:      map[int64]bool{},
			appearance: []int64{1},
			wantNeed:   true,
		},
		{
			name:       "all owned",
			owned:      map[int64]bool{1: true, 2: true, 3: true},
			appearance: []int64{1, 2, 3},
			wantNeed:   false,
		},
		{
			name:       "one missing",
			owned:      map[int64]bool{1: true, 3: true},
			appearance: []int64{1, 2, 3},
			wantNeed:   true,
		},
		{
			name:       "duplicates",
			owned:      map[int64]bool{1: true},
			appearance: []int64{1, 1, 1},
			wantNeed:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			appearanceIDsOwned = tc.owned

			got := NeedAppearance(tc.appearance)
			if got != tc.wantNeed {
				t.Fatalf("NeedAppearance(%v) = %v, want %v",
					tc.appearance, got, tc.wantNeed)
			}
		})
	}
}

func TestNeedAppearanceFlaky(t *testing.T) {
	oldFlaky := flakyAppearanceID
	defer func() {
		flakyAppearanceID = oldFlaky
	}()

	flakyAppearanceID = func(id int64) bool {
		return id == 42
	}

	appearanceIDsOwned = map[int64]bool{}

	if NeedAppearance([]int64{42}) {
		t.Fatal("flaky appearance should not be needed")
	}

	if !NeedAppearance([]int64{43}) {
		t.Fatal("non-flaky missing appearance should be needed")
	}

	if NeedAppearance([]int64{42, 42}) {
		t.Fatal("duplicate flaky appearances should not be needed")
	}
}

func TestGetAppearanceIDsOwned(t *testing.T) {
	old := collectionsTransmogs
	defer func() {
		collectionsTransmogs = old
	}()

	collectionsTransmogs = func() (any, bool) {
		return map[string]any{
			"slots": []any{
				map[string]any{
					"appearances": []any{
						map[string]any{"id": int64(10)},
						map[string]any{"id": int64(20)},
					},
				},
				map[string]any{
					"appearances": []any{
						map[string]any{"id": int64(30)},
						map[string]any{"id": int64(20)},
					},
				},
			},
		}, true
	}

	got, err := getAppearanceIDsOwned()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := map[int64]bool{
		10: true,
		20: true,
		30: true,
	}

	if len(got) != len(want) {
		t.Fatalf("got %d IDs, want %d", len(got), len(want))
	}

	for id := range want {
		if !got[id] {
			t.Errorf("missing appearance ID %d", id)
		}
	}
}

func TestGetAppearanceIDsOwnedError(t *testing.T) {
	old := collectionsTransmogs
	defer func() {
		collectionsTransmogs = old
	}()

	collectionsTransmogs = func() (any, bool) {
		return nil, false
	}

	_, err := getAppearanceIDsOwned()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetAppearanceIDsOwnedEmpty(t *testing.T) {
	old := collectionsTransmogs
	defer func() {
		collectionsTransmogs = old
	}()

	collectionsTransmogs = func() (any, bool) {
		return map[string]any{
			"slots": []any{},
		}, true
	}

	got, err := getAppearanceIDsOwned()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("expected empty map, got %d entries", len(got))
	}
}
