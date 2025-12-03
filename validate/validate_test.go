package validate

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"github.com/hafenkran/gocantile"
)

func TestValidateTileMatrixSetSuccess(t *testing.T) {
	tms, err := gocantile.LoadTileMatrixSet("WebMercatorQuad")
	if err != nil {
		t.Fatalf("load embedded: %v", err)
	}
	if err := ValidateTileMatrixSet(tms.TileMatrixSet); err != nil {
		t.Fatalf("expected valid TileMatrixSet, got: %v", err)
	}
}

func TestValidateTileMatrixSetFail(t *testing.T) {
	// Missing required tileMatrices field.
	raw := []byte(`{"crs":"EPSG:3857"}`)
	if err := ValidateTileMatrixSetJSON(raw); err == nil {
		t.Fatalf("expected validation error for missing tileMatrices")
	}
}

func TestValidateTileMatrixSetMalformedJSON(t *testing.T) {
	if err := ValidateTileMatrixSetJSON([]byte(`{`)); err == nil {
		t.Fatalf("expected error for malformed JSON")
	}
}

func TestValidateTileSetSuccess(t *testing.T) {
	raw, err := os.ReadFile("../data/tileset/tiles.WebMercatorQuad.json")
	if err != nil {
		t.Fatalf("read tileset: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal tileset: %v", err)
	}
	// Fix geometryDimension type if provided as string in sample.
	if layers, ok := m["layers"].([]interface{}); ok {
		for i, layer := range layers {
			lmap, ok := layer.(map[string]interface{})
			if !ok {
				continue
			}
			switch gd := lmap["geometryDimension"].(type) {
			case string:
				if n, err := strconv.Atoi(gd); err == nil {
					lmap["geometryDimension"] = n
				}
			case float64:
				lmap["geometryDimension"] = int(gd)
			}
			layers[i] = lmap
		}
		m["layers"] = layers
	}
	fixed, _ := json.Marshal(m)
	if err := ValidateTileSetJSON(fixed); err != nil {
		t.Fatalf("expected valid TileSet after fix, got: %v", err)
	}
}

func TestValidateTileSetFail(t *testing.T) {
	// Missing required fields.
	raw := []byte(`{"links":[]}`)
	if err := ValidateTileSetJSON(raw); err == nil {
		t.Fatalf("expected validation error for missing tileMatrixSet")
	}
}
