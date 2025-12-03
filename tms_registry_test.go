package gocantile

import (
	"strings"
	"testing"
)

func TestRegistryLoad(t *testing.T) {
	names := AvailableTileMatrixSets()
	if len(names) == 0 {
		t.Fatalf("expected embedded tile matrix sets")
	}
	found := false
	for _, n := range names {
		if n == "WebMercatorQuad" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected WebMercatorQuad in registry, got %v", names)
	}

	set, err := LoadTileMatrixSet("WebMercatorQuad")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(set.TileMatrices) == 0 {
		t.Fatalf("expected tile matrices")
	}
}

func TestRegistryLoadMissing(t *testing.T) {
	if _, err := LoadTileMatrixSet("does-not-exist"); err == nil {
		t.Fatalf("expected error for missing entry")
	}
}

func TestRegistryLoadUnmarshalError(t *testing.T) {
	const name = "Broken"
	orig, had := embeddedTMSMap[name]
	embeddedTMSMap[name] = []byte(`{invalid json`)
	defer func() {
		if had {
			embeddedTMSMap[name] = orig
		} else {
			delete(embeddedTMSMap, name)
		}
	}()

	if _, err := LoadTileMatrixSet(name); err == nil || !strings.Contains(err.Error(), "failed to unmarshal") {
		t.Fatalf("expected unmarshal error, got: %v", err)
	}
}

func TestProjectorFromEmbeddedTMS(t *testing.T) {
	set, err := LoadTileMatrixSet("WebMercatorQuad")
	if err != nil {
		t.Fatalf("load TMS: %v", err)
	}
	p, err := ProjectorFromTMS(set)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if p.TargetCRS != "EPSG:3857" {
		t.Fatalf("unexpected target CRS: %s", p.TargetCRS)
	}
}
