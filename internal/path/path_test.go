package path

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreate(t *testing.T) {
	root := t.TempDir()
	if err := create(root); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{"bin", "data", "exports", "reports"} {
		info, err := os.Stat(filepath.Join(root, d))
		if err != nil || !info.IsDir() {
			t.Errorf("missing %s", d)
		}
	}
	if err := create(root); err != nil {
		t.Fatal(err)
	}
}

func TestFindRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module test\n"), 0600); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0700); err != nil {
		t.Fatal(err)
	}
	got, err := findRoot(nested)
	if err != nil || got != root {
		t.Fatalf("findRoot=%q,%v", got, err)
	}
	file := filepath.Join(nested, "x")
	_ = os.WriteFile(file, []byte{}, 0600)
	got, err = findRoot(file)
	if err != nil || got != root {
		t.Fatalf("file findRoot=%q,%v", got, err)
	}
}

func TestFindRootMissing(t *testing.T) {
	_, err := findRoot(t.TempDir())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewExplicitRoot(t *testing.T) {
	root := t.TempDir()
	p, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	checks := map[string]string{"Appearances": filepath.Join(root, "data", "appearances"), "Items": filepath.Join(root, "data", "items"), "Arbitrage": filepath.Join(root, "exports", "arbitrageLatest"), "BattlePets": filepath.Join(root, "reports", "battlePets"), "PriceCache": filepath.Join(root, "exports", "PriceCache.lua"), "RecipesNeeded": filepath.Join(root, "reports", "recipesNeeded"), "Recommendations": filepath.Join(root, "reports", "shopping"), "Secret": filepath.Join(root, "bin", "secret")}
	for name, want := range checks {
		var got string
		switch name {
		case "Appearances":
			got = p.Appearances
		case "Items":
			got = p.Items
		case "Arbitrage":
			got = p.Arbitrage
		case "BattlePets":
			got = p.BattlePets
		case "PriceCache":
			got = p.PriceCache
		case "RecipesNeeded":
			got = p.RecipesNeeded
		case "Recommendations":
			got = p.Recommendations
		case "Secret":
			got = p.Secret
		}
		if got != want {
			t.Errorf("%s=%q want %q", name, got, want)
		}
	}
}
