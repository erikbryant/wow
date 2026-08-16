package appearanceset

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/erikbryant/wow/internal/persist"
)

func TestContains(t *testing.T) {
	as := &Persistence{
		Persistence: persist.New[int64, bool](filepath.Join(t.TempDir(), "appearances")),
	}

	as.Set(100, true)
	as.Set(200, false)

	tests := []struct {
		name string
		ids  []int64
		want bool
	}{
		{
			name: "contains matching appearance",
			ids:  []int64{100},
			want: true,
		},
		{
			name: "contains matching appearance among nonmatches",
			ids:  []int64{1, 2, 100, 3},
			want: true,
		},
		{
			name: "false value is not in set",
			ids:  []int64{200},
			want: false,
		},
		{
			name: "missing appearance",
			ids:  []int64{300},
			want: false,
		},
		{
			name: "empty list",
			ids:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := as.Contains(tt.ids); got != tt.want {
				t.Fatalf("Contains(%v) = %v, want %v", tt.ids, got, tt.want)
			}
		})
	}
}

func TestContainsDoesNotModifyPersistence(t *testing.T) {
	as := &Persistence{
		Persistence: persist.New[int64, bool](filepath.Join(t.TempDir(), "appearances")),
	}

	as.Set(100, true)
	as.Set(200, false)

	before := as.Len()
	if !as.Contains([]int64{999, 100, 200}) {
		t.Fatal("Contains returned false for an appearance that is in a set")
	}

	if got := as.Len(); got != before {
		t.Fatalf("Contains changed persistence length from %d to %d", before, got)
	}
}

func TestNewLoadsPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appearances")

	stored := persist.New[int64, bool](path)
	stored.Set(100, true)
	stored.Set(200, false)
	if err := stored.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	as, err := New(path)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if got, want := as.Len(), 2; got != want {
		t.Fatalf("Len() = %d, want %d", got, want)
	}

	if got := as.Contains([]int64{100}); !got {
		t.Fatal("Contains(100) = false, want true")
	}

	if got := as.Contains([]int64{200}); got {
		t.Fatal("Contains(200) = true, want false")
	}

	if as.Dirty() {
		t.Fatal("loaded persistence is dirty")
	}
}

func TestNewMissingPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist")

	as, err := New(path)
	if err == nil {
		t.Fatal("New() error = nil, want error")
	}

	if as != nil {
		t.Fatal("New() returned non-nil persistence on error")
	}

}

func TestNewCorruptPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appearances")

	stored := persist.New[int64, bool](path)
	if err := stored.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Replace the valid gob with invalid data.
	if err := os.WriteFile(path+".gob", []byte("not a gob"), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	as, err := New(path)
	if err == nil {
		t.Fatal("New() error = nil, want error")
	}

	if as != nil {
		t.Fatal("New() returned non-nil persistence on error")
	}
}
