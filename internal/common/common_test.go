package common

import (
	"testing"
)

func TestCoppers(t *testing.T) {
	tests := []struct {
		name   string
		gold   int64
		silver int64
		copper int64
		want   int64
	}{
		{"zero", 0, 0, 0, 0},
		{"copper", 0, 0, 5, 5},
		{"silver", 0, 5, 0, 500},
		{"gold", 5, 0, 0, 50000},
		{"mixed", 12, 34, 56, 123456},
		{"large", 9999, 99, 99, 99999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Coppers(tt.gold, tt.silver, tt.copper)
			if got != tt.want {
				t.Fatalf("Coppers(%d,%d,%d)=%d want %d",
					tt.gold, tt.silver, tt.copper,
					got, tt.want)
			}
		})
	}
}

func TestGold(t *testing.T) {
	tests := []struct {
		coppers int64
		want    string
	}{
		{0, "0.00.00"},
		{1, "0.00.01"},
		{99, "0.00.99"},
		{100, "0.01.00"},
		{101, "0.01.01"},
		{123456, "12.34.56"},
		{99999999, "9999.99.99"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := Gold(tt.coppers)
			if got != tt.want {
				t.Fatalf("Gold(%d)=%q want %q",
					tt.coppers, got, tt.want)
			}
		})
	}
}

func TestGoldRoundTrip(t *testing.T) {
	tests := []struct {
		g int64
		s int64
		c int64
	}{
		{0, 0, 0},
		{1, 2, 3},
		{12, 34, 56},
		{999, 99, 99},
	}

	for _, tt := range tests {
		value := Coppers(tt.g, tt.s, tt.c)

		if got := Gold(value); got != Gold(Coppers(tt.g, tt.s, tt.c)) {
			t.Fatalf("round-trip failed")
		}
	}
}

func TestQualityID(t *testing.T) {
	tests := []struct {
		name string
		want int64
	}{
		{"Poor", 0},
		{"Common", 1},
		{"Uncommon", 2},
		{"Rare", 3},
		{"Epic", 4},
		{"Legendary", 5},
		{"Artifact", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := QualityID(tt.name); got != tt.want {
				t.Fatalf("QualityID(%q)=%d want %d",
					tt.name, got, tt.want)
			}
		})
	}
}

func TestQualityIDCaseInsensitive(t *testing.T) {
	tests := []string{
		"poor",
		"POOR",
		"pOoR",
		"epic",
		"EPIC",
		"LeGeNdArY",
	}

	for _, quality := range tests {
		if QualityID(quality) < 0 {
			t.Fatalf("%q should be recognized", quality)
		}
	}
}

func TestQualityIDUnknown(t *testing.T) {
	if got := QualityID("Banana"); got != -1 {
		t.Fatalf("QualityID returned %d, want -1", got)
	}
}
