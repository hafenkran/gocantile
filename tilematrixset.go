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
}

func WrapTileMatrixSet(set tms.TileMatrixSet) *TileMatrixSet {
	return &TileMatrixSet{TileMatrixSet: set}
}

func (t *TileMatrixSet) sortedMatrices() []tms.TileMatrix {
	t.once.Do(func() {
		mats := make([]tms.TileMatrix, len(t.TileMatrices))
		copy(mats, t.TileMatrices)
		sort.SliceStable(mats, func(i, j int) bool {
			zi, errI := parseZoom(mats[i].Id)
			zj, errJ := parseZoom(mats[j].Id)
			if errI == nil && errJ == nil {
				return zi < zj
			}
			return i < j
		})
		t.matrices = mats
	})
	return t.matrices
}

func (t *TileMatrixSet) MinZoom() int {
	if len(t.TileMatrices) == 0 {
		return 0
	}
	return 0
}

func (t *TileMatrixSet) MaxZoom() int {
	mats := t.sortedMatrices()
	if len(mats) == 0 {
		return 0
	}
	return len(mats) - 1
}

func (t *TileMatrixSet) ResolutionForZoom(z int) (float64, error) {
	mats := t.sortedMatrices()
	if z < 0 || z >= len(mats) {
		return 0, fmt.Errorf("zoom %d out of range", z)
	}
	return mats[z].CellSize, nil
}

func (t *TileMatrixSet) ZoomForResolution(res, tol float64) (int, error) {
	mats := t.sortedMatrices()
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
	mats := t.sortedMatrices()
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
	mats := t.sortedMatrices()
	if tile.Zoom < 0 || tile.Zoom >= len(mats) {
		return grid.Bounds{}, fmt.Errorf("zoom %d out of range", tile.Zoom)
	}
	adapter := grid.TileMatrix{TM: mats[tile.Zoom]}
	return adapter.BoundsForTile(tile.TileIndex)
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
	mats := t.sortedMatrices()
	if tile.Zoom < 0 || tile.Zoom >= len(mats) {
		return grid.Bounds{}, fmt.Errorf("zoom %d out of range", tile.Zoom)
	}
	adapter := grid.TileMatrix{TM: mats[tile.Zoom]}
	return adapter.BoundsForTileLonLat(tile.TileIndex, p)
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
	mats := t.sortedMatrices()
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

// TilesForGeometry returns tiles covering the geometry across zoom levels
// [minZoom, maxZoom] inclusive. Optional buffer expands the geometry bounds
// before tiling (in CRS units).
func (t *TileMatrixSet) TilesForGeometry(g orb.Geometry, minZoom, maxZoom int, buffer float64) (grid.TilesList, error) {
	if minZoom < 0 || maxZoom < minZoom {
		return nil, fmt.Errorf("invalid zoom range min=%d max=%d", minZoom, maxZoom)
	}
	mats := t.sortedMatrices()
	if maxZoom >= len(mats) {
		return nil, fmt.Errorf("max zoom %d out of range", maxZoom)
	}
	bound := g.Bound()
	if buffer != 0 {
		bound.Min[0] -= buffer
		bound.Min[1] -= buffer
		bound.Max[0] += buffer
		bound.Max[1] += buffer
		// Replace geometry with buffered bbox polygon for precise clipping at tile level.
		g = orb.Polygon{{
			{bound.Min[0], bound.Min[1]},
			{bound.Max[0], bound.Min[1]},
			{bound.Max[0], bound.Max[1]},
			{bound.Min[0], bound.Max[1]},
			{bound.Min[0], bound.Min[1]},
		}}
	}

	var tiles grid.TilesList
	for z := minZoom; z <= maxZoom; z++ {
		adapter := grid.TileMatrix{TM: mats[z]}
		rangeTiles := adapter.TilesForGeometry(g)
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
