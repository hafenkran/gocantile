package gocantile

import (
	"testing"

	"github.com/paulmach/orb"
)

func loadWebMercatorQuad(t *testing.T) *TileMatrixSet {
	t.Helper()
	tms, err := LoadTileMatrixSet("WebMercatorQuad")
	if err != nil {
		t.Fatalf("load TMS: %v", err)
	}
	return tms
}

func TestTileMatrixSetBasics(t *testing.T) {
	tms := loadWebMercatorQuad(t)

	if tms.MinZoom() != 0 {
		t.Fatalf("expected min zoom 0, got %d", tms.MinZoom())
	}
	if tms.MaxZoom() < 1 {
		t.Fatalf("unexpected max zoom %d", tms.MaxZoom())
	}

	res, err := tms.ResolutionForZoom(0)
	if err != nil {
		t.Fatalf("resolution err: %v", err)
	}
	if res <= 0 {
		t.Fatalf("expected positive resolution")
	}

	z, err := tms.ZoomForResolution(res, 0)
	if err != nil {
		t.Fatalf("zoom for res err: %v", err)
	}
	if z != 0 {
		t.Fatalf("expected zoom 0, got %d", z)
	}
}

func TestTileMatrixSetTileForLonLat(t *testing.T) {
	tms := loadWebMercatorQuad(t)

	tile, ok, err := tms.TileForLonLat(0, 0, 0, nil)
	if err != nil {
		t.Fatalf("tile err: %v", err)
	}
	if !ok {
		t.Fatalf("expected tile")
	}
	if tile.Col != 0 || tile.Row != 0 {
		t.Fatalf("expected (0,0), got %+v", tile)
	}
}

func TestTileMatrixSetBoundsHelpers(t *testing.T) {
	tms := loadWebMercatorQuad(t)

	bbox, err := tms.XYBBox()
	if err != nil {
		t.Fatalf("xy bbox err: %v", err)
	}
	if bbox.MinX >= 0 || bbox.MaxX <= 0 {
		t.Fatalf("unexpected bbox %+v", bbox)
	}

	tile := Tile{Zoom: 4, TileIndex: TileIndex{Col: 10, Row: 10}}
	xyb, err := tms.XYBounds(tile)
	if err != nil {
		t.Fatalf("xy bounds err: %v", err)
	}
	if xyb.MinX >= xyb.MaxX || xyb.MinY >= xyb.MaxY {
		t.Fatalf("invalid xy bounds %+v", xyb)
	}

	llb, err := tms.Bounds(tile, nil)
	if err != nil {
		t.Fatalf("lon/lat bounds err: %v", err)
	}
	if llb.MinX >= llb.MaxX || llb.MinY >= llb.MaxY {
		t.Fatalf("invalid lon/lat bounds %+v", llb)
	}
}

func TestTileMatrixSetTilesForGeometryMultiZoom(t *testing.T) {
	tms := loadWebMercatorQuad(t)

	poly := orb.Polygon{
		{
			{-10_000_000, 10_000_000},
			{10_000_000, 10_000_000},
			{10_000_000, -10_000_000},
			{-10_000_000, -10_000_000},
			{-10_000_000, 10_000_000},
		},
	}
	tiles, err := tms.TilesForGeometry(poly, 0, 2, 0)
	if err != nil {
		t.Fatalf("tiles err: %v", err)
	}
	if len(tiles) == 0 {
		t.Fatalf("expected tiles, got none")
	}
	seenZooms := map[int]bool{}
	for _, ti := range tiles {
		seenZooms[ti.Zoom] = true
	}
	if len(seenZooms) < 2 {
		t.Fatalf("expected tiles across multiple zoom levels, got %d", len(seenZooms))
	}
}

func TestTileMatrixSetTilesForGeometryMultiZoomLinesAndBuffer(t *testing.T) {
	tms := loadWebMercatorQuad(t)

	line := orb.LineString{
		{-1_000_000, 0},
		{1_000_000, 0},
	}
	tilesNoBuffer, err := tms.TilesForGeometry(line, 0, 1, 0)
	if err != nil {
		t.Fatalf("tiles err: %v", err)
	}
	if len(tilesNoBuffer) == 0 {
		t.Fatalf("expected tiles without buffer")
	}

	tilesWithBuffer, err := tms.TilesForGeometry(line, 0, 1, 500_000)
	if err != nil {
		t.Fatalf("tiles err buffer: %v", err)
	}
	if len(tilesWithBuffer) < len(tilesNoBuffer) {
		t.Fatalf("expected buffered tiles >= without buffer (%d < %d)", len(tilesWithBuffer), len(tilesNoBuffer))
	}

	seenZooms := map[int]bool{}
	for _, ti := range tilesWithBuffer {
		seenZooms[ti.Zoom] = true
	}
	if len(seenZooms) != 2 {
		t.Fatalf("expected tiles across 2 zoom levels, got %d", len(seenZooms))
	}
}

func TestTileMatrixSetTilesForGeometryZoom11(t *testing.T) {
	tms := loadWebMercatorQuad(t)

	bbox := orb.Polygon{{
		{13.088626854245092, 52.416237574678775},
		{13.141214943591734, 52.39375459376379},
		{13.19066498925963, 52.41365065742863},
		{13.278008354068618, 52.4059506314081},
		{13.37902892106618, 52.38669145045344},
		{13.423224087269745, 52.41301288273135},
		{13.553708074084255, 52.386044489654665},
		{13.678926847959389, 52.37641341523192},
		{13.765208845079911, 52.438683015289115},
		{13.620002424454327, 52.471381583541444},
		{13.61261837506575, 52.543763191663174},
		{13.503190877114008, 52.60578878619489},
		{13.476885923873368, 52.66708748539642},
		{13.42112096918703, 52.639003383600425},
		{13.354831264738522, 52.626870625748666},
		{13.301168457214828, 52.640281184385856},
		{13.295908937459757, 52.64921883591316},
		{13.242233243360516, 52.63134254685107},
		{13.219061979177553, 52.623679237434146},
		{13.2211874886423, 52.58853085147888},
		{13.202245296940873, 52.59492323998214},
		{13.156976833112651, 52.59556147405988},
		{13.146444452978045, 52.56102746429869},
		{13.122258221278315, 52.518147724621144},
		{13.169612059923594, 52.509822079251165},
		{13.142270264184475, 52.491892517900055},
		{13.112820187700976, 52.45536397180163},
		{13.121249728274336, 52.44062010113532},
		{13.088626854245092, 52.416237574678775},
	}}

	proj := NewWGS84Projector("EPSG:3857")
	var projected orb.Polygon
	for _, ring := range bbox {
		var pr orb.Ring
		for _, pt := range ring {
			x, y, err := proj.Forward(pt[0], pt[1])
			if err != nil {
				t.Fatalf("projection err: %v", err)
			}
			pr = append(pr, orb.Point{x, y})
		}
		projected = append(projected, pr)
	}
	tiles, err := tms.TilesForGeometry(projected, 11, 11, 0)
	if err != nil {
		t.Fatalf("tiles err: %v", err)
	}
	if len(tiles) != 15 {
		t.Fatalf("expected 15 tiles for polygon. got %d", len(tiles))
	}
	for _, ti := range tiles {
		if ti.Zoom != 11 {
			t.Fatalf("expected zoom 11, got %d", ti.Zoom)
		}
	}
}
