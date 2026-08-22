package path

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Paths struct {
	Appearances     string
	Arbitrage       string
	BattlePets      string
	Items           string
	ItemsReport     string
	PriceCache      string
	RecipesNeeded   string
	Recommendations string
	Secret          string
}

const (
	moduleFile = "go.mod"

	binDir     = "bin"
	dataDir    = "data"
	exportsDir = "exports"
	reportsDir = "reports"
)

// create creates the directories owned by the application. Safe to call more than once.
func create(rootDir string) error {
	directories := []string{
		binDir,
		dataDir,
		exportsDir,
		reportsDir,
	}

	for _, dir := range directories {
		path := filepath.Join(rootDir, dir)
		if err := os.MkdirAll(path, 0700); err != nil {
			return fmt.Errorf("create path %s: %w", path, err)
		}
	}

	return nil
}

// findRoot searches start and its ancestors for go.mod.
//
// The directory containing go.mod is considered the repository root.
//
// For example, given:
//
//	/Users/me/dev/wow/cmd/wow
//
// findRoot will search:
//
//	/Users/me/dev/wow/cmd/wow/go.mod
//	/Users/me/dev/wow/cmd/go.mod
//	/Users/me/dev/wow/go.mod
//
// and return:
//
//	/Users/me/dev/wow
func findRoot(start string) (string, error) {
	start, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("get absolute path for %q: %w", start, err)
	}

	info, err := os.Stat(start)
	if err != nil {
		return "", fmt.Errorf("stat %q: %w", start, err)
	}

	if !info.IsDir() {
		start = filepath.Dir(start)
	}

	dir := start

	for {
		modulePath := filepath.Join(dir, moduleFile)

		info, err := os.Stat(modulePath)
		if err == nil {
			if info.IsDir() {
				return "", fmt.Errorf("%q is a directory, not a file", modulePath)
			}

			return dir, nil
		}

		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("check for %q: %w", modulePath, err)
		}

		parent := filepath.Dir(dir)

		if parent == dir {
			break
		}

		dir = parent
	}

	return "", fmt.Errorf("could not find %s starting from %q", moduleFile, start)
}

// New creates a new set of directories. Either rooted at 'rootPath' or if that is empty, cwd.
func New(rootPath string) (*Paths, error) {
	var err error

	if rootPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get cwd: %w", err)
		}

		rootPath, err = findRoot(cwd)
		if err != nil {
			return nil, fmt.Errorf("find root path: %w", err)
		}
	}

	p := Paths{
		Appearances:     filepath.Join(rootPath, dataDir, "appearances"),
		Arbitrage:       filepath.Join(rootPath, exportsDir, "arbitrageLatest"),
		BattlePets:      filepath.Join(rootPath, reportsDir, "battlePets"),
		Items:           filepath.Join(rootPath, dataDir, "items"),
		ItemsReport:     filepath.Join(rootPath, reportsDir, "items"),
		PriceCache:      filepath.Join(rootPath, exportsDir, "PriceCache.lua"),
		RecipesNeeded:   filepath.Join(rootPath, reportsDir, "recipesNeeded"),
		Recommendations: filepath.Join(rootPath, reportsDir, "shopping"),
		Secret:          filepath.Join(rootPath, binDir, "secret"),
	}

	err = create(rootPath)
	if err != nil {
		return nil, fmt.Errorf("create paths: %w", err)
	}

	return &p, nil
}
