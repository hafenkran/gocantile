package validate

import (
	"embed"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/hafenkran/gocantile"
)

var originalSchemaFS = gocantile.SchemaFS

func TestValidateTileMatrixSetSuccess(t *testing.T) {
	resetCompileState(t, originalSchemaFS)
	tms, err := gocantile.LoadTileMatrixSet("WebMercatorQuad")
	if err != nil {
		t.Fatalf("load embedded: %v", err)
	}
	if err := ValidateTileMatrixSet(tms.TileMatrixSet); err != nil {
		t.Fatalf("expected valid TileMatrixSet, got: %v", err)
	}
}

func TestValidateTileMatrixSetFail(t *testing.T) {
	resetCompileState(t, originalSchemaFS)
	// Missing required tileMatrices field.
	raw := []byte(`{"crs":"EPSG:3857"}`)
	if err := ValidateTileMatrixSetJSON(raw); err == nil {
		t.Fatalf("expected validation error for missing tileMatrices")
	}
}

func TestValidateTileMatrixSetMalformedJSON(t *testing.T) {
	resetCompileState(t, originalSchemaFS)
	if err := ValidateTileMatrixSetJSON([]byte(`{`)); err == nil {
		t.Fatalf("expected error for malformed JSON")
	}
}

func TestValidateTileSetSuccess(t *testing.T) {
	resetCompileState(t, originalSchemaFS)
	m := loadSampleTileSet(t)
	fixed, _ := json.Marshal(m)
	if err := ValidateTileSetJSON(fixed); err != nil {
		t.Fatalf("expected valid TileSet after fix, got: %v", err)
	}
}

func TestValidateTileSetFromStruct(t *testing.T) {
	resetCompileState(t, originalSchemaFS)
	m := loadSampleTileSet(t)
	if err := ValidateTileSet(m); err != nil {
		t.Fatalf("expected valid TileSet from struct, got: %v", err)
	}
}

func TestValidateTileSetMarshalError(t *testing.T) {
	resetCompileState(t, originalSchemaFS)
	type bad struct {
		C chan int
	}
	err := ValidateTileSet(bad{})
	if err == nil {
		t.Fatalf("expected marshal error for unsupported type")
	}
	if !strings.Contains(err.Error(), "marshal TileSet") {
		t.Fatalf("expected marshal error, got: %v", err)
	}
}

func TestValidateTileSetFail(t *testing.T) {
	resetCompileState(t, originalSchemaFS)
	// Missing required fields.
	raw := []byte(`{"links":[]}`)
	if err := ValidateTileSetJSON(raw); err == nil {
		t.Fatalf("expected validation error for missing tileMatrixSet")
	}
}

func TestValidateWithSchemaMissing(t *testing.T) {
	resetCompileState(t, originalSchemaFS)
	if err := validateWithSchema("missing.json", []byte(`{}`)); err == nil {
		t.Fatalf("expected missing schema error")
	}
}

func TestCompileSchemasReadDirError(t *testing.T) {
	resetCompileState(t, embed.FS{})
	defer resetCompileState(t, originalSchemaFS)

	if err := compileSchemas(); err == nil || !strings.Contains(err.Error(), "read embedded schema dir") {
		t.Fatalf("expected read dir error, got: %v", err)
	}
}

func loadSampleTileSet(t *testing.T) map[string]interface{} {
	t.Helper()
	resetCompileState(t, originalSchemaFS)
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
	return m
}

func resetCompileState(t *testing.T, fs embed.FS) {
	t.Helper()
	compileOnce = sync.Once{}
	compileErr = nil
	schemaCache = nil
	gocantile.SchemaFS = fs
}
