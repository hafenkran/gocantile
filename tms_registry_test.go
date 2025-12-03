package gocantile

import "testing"

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
