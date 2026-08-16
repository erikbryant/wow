package common

import (
	"encoding/json"
	"math"
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

func TestJSONString(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "string",
			value: "123",
			want:  "123",
		},
		{
			name:  "float",
			value: float64(123),
			want:  "123",
		},
		{
			name:  "json number",
			value: json.Number("456"),
			want:  "456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JSONString(tt.value)
			if got != tt.want {
				t.Errorf("jsonString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestJSONInt64(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    int64
		wantErr bool
	}{
		{
			name:  "positive integer",
			value: json.Number("123"),
			want:  123,
		},
		{
			name:  "negative integer",
			value: json.Number("-123"),
			want:  -123,
		},
		{
			name:  "zero",
			value: json.Number("0"),
			want:  0,
		},
		{
			name:  "max int64",
			value: json.Number("9223372036854775807"),
			want:  math.MaxInt64,
		},
		{
			name:    "above max int64",
			value:   json.Number("9223372036854775808"),
			wantErr: true,
		},
		{
			name:    "below min int64",
			value:   json.Number("-9223372036854775809"),
			wantErr: true,
		},
		{
			name:    "fractional number",
			value:   json.Number("123.45"),
			wantErr: true,
		},
		{
			name:    "string",
			value:   "123",
			wantErr: true,
		},
		{
			name:    "float64",
			value:   float64(123),
			wantErr: true,
		},
		{
			name:    "bool",
			value:   true,
			wantErr: true,
		},
		{
			name:    "nil",
			value:   nil,
			wantErr: true,
		},
		{
			name:    "object",
			value:   map[string]any{"id": json.Number("123")},
			wantErr: true,
		},
		{
			name:    "array",
			value:   []any{json.Number("123")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JSONInt64(tt.value)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("jsonInt64(%v) returned %d, want error", tt.value, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("jsonInt64(%v) returned unexpected error: %v", tt.value, err)
			}

			if got != tt.want {
				t.Errorf("jsonInt64(%v) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}
