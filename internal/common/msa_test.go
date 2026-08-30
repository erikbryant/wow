package common

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestMsaValue(t *testing.T) {
	data := map[string]any{
		"id":   json.Number("123"),
		"name": "Widget",
		"nil":  nil,
		"nested": map[string]any{
			"name": "Nested Widget",
			"deep": map[string]any{
				"value": json.Number("456"),
			},
		},
		"array": []any{
			map[string]any{
				"name": "First",
			},
		},
	}

	tests := []struct {
		name    string
		msi     any
		keys    []string
		want    any
		wantErr string
	}{
		{
			name: "top level value",
			msi:  data,
			keys: []string{"id"},
			want: json.Number("123"),
		},
		{
			name: "top level string",
			msi:  data,
			keys: []string{"name"},
			want: "Widget",
		},
		{
			name: "nested value",
			msi:  data,
			keys: []string{"nested", "name"},
			want: "Nested Widget",
		},
		{
			name: "deeply nested value",
			msi:  data,
			keys: []string{"nested", "deep", "value"},
			want: json.Number("456"),
		},
		{
			name: "existing nil value",
			msi:  data,
			keys: []string{"nil"},
			want: nil,
		},
		{
			name:    "missing top level key",
			msi:     data,
			keys:    []string{"missing"},
			wantErr: `key "missing" not found`,
		},
		{
			name:    "missing nested key",
			msi:     data,
			keys:    []string{"nested", "missing"},
			wantErr: `key "missing" not found`,
		},
		{
			name:    "intermediate value is string",
			msi:     data,
			keys:    []string{"name", "something"},
			wantErr: `cannot access key "something": value has type string`,
		},
		{
			name:    "intermediate value is number",
			msi:     data,
			keys:    []string{"id", "something"},
			wantErr: `cannot access key "something": value has type json.Number`,
		},
		{
			name:    "intermediate value is nil",
			msi:     data,
			keys:    []string{"nil", "something"},
			wantErr: `cannot access key "something": value has type <nil>`,
		},
		{
			name:    "intermediate value is slice",
			msi:     data,
			keys:    []string{"array", "something"},
			wantErr: `cannot access key "something": value has type []interface {}`,
		},
		{
			name:    "nil input",
			msi:     nil,
			keys:    []string{"anything"},
			wantErr: `cannot access key "anything": value has type <nil>`,
		},
		{
			name: "empty path",
			msi:  data,
			keys: nil,
			want: data,
		},
		{
			name: "empty path with nil input",
			msi:  nil,
			keys: nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MsaValue(tt.msi, tt.keys)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("MsaValue() error = nil, want %q", tt.wantErr)
				}

				if err.Error() != tt.wantErr {
					t.Fatalf(
						"MsaValue() error = %q, want %q",
						err,
						tt.wantErr,
					)
				}

				if got != nil {
					t.Fatalf(
						"MsaValue() value = %#v, want nil",
						got,
					)
				}

				return
			}

			if err != nil {
				t.Fatalf("MsaValue() unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf(
					"MsaValue() = %#v, want %#v",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestMsaValued(t *testing.T) {
	data := map[string]any{
		"id":   json.Number("123"),
		"name": "Widget",
		"nil":  nil,
		"nested": map[string]any{
			"name": "Nested Widget",
			"nil":  nil,
		},
	}

	tests := []struct {
		name    string
		msi     any
		keys    []string
		d       any
		want    any
		wantErr string
	}{
		{
			name: "existing value",
			msi:  data,
			keys: []string{"id"},
			d:    json.Number("999"),
			want: json.Number("123"),
		},
		{
			name: "existing string",
			msi:  data,
			keys: []string{"name"},
			d:    "Default",
			want: "Widget",
		},
		{
			name: "existing nested value",
			msi:  data,
			keys: []string{"nested", "name"},
			d:    "Default",
			want: "Nested Widget",
		},
		{
			name: "existing nil uses default",
			msi:  data,
			keys: []string{"nil"},
			d:    "Default",
			want: "Default",
		},
		{
			name: "existing nested nil uses default",
			msi:  data,
			keys: []string{"nested", "nil"},
			d:    "Default",
			want: "Default",
		},
		{
			name:    "missing key returns error",
			msi:     data,
			keys:    []string{"missing"},
			d:       "Default",
			wantErr: `key "missing" not found`,
		},
		{
			name:    "missing nested key returns error",
			msi:     data,
			keys:    []string{"nested", "missing"},
			d:       "Default",
			wantErr: `key "missing" not found`,
		},
		{
			name:    "cannot traverse string",
			msi:     data,
			keys:    []string{"name", "missing"},
			d:       "Default",
			wantErr: `cannot access key "missing": value has type string`,
		},
		{
			name:    "cannot traverse nil",
			msi:     data,
			keys:    []string{"nil", "missing"},
			d:       "Default",
			wantErr: `cannot access key "missing": value has type <nil>`,
		},
		{
			name:    "nil input",
			msi:     nil,
			keys:    []string{"anything"},
			d:       "Default",
			wantErr: `cannot access key "anything": value has type <nil>`,
		},
		{
			name: "empty path",
			msi:  data,
			keys: nil,
			d:    "Default",
			want: data,
		},
		{
			name: "empty path with nil input",
			msi:  nil,
			keys: nil,
			d:    "Default",
			want: "Default",
		},
		{
			name:    "default can be nil",
			msi:     data,
			keys:    []string{"missing"},
			d:       nil,
			wantErr: `key "missing" not found`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MsaValued(tt.msi, tt.keys, tt.d)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("MsaValued() error = nil, want %q", tt.wantErr)
				}

				if err.Error() != tt.wantErr {
					t.Fatalf(
						"MsaValued() error = %q, want %q",
						err,
						tt.wantErr,
					)
				}

				return
			}

			if err != nil {
				t.Fatalf("MsaValued() unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf(
					"MsaValued() = %#v, want %#v",
					got,
					tt.want,
				)
			}
		})
	}
}
