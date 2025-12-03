package gocantile

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/hafenkran/gocantile/tms"
)

//go:embed data/tilematrixset/*.json
var embeddedTMS embed.FS

var (
	embeddedTMSMap   map[string][]byte
	embeddedTMSNames []string
)

func init() {
	embeddedTMSMap = make(map[string][]byte)
	entries, err := embeddedTMS.ReadDir("data/tilematrixset")
	if err != nil {
		panic(fmt.Errorf("tms registry: failed to read embedded sets: %w", err))
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		raw, err := embeddedTMS.ReadFile("data/tilematrixset/" + entry.Name())
		if err != nil {
			panic(fmt.Errorf("tms registry: failed to read %s: %w", entry.Name(), err))
		}
		name := strings.TrimSuffix(entry.Name(), ".json")
		embeddedTMSMap[name] = raw
		embeddedTMSNames = append(embeddedTMSNames, name)
	}
	sort.Strings(embeddedTMSNames)
}

// AvailableTileMatrixSets returns the names of all embedded TileMatrixSets.
func AvailableTileMatrixSets() []string {
	out := make([]string, len(embeddedTMSNames))
	copy(out, embeddedTMSNames)
	return out
}

// LoadTileMatrixSet returns an embedded TileMatrixSet by name (case-sensitive,
// without ".json").
func LoadTileMatrixSet(name string) (*TileMatrixSet, error) {
	raw, ok := embeddedTMSMap[name]
	if !ok {
		return nil, fmt.Errorf("tilematrixset %q not found", name)
	}
	var set tms.TileMatrixSet
	if err := json.Unmarshal(raw, &set); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %q: %w", name, err)
	}
	return WrapTileMatrixSet(set), nil
}
