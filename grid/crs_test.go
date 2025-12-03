package grid

import (
	"strings"
	"testing"

	"github.com/hafenkran/gocantile/tms"
)

func TestExtractCRSEPSGString(t *testing.T) {
	set := tms.TileMatrixSet{Crs: "EPSG:3857"}
	crs, err := ExtractCRS(set)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if crs != "EPSG:3857" {
		t.Fatalf("unexpected crs: %s", crs)
	}
}

func TestExtractCRSURI(t *testing.T) {
	set := tms.TileMatrixSet{Crs: map[string]interface{}{
		"uri": "http://www.opengis.net/def/crs/EPSG/0/32632",
	}}
	crs, err := ExtractCRS(set)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if crs != "EPSG:32632" {
		t.Fatalf("unexpected crs: %s", crs)
	}
}

func TestExtractCRSURN(t *testing.T) {
	set := tms.TileMatrixSet{Crs: "urn:ogc:def:crs:EPSG::4326"}
	crs, err := ExtractCRS(set)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if crs != "EPSG:4326" {
		t.Fatalf("unexpected crs: %s", crs)
	}
}

func TestExtractCRSProjJSON(t *testing.T) {
	set := tms.TileMatrixSet{Crs: map[string]interface{}{
		"wkt": map[string]interface{}{
			"type": "ProjectedCRS",
			"name": "Dummy",
		},
	}}
	crs, err := ExtractCRS(set)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(crs, `"ProjectedCRS"`) {
		t.Fatalf("expected projjson string, got %s", crs)
	}
}

func TestExtractCRSWKTUnsupported(t *testing.T) {
	set := tms.TileMatrixSet{Crs: map[string]interface{}{
		"wkt": []int{1, 2, 3},
	}}
	if _, err := ExtractCRS(set); err == nil {
		t.Fatalf("expected error for unsupported wkt type")
	}
}

func TestExtractCRSUnsupportedType(t *testing.T) {
	set := tms.TileMatrixSet{Crs: []int{1, 2}}
	if _, err := ExtractCRS(set); err == nil {
		t.Fatalf("expected error for unsupported crs type")
	}
}
