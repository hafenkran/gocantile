package grid

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hafenkran/gocantile/tms"
	"github.com/paulmach/orb"
)

func newTestAdapter() TileMatrix {
	return TileMatrix{
		TM: tms.TileMatrix{
			CellSize:      24000,
			TileWidth:     1,
			TileHeight:    1,
			MatrixWidth:   2,
			MatrixHeight:  2,
			PointOfOrigin: []float64{0, 48000},
		},
	}
}

func newMercatorAdapter() TileMatrix {
	// WebMercator zoom 0 from OGC definition.
	return TileMatrix{
		TM: tms.TileMatrix{
			CellSize:      156543.03392804097,
			TileWidth:     256,
			TileHeight:    256,
			MatrixWidth:   1,
			MatrixHeight:  1,
			PointOfOrigin: []float64{-20037508.3427892, 20037508.3427892},
		},
	}
}

func loadTileMatrixSet(t *testing.T, name string) tms.TileMatrixSet {
	t.Helper()
	path := filepath.Join("..", "data", "tilematrixset", name+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var set tms.TileMatrixSet
	if err := json.Unmarshal(raw, &set); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return set
}

func TestTileForXY(t *testing.T) {
	adapter := newTestAdapter()

	tile, ok := adapter.TileForXY(12000, 36000) // center of tile (0,0)
	if !ok {
		t.Fatalf("expected tile, got ok=false")
	}
	if tile.Col != 0 || tile.Row != 0 {
		t.Fatalf("expected tile (0,0), got (%d,%d)", tile.Col, tile.Row)
	}

	if _, ok := adapter.TileForXY(-1000, 36000); ok {
		t.Fatalf("expected out-of-range point to return ok=false")
	}
}

func almostEqual(a, b, eps float64) bool {
	if a > b {
		return a-b < eps
	}
	return b-a < eps
}

func TestTileForLonLatAndBoundsLonLat(t *testing.T) {
	adapter := newMercatorAdapter()
	proj := NewWGS84Projector("EPSG:3857")

	tile, ok := adapter.TileForLonLat(0, 0, proj)
	if !ok {
		t.Fatalf("expected tile")
	}
	if tile.Col != 0 || tile.Row != 0 {
		t.Fatalf("expected tile (0,0), got (%d,%d)", tile.Col, tile.Row)
	}

	llBounds, err := adapter.BoundsForTileLonLat(tile, proj)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !almostEqual(llBounds.MinX, -180, 1e-6) || !almostEqual(llBounds.MaxX, 180, 1e-6) {
		t.Fatalf("unexpected lon bounds: %+v", llBounds)
	}
	// Lat bounds should clamp to WebMercatorMaxLat.
	if !almostEqual(llBounds.MinY, -85.0511287, 1e-4) || !almostEqual(llBounds.MaxY, 85.0511287, 1e-4) {
		t.Fatalf("unexpected lat bounds: %+v", llBounds)
	}
}

func TestBoundsForTile(t *testing.T) {
	adapter := newTestAdapter()

	b, err := adapter.BoundsForTile(TileIndex{Col: 1, Row: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if b.MinX != 24000 || b.MaxX != 48000 || b.MinY != 0 || b.MaxY != 24000 {
		t.Fatalf("unexpected bounds: %+v", b)
	}

	if _, err := adapter.BoundsForTile(TileIndex{Col: 2, Row: 0}); err == nil {
		t.Fatalf("expected error for out-of-range tile")
	}
}

func TestResolutionAndTileSize(t *testing.T) {
	adapter := newTestAdapter()

	if res := adapter.Resolution(); res != 24000 {
		t.Fatalf("unexpected resolution: %f", res)
	}
	w, h := adapter.TileSize()
	if w != 24000 || h != 24000 {
		t.Fatalf("unexpected tile size: %f x %f", w, h)
	}
}

func TestCellSizeInternal(t *testing.T) {
	adapter := newTestAdapter()
	if got := adapter.cellSize(); got != adapter.TM.CellSize {
		t.Fatalf("expected cell size %f, got %f", adapter.TM.CellSize, got)
	}
	if adapter.cellSize() != adapter.Resolution() {
		t.Fatalf("expected cell size to match Resolution")
	}
}

func TestTileRangeForBounds(t *testing.T) {
	adapter := newTestAdapter()

	tr, ok := adapter.TileRangeForBounds(Bounds{
		MinX: -1000, MinY: -1000,
		MaxX: 50000, MaxY: 50000,
	})
	if !ok {
		t.Fatalf("expected range")
	}
	if tr.MinCol != 0 || tr.MaxCol != 1 || tr.MinRow != 0 || tr.MaxRow != 1 {
		t.Fatalf("unexpected tile range: %+v", tr)
	}

	if _, ok := adapter.TileRangeForBounds(Bounds{
		MinX: 100000, MinY: 100000,
		MaxX: 120000, MaxY: 120000,
	}); ok {
		t.Fatalf("expected empty range")
	}
}

func TestTilesForBounds(t *testing.T) {
	adapter := newTestAdapter()
	tiles := adapter.TilesForBounds(Bounds{
		MinX: -1000, MinY: -1000,
		MaxX: 50000, MaxY: 50000,
	})
	if len(tiles) != 4 {
		t.Fatalf("expected 4 tiles, got %d", len(tiles))
	}
}

func TestTilesForGeometry(t *testing.T) {
	adapter := newTestAdapter()
	poly := orb.Polygon{
		{
			{1000, 47000},
			{23000, 47000},
			{23000, 35000},
			{1000, 35000},
			{1000, 47000},
		},
	}
	tiles := adapter.TilesForGeometry(poly)
	if len(tiles) != 1 {
		t.Fatalf("expected 1 tile, got %d", len(tiles))
	}
	if tiles[0].Col != 0 || tiles[0].Row != 0 {
		t.Fatalf("expected tile (0,0), got (%d,%d)", tiles[0].Col, tiles[0].Row)
	}

	// Geometry crossing two tiles should return both.
	poly2 := orb.Polygon{
		{
			{23000, 47000},
			{26000, 47000},
			{26000, 35000},
			{23000, 35000},
			{23000, 47000},
		},
	}
	tiles = adapter.TilesForGeometry(poly2)
	if len(tiles) != 2 {
		t.Fatalf("expected 2 tiles, got %d", len(tiles))
	}
}

func TestTilesForGeometryWebMercator(t *testing.T) {
	// Use WebMercator zoom 1 bounds (~world split into 2x2).
	set := loadTileMatrixSet(t, "WebMercatorQuad")
	if len(set.TileMatrices) < 2 {
		t.Fatalf("expected at least 2 zoom levels")
	}
	adapter := TileMatrix{TM: set.TileMatrices[1]}

	// Polygon around the prime meridian near equator, should hit the right tile
	// at zoom 1 (col 1, row 1 in WebMercator origin top-left).
	poly := orb.Polygon{
		{
			{1000, -1000},
			{1000000, -1000},
			{1000000, -1000000},
			{1000, -1000000},
			{1000, -1000},
		},
	}
	tiles := adapter.TilesForGeometry(poly)
	expected := map[TileIndex]struct{}{
		{Col: 1, Row: 1}: {},
	}
	if len(tiles) != len(expected) {
		t.Fatalf("expected %d tile(s), got %d", len(expected), len(tiles))
	}
	for _, ti := range tiles {
		if _, ok := expected[ti]; !ok {
			t.Fatalf("unexpected tile %+v", ti)
		}
	}
}

func TestTilesForGeometryWebMercatorManyTiles(t *testing.T) {
	poly := orb.Polygon{
		{
			{-15_000_000, 15_000_000},
			{15_000_000, 15_000_000},
			{15_000_000, -15_000_000},
			{-15_000_000, -15_000_000},
			{-15_000_000, 15_000_000},
		},
	}
	level := 4 // Use WebMercator zoom 4 (16x16 tiles).
	expectedCount := 144

	set := loadTileMatrixSet(t, "WebMercatorQuad")
	if len(set.TileMatrices) == 0 {
		t.Fatalf("expected zoom levels")
	}
	// zoom level 4 -> 16x16 tiles
	if len(set.TileMatrices) <= level {
		t.Fatalf("expected at least %d zoom levels, got %d", level+1, len(set.TileMatrices))
	}
	adapter := TileMatrix{TM: set.TileMatrices[level]}

	tr, ok := adapter.TileRangeForBounds(Bounds{
		MinX: poly.Bound().Min[0],
		MinY: poly.Bound().Min[1],
		MaxX: poly.Bound().Max[0],
		MaxY: poly.Bound().Max[1],
	})
	if !ok {
		t.Fatalf("expected tile range for bounds")
	}

	tiles := adapter.TilesForGeometry(poly)
	if len(tiles) != expectedCount {
		t.Fatalf("expected %d tiles (range %+v), got %d", expectedCount, tr, len(tiles))
	}
}

func TestBottomLeftOrigin(t *testing.T) {
	adapter := TileMatrix{
		TM: tms.TileMatrix{
			CellSize:       1,
			TileWidth:      1,
			TileHeight:     1,
			MatrixWidth:    2,
			MatrixHeight:   2,
			PointOfOrigin:  []float64{0, 0},
			CornerOfOrigin: tms.TileMatrixJsonCornerOfOriginBottomLeft,
		},
	}
	tile, ok := adapter.TileForXY(0.5, 0.5)
	if !ok {
		t.Fatalf("expected tile")
	}
	if tile.Col != 0 || tile.Row != 0 {
		t.Fatalf("expected (0,0), got %+v", tile)
	}
	b, err := adapter.BoundsForTile(tile)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if b.MinY != 0 {
		t.Fatalf("expected minY 0, got %f", b.MinY)
	}
}

func TestVariableMatrixWidths(t *testing.T) {
	adapter := TileMatrix{
		TM: tms.TileMatrix{
			CellSize:      1,
			TileWidth:     1,
			TileHeight:    1,
			MatrixWidth:   4,
			MatrixHeight:  4,
			PointOfOrigin: []float64{0, 4},
			VariableMatrixWidths: []tms.VariableMatrixWidthJson{
				{Coalesce: 2, MinTileRow: 0, MaxTileRow: 1},
			},
		},
	}
	// Top rows coalesce 2 tiles into 1, effective tile columns per row = 2.
	b, err := adapter.BoundsForTile(TileIndex{Col: 1, Row: 0})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if b.MaxX != 4 { // 2 tiles * 2 units each
		t.Fatalf("expected maxX 4, got %f", b.MaxX)
	}
	if _, ok := adapter.TileForXY(5, 3.5); ok {
		t.Fatalf("expected out-of-range for coalesced width")
	}
}
