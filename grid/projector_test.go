package grid

import (
	"testing"

	"github.com/hafenkran/gocantile/tms"
	"github.com/paulmach/orb"
)

func TestProjProjectorForwardInverse(t *testing.T) {
	p := NewWGS84Projector("EPSG:3857")

	x, y, err := p.Forward(0, 0)
	if err != nil {
		t.Fatalf("forward err: %v", err)
	}
	if x == 0 && y == 0 {
		// fine, origin maps to origin.
	} else {
		t.Fatalf("expected origin to map to origin, got %f %f", x, y)
	}

	lon, lat, err := p.Inverse(x, y)
	if err != nil {
		t.Fatalf("inverse err: %v", err)
	}
	if diff := abs(lon - 0); diff > 1e-9 {
		t.Fatalf("unexpected lon: %f", lon)
	}
	if diff := abs(lat - 0); diff > 1e-9 {
		t.Fatalf("unexpected lat: %f", lat)
	}
}

func TestProjectorFromTMS(t *testing.T) {
	set := tms.TileMatrixSet{
		Crs: "EPSG:3857",
	}
	p, err := ProjectorFromTMS(set)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if p.TargetCRS != "EPSG:3857" {
		t.Fatalf("unexpected target CRS: %s", p.TargetCRS)
	}
}

func TestProjectorFromTMSURI(t *testing.T) {
	set := tms.TileMatrixSet{
		Crs: map[string]interface{}{"uri": "http://www.opengis.net/def/crs/EPSG/0/32632"},
	}
	p, err := ProjectorFromTMS(set)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if p.TargetCRS != "EPSG:32632" {
		t.Fatalf("unexpected target CRS: %s", p.TargetCRS)
	}
}

func TestProjectorFromTMSMissingCRS(t *testing.T) {
	set := tms.TileMatrixSet{}
	if _, err := ProjectorFromTMS(set); err == nil {
		t.Fatalf("expected error for missing CRS")
	}
}

func TestProjectGeometryPoint(t *testing.T) {
	p := orb.Point{0, 0}
	out, err := ProjectGeometry(p, "EPSG:4326", "EPSG:3857")
	if err != nil {
		t.Fatalf("project point: %v", err)
	}
	pt, ok := out.(orb.Point)
	if !ok {
		t.Fatalf("expected orb.Point")
	}
	if pt[0] != 0 || pt[1] != 0 {
		t.Fatalf("expected origin, got %+v", pt)
	}
}

func TestProjectGeometryPolygon(t *testing.T) {
	poly := orb.Polygon{
		{
			{-1, -1},
			{1, -1},
			{1, 1},
			{-1, 1},
			{-1, -1},
		},
	}
	out, err := ProjectGeometry(poly, "EPSG:4326", "EPSG:3857")
	if err != nil {
		t.Fatalf("project polygon: %v", err)
	}
	pj, ok := out.(orb.Polygon)
	if !ok {
		t.Fatalf("expected orb.Polygon")
	}
	pt := pj[0][0]
	if !approxEqual(pt[0], -111319.49079327357, 1e-3) || !approxEqual(pt[1], -111325.14286638486, 1e-3) {
		t.Fatalf("unexpected projected coord %+v", pt)
	}
}

func TestProjectGeometryCollection(t *testing.T) {
	coll := orb.Collection{
		orb.Point{0, 0},
		orb.LineString{{0, 0}, {1, 1}},
	}
	out, err := ProjectGeometry(coll, "EPSG:4326", "EPSG:3857")
	if err != nil {
		t.Fatalf("project collection: %v", err)
	}
	pj, ok := out.(orb.Collection)
	if !ok {
		t.Fatalf("expected orb.Collection")
	}
	pt := pj[0].(orb.Point)
	if pt[0] != 0 || pt[1] != 0 {
		t.Fatalf("expected origin point, got %+v", pt)
	}
	ls := pj[1].(orb.LineString)
	expX := 111319.49079327357
	expY := 111325.14286638486
	if !approxEqual(ls[1][0], expX, 1e-3) || !approxEqual(ls[1][1], expY, 1e-3) {
		t.Fatalf("unexpected line endpoint %+v", ls[1])
	}
}

func TestProjectGeometrySameCRS(t *testing.T) {
	p := orb.Point{1, 2}
	out, err := ProjectGeometry(p, "EPSG:4326", "EPSG:4326")
	if err != nil {
		t.Fatalf("project same CRS: %v", err)
	}
	pt := out.(orb.Point)
	if pt != p {
		t.Fatalf("expected unchanged point, got %+v", pt)
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func approxEqual(a, b, eps float64) bool {
	if a > b {
		return a-b < eps
	}
	return b-a < eps
}
