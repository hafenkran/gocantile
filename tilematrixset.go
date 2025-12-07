package gocantile

import (
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/hafenkran/gocantile/grid"
	"github.com/hafenkran/gocantile/tms"
	"github.com/paulmach/orb"
)

type TileMatrixSet struct {
	tms.TileMatrixSet

	once     sync.Once
	matrices []tms.TileMatrix
	idToZoom map[string]int
	initErr  error
}

func WrapTileMatrixSet(set tms.TileMatrixSet) *TileMatrixSet {
	return &TileMatrixSet{TileMatrixSet: set}
}

func (t *TileMatrixSet) ensureInit() error {
	t.once.Do(func() {
		if len(t.TileMatrices) == 0 {
			t.matrices = nil
			t.idToZoom = map[string]int{}
			return
		}
		mats := make([]tms.TileMatrix, len(t.TileMatrices))
		copy(mats, t.TileMatrices)

		seen := make(map[string]struct{}, len(mats))
		numericCount := 0
		for i, tm := range mats {
			if tm.Id == "" {
				t.initErr = fmt.Errorf("tile matrix at index %d missing id", i)
				return
			}
			if _, ok := seen[tm.Id]; ok {
				t.initErr = fmt.Errorf("duplicate tile matrix id %q", tm.Id)
				return
			}
			seen[tm.Id] = struct{}{}
			if _, err := parseZoom(tm.Id); err == nil {
				numericCount++
			}
		}

		if numericCount == len(mats) {
			sort.SliceStable(mats, func(i, j int) bool {
				zi, _ := parseZoom(mats[i].Id)
				zj, _ := parseZoom(mats[j].Id)
				return zi < zj
			})
			for i, tm := range mats {
				zi, _ := parseZoom(tm.Id)
				if zi != i {
					t.initErr = fmt.Errorf("numeric tile matrix id %q does not match zoom index %d", tm.Id, i)
					return
				}
			}
		}

		t.idToZoom = make(map[string]int, len(mats))
		for i, tm := range mats {
			t.idToZoom[tm.Id] = i
		}
		t.matrices = mats
	})
	return t.initErr
}

func (t *TileMatrixSet) sortedMatrices() ([]tms.TileMatrix, error) {
	if err := t.ensureInit(); err != nil {
		return nil, err
	}
	return t.matrices, nil
}

func (t *TileMatrixSet) MinZoom() int {
	mats, err := t.sortedMatrices()
	if err != nil || len(mats) == 0 {
		return 0
	}
	return 0
}

func (t *TileMatrixSet) MaxZoom() int {
	mats, err := t.sortedMatrices()
	if err != nil || len(mats) == 0 {
		return 0
	}
	return len(mats) - 1
}

func (t *TileMatrixSet) ResolutionForZoom(z int) (float64, error) {
	mats, err := t.sortedMatrices()
	if err != nil {
		return 0, err
	}
	if z < 0 || z >= len(mats) {
		return 0, fmt.Errorf("zoom %d out of range", z)
	}
	return mats[z].CellSize, nil
}

// ResolutionForID returns the cell size for the TileMatrix with the given ID.
func (t *TileMatrixSet) ResolutionForID(id string) (float64, error) {
	zoom, err := t.ZoomForID(id)
	if err != nil {
		return 0, err
	}
	return t.ResolutionForZoom(zoom)
}

func (t *TileMatrixSet) ZoomForResolution(res, tol float64) (int, error) {
	mats, err := t.sortedMatrices()
	if err != nil {
		return 0, err
	}
	if len(mats) == 0 {
		return 0, fmt.Errorf("no tile matrices")
	}
	best := 0
	bestDiff := math.MaxFloat64
	for i, tm := range mats {
		diff := math.Abs(tm.CellSize - res)
		if tm.CellSize <= res+tol && diff < bestDiff {
			best = i
			bestDiff = diff
		}
	}
	return best, nil
}

// XYBBox returns the bounding box of the TileMatrixSet in the matrix CRS.
func (t *TileMatrixSet) XYBBox() (grid.Bounds, error) {
	if t.BoundingBox != nil {
		ll := t.BoundingBox.LowerLeft
		ur := t.BoundingBox.UpperRight
		if len(ll) >= 2 && len(ur) >= 2 {
			return grid.Bounds{MinX: ll[0], MinY: ll[1], MaxX: ur[0], MaxY: ur[1]}, nil
		}
	}
	mats, err := t.sortedMatrices()
	if err != nil {
		return grid.Bounds{}, err
	}
	if len(mats) == 0 {
		return grid.Bounds{}, fmt.Errorf("no tile matrices")
	}
	adapter := grid.TileMatrix{TM: mats[0]}
	maxCol := int(math.Round(adapter.TM.MatrixWidth)) - 1
	maxRow := int(math.Round(adapter.TM.MatrixHeight)) - 1
	minTile := grid.TileIndex{Col: 0, Row: 0}
	maxTile := grid.TileIndex{Col: maxCol, Row: maxRow}
	bmin, err := adapter.BoundsForTile(minTile)
	if err != nil {
		return grid.Bounds{}, err
	}
	bmax, err := adapter.BoundsForTile(maxTile)
	if err != nil {
		return grid.Bounds{}, err
	}
	return grid.Bounds{
		MinX: bmin.MinX,
		MinY: math.Min(bmin.MinY, bmax.MinY),
		MaxX: bmax.MaxX,
		MaxY: math.Max(bmin.MaxY, bmax.MaxY),
	}, nil
}

// XYBounds returns the bounds in the matrix CRS of the given tile.
func (t *TileMatrixSet) XYBounds(tile grid.Tile) (grid.Bounds, error) {
	mats, err := t.sortedMatrices()
	if err != nil {
		return grid.Bounds{}, err
	}
	if tile.Zoom < 0 || tile.Zoom >= len(mats) {
		return grid.Bounds{}, fmt.Errorf("zoom %d out of range", tile.Zoom)
	}
	adapter := grid.TileMatrix{TM: mats[tile.Zoom]}
	return adapter.BoundsForTile(tile.TileIndex)
}

// XYBoundsForID returns tile bounds in the matrix CRS for a TileMatrix ID.
func (t *TileMatrixSet) XYBoundsForID(tile grid.TileIndex, id string) (grid.Bounds, error) {
	zoom, err := t.ZoomForID(id)
	if err != nil {
		return grid.Bounds{}, err
	}
	return t.XYBounds(grid.Tile{Zoom: zoom, TileIndex: tile})
}

// Bounds returns the lon/lat bounds (degrees) of the given tile. If p is nil,
// a projector is created from the TMS CRS.
func (t *TileMatrixSet) Bounds(tile grid.Tile, p grid.Projector) (grid.Bounds, error) {
	if p == nil {
		pp, err := grid.ProjectorFromTMS(t.TileMatrixSet)
		if err != nil {
			return grid.Bounds{}, err
		}
		p = pp
	}
	mats, err := t.sortedMatrices()
	if err != nil {
		return grid.Bounds{}, err
	}
	if tile.Zoom < 0 || tile.Zoom >= len(mats) {
		return grid.Bounds{}, fmt.Errorf("zoom %d out of range", tile.Zoom)
	}
	adapter := grid.TileMatrix{TM: mats[tile.Zoom]}
	return adapter.BoundsForTileLonLat(tile.TileIndex, p)
}

// BoundsForID returns lon/lat bounds for a TileMatrix ID.
func (t *TileMatrixSet) BoundsForID(tile grid.TileIndex, id string, p grid.Projector) (grid.Bounds, error) {
	zoom, err := t.ZoomForID(id)
	if err != nil {
		return grid.Bounds{}, err
	}
	return t.Bounds(grid.Tile{Zoom: zoom, TileIndex: tile}, p)
}

// TileForLonLat returns the tile (z/x/y) for the given lon/lat at the specified
// zoom level. If p is nil, a projector is created from the TMS CRS.
func (t *TileMatrixSet) TileForLonLat(lon, lat float64, zoom int, p grid.Projector) (grid.Tile, bool, error) {
	if p == nil {
		pp, err := grid.ProjectorFromTMS(t.TileMatrixSet)
		if err != nil {
			return grid.Tile{}, false, err
		}
		p = pp
	}
	mats, err := t.sortedMatrices()
	if err != nil {
		return grid.Tile{}, false, err
	}
	if zoom < 0 || zoom >= len(mats) {
		return grid.Tile{}, false, fmt.Errorf("zoom %d out of range", zoom)
	}
	adapter := grid.TileMatrix{TM: mats[zoom]}
	idx, ok := adapter.TileForLonLat(lon, lat, p)
	if !ok {
		return grid.Tile{}, false, nil
	}
	return grid.Tile{Zoom: zoom, TileIndex: idx}, true, nil
}

// TileForLonLatID resolves the zoom from the TileMatrix ID and returns the tile.
func (t *TileMatrixSet) TileForLonLatID(lon, lat float64, id string, p grid.Projector) (grid.Tile, bool, error) {
	zoom, err := t.ZoomForID(id)
	if err != nil {
		return grid.Tile{}, false, err
	}
	return t.TileForLonLat(lon, lat, zoom, p)
}

// TilesForGeometry returns tiles covering the geometry across zoom levels
// [minZoom, maxZoom] inclusive. Optional buffer expands the geometry bounds
// before tiling (in CRS units).
func (t *TileMatrixSet) TilesForGeometry(g orb.Geometry, minZoom, maxZoom int, buffer float64) (grid.TilesList, error) {
	if minZoom < 0 || maxZoom < minZoom {
		return nil, fmt.Errorf("invalid zoom range min=%d max=%d", minZoom, maxZoom)
	}
	mats, err := t.sortedMatrices()
	if err != nil {
		return nil, err
	}
	if maxZoom >= len(mats) {
		return nil, fmt.Errorf("max zoom %d out of range", maxZoom)
	}

	var tiles grid.TilesList
	for z := minZoom; z <= maxZoom; z++ {
		adapter := grid.TileMatrix{TM: mats[z]}
		rangeTiles := adapter.TilesForGeometry(g, buffer)
		for _, idx := range rangeTiles {
			tiles = append(tiles, grid.Tile{Zoom: z, TileIndex: idx})
		}
	}
	return tiles, nil
}

// TilesForGeometryWithEPSG projects the geometry from sourceEPSG into the TMS
// CRS and then computes tiles for the zoom range.
func (t *TileMatrixSet) TilesForGeometryWithEPSG(g orb.Geometry, sourceEPSG string, minZoom, maxZoom int, buffer float64) (grid.TilesList, error) {
	targetCRS, err := grid.ExtractCRS(t.TileMatrixSet)
	if err != nil {
		return nil, err
	}
	projected, err := grid.ProjectGeometry(g, sourceEPSG, targetCRS)
	if err != nil {
		return nil, err
	}
	return t.TilesForGeometry(projected, minZoom, maxZoom, buffer)
}

func parseZoom(id string) (int, error) {
	var z int
	_, err := fmt.Sscanf(id, "%d", &z)
	return z, err
}

// ZoomForID returns the zero-based zoom index for the given TileMatrix ID.
func (t *TileMatrixSet) ZoomForID(id string) (int, error) {
	if err := t.ensureInit(); err != nil {
		return 0, err
	}
	z, ok := t.idToZoom[id]
	if !ok {
		return 0, fmt.Errorf("tile matrix id %q not found", id)
	}
	return z, nil
}

// TileMatrixForID returns the TileMatrix for the given ID.
func (t *TileMatrixSet) TileMatrixForID(id string) (tms.TileMatrix, error) {
	mats, err := t.sortedMatrices()
	if err != nil {
		return tms.TileMatrix{}, err
	}
	z, err := t.ZoomForID(id)
	if err != nil {
		return tms.TileMatrix{}, err
	}
	if z < 0 || z >= len(mats) {
		return tms.TileMatrix{}, fmt.Errorf("zoom %d out of range for id %q", z, id)
	}
	return mats[z], nil
}
