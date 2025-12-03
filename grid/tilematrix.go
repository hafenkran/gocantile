package grid

import (
	"fmt"
	"math"

	"github.com/hafenkran/gocantile/tms"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/clip"
)

// TileMatrix wraps a TileMatrix from the OGC-generated structs and
// provides tile math for a single zoom level.
type TileMatrix struct {
	TM tms.TileMatrix
}

func (a TileMatrix) cellSize() float64 {
	return a.TM.CellSize
}

func (a TileMatrix) tileSizeX() float64 {
	return a.TM.TileWidth * a.TM.CellSize
}

func (a TileMatrix) tileSizeY() float64 {
	return a.TM.TileHeight * a.TM.CellSize
}

func (a TileMatrix) rowInfo() map[int]tileMatrixRowInfo {
	if len(a.TM.VariableMatrixWidths) == 0 {
		return nil
	}
	info := make(map[int]tileMatrixRowInfo)
	for _, v := range a.TM.VariableMatrixWidths {
		minRow := int(math.Round(v.MinTileRow))
		maxRow := int(math.Round(v.MaxTileRow))
		for r := minRow; r <= maxRow; r++ {
			info[r] = tileMatrixRowInfo{
				width:    a.matrixWidth() / int(math.Round(v.Coalesce)), // effective tiles after coalescing
				coalesce: int(math.Round(v.Coalesce)),
			}
		}
	}
	return info
}

// Resolution returns the ground resolution (units per pixel) for this matrix.
// In OGC TMS 2.0.0 this is equivalent to cellSize.
func (a TileMatrix) Resolution() float64 {
	return a.TM.CellSize
}

// TileSize returns the tile dimensions in ground units.
func (a TileMatrix) TileSize() (float64, float64) {
	return a.tileSizeX(), a.tileSizeY()
}

func (a TileMatrix) matrixWidth() int {
	return int(math.Round(a.TM.MatrixWidth))
}

func (a TileMatrix) matrixHeight() int {
	return int(math.Round(a.TM.MatrixHeight))
}

func (a TileMatrix) origin() (float64, float64, bool) {
	if len(a.TM.PointOfOrigin) < 2 {
		return 0, 0, false
	}
	return a.TM.PointOfOrigin[0], a.TM.PointOfOrigin[1], true
}

// TileForXY converts (x,y) in CRS coordinates (e.g., meters) into a tile index
// inside this TileMatrix.
func (a TileMatrix) TileForXY(x, y float64) (TileIndex, bool) {
	originX, originY, ok := a.origin()
	if !ok {
		return TileIndex{}, false
	}

	colFloat := (x - originX) / a.tileSizeX()
	var rowFloat float64
	if a.TM.CornerOfOrigin == tms.TileMatrixJsonCornerOfOriginBottomLeft {
		rowFloat = (y - originY) / a.tileSizeY()
	} else {
		rowFloat = (originY - y) / a.tileSizeY()
	}

	col := int(math.Floor(colFloat))
	row := int(math.Floor(rowFloat))

	if col < 0 || col >= a.matrixWidth() || row < 0 || row >= a.matrixHeight() {
		return TileIndex{}, false
	}

	if ri := a.rowInfo(); ri != nil {
		if info, ok := ri[row]; ok {
			if col >= info.width*info.coalesce {
				return TileIndex{}, false
			}
		}
	}

	return TileIndex{Col: col, Row: row}, true
}

// TileForLonLat projects lon/lat (degrees) into the matrix CRS via the provided
// projector and returns the tile index.
func (a TileMatrix) TileForLonLat(lon, lat float64, p Projector) (TileIndex, bool) {
	x, y, err := p.Forward(lon, lat)
	if err != nil {
		return TileIndex{}, false
	}
	return a.TileForXY(x, y)
}

func (a TileMatrix) BoundsForTile(t TileIndex) (Bounds, error) {
	originX, originY, ok := a.origin()
	if !ok {
		return Bounds{}, fmt.Errorf("tile matrix origin not defined")
	}

	if t.Col < 0 || t.Col >= a.matrixWidth() || t.Row < 0 || t.Row >= a.matrixHeight() {
		return Bounds{}, fmt.Errorf("tile out of range col=%d row=%d", t.Col, t.Row)
	}

	row := t.Row

	rowWidth := a.matrixWidth()
	coalesce := 1
	if ri := a.rowInfo(); ri != nil {
		if info, ok := ri[t.Row]; ok {
			rowWidth = info.width
			coalesce = info.coalesce
		}
	}
	if t.Col >= rowWidth {
		return Bounds{}, fmt.Errorf("tile out of range col=%d row=%d (row width %d)", t.Col, t.Row, rowWidth)
	}

	tileWidth := a.tileSizeX() * float64(coalesce)
	minX := originX + float64(t.Col)*tileWidth
	maxX := minX + tileWidth

	var minY, maxY float64
	if a.TM.CornerOfOrigin == tms.TileMatrixJsonCornerOfOriginBottomLeft {
		minY = originY + float64(row)*a.tileSizeY()
		maxY = minY + a.tileSizeY()
	} else {
		maxY = originY - float64(row)*a.tileSizeY()
		minY = maxY - a.tileSizeY()
	}

	return Bounds{MinX: minX, MinY: minY, MaxX: maxX, MaxY: maxY}, nil
}

// BoundsForTileLonLat returns the lon/lat bounds (degrees) of a tile by
// projecting its CRS bounds with the provided projector.
func (a TileMatrix) BoundsForTileLonLat(t TileIndex, p Projector) (Bounds, error) {
	b, err := a.BoundsForTile(t)
	if err != nil {
		return Bounds{}, err
	}
	minLon, minLat, err := p.Inverse(b.MinX, b.MinY)
	if err != nil {
		return Bounds{}, err
	}
	maxLon, maxLat, err := p.Inverse(b.MaxX, b.MaxY)
	if err != nil {
		return Bounds{}, err
	}
	return Bounds{MinX: minLon, MinY: minLat, MaxX: maxLon, MaxY: maxLat}, nil
}

// TileRangeForBounds returns the inclusive tile range covering the provided
// bounds in the matrix CRS. Returns false if the range is completely outside
// the matrix extent.
func (a TileMatrix) TileRangeForBounds(b Bounds) (TileRange, bool) {
	originX, originY, ok := a.origin()
	if !ok {
		return TileRange{}, false
	}

	tileSizeX := a.tileSizeX()
	tileSizeY := a.tileSizeY()

	minCol := int(math.Floor((b.MinX - originX) / tileSizeX))
	maxCol := int(math.Ceil((b.MaxX-originX)/tileSizeX)) - 1

	var minRow, maxRow int
	if a.TM.CornerOfOrigin == tms.TileMatrixJsonCornerOfOriginBottomLeft {
		minRow = int(math.Floor((b.MinY - originY) / tileSizeY))
		maxRow = int(math.Ceil((b.MaxY-originY)/tileSizeY)) - 1
	} else {
		minRow = int(math.Floor((originY - b.MaxY) / tileSizeY))
		maxRow = int(math.Ceil((originY-b.MinY)/tileSizeY)) - 1
	}

	// Clip to matrix bounds.
	if maxCol < 0 || maxRow < 0 || minCol >= a.matrixWidth() || minRow >= a.matrixHeight() {
		return TileRange{}, false
	}
	if minCol < 0 {
		minCol = 0
	}
	if minRow < 0 {
		minRow = 0
	}
	if maxCol >= a.matrixWidth() {
		maxCol = a.matrixWidth() - 1
	}
	if maxRow >= a.matrixHeight() {
		maxRow = a.matrixHeight() - 1
	}

	if minCol > maxCol || minRow > maxRow {
		return TileRange{}, false
	}

	return TileRange{MinCol: minCol, MaxCol: maxCol, MinRow: minRow, MaxRow: maxRow}, true
}

// TilesForBounds returns all tiles covering the given bounds (in CRS units),
// clipped to the matrix extent.
func (a TileMatrix) TilesForBounds(b Bounds) []TileIndex {
	tr, ok := a.TileRangeForBounds(b)
	if !ok {
		return nil
	}
	var tiles []TileIndex
	for r := tr.MinRow; r <= tr.MaxRow; r++ {
		for c := tr.MinCol; c <= tr.MaxCol; c++ {
			tiles = append(tiles, TileIndex{Col: c, Row: r})
		}
	}
	return tiles
}

// TilesForGeometry returns tiles intersecting the geometry's bounding box,
// clipped to the matrix extent. Intersection is bounding-box based; for precise
// clipping you can post-filter with your own geometry tests.
func (a TileMatrix) TilesForGeometry(g orb.Geometry) []TileIndex {
	bound := g.Bound()
	b := Bounds{
		MinX: bound.Min[0],
		MinY: bound.Min[1],
		MaxX: bound.Max[0],
		MaxY: bound.Max[1],
	}
	tr, ok := a.TileRangeForBounds(b)
	if !ok {
		return nil
	}
	geomBound := orb.Bound{
		Min: orb.Point{bound.Min[0], bound.Min[1]},
		Max: orb.Point{bound.Max[0], bound.Max[1]},
	}
	var tiles []TileIndex
	for r := tr.MinRow; r <= tr.MaxRow; r++ {
		for c := tr.MinCol; c <= tr.MaxCol; c++ {
			tb, _ := a.BoundsForTile(TileIndex{Col: c, Row: r})
			tileBound := orb.Bound{
				Min: orb.Point{tb.MinX, tb.MinY},
				Max: orb.Point{tb.MaxX, tb.MaxY},
			}
			clipped := clip.Geometry(tileBound, g)
			if clipped != nil || tileBound.Intersects(geomBound) {
				tiles = append(tiles, TileIndex{Col: c, Row: r})
			}
		}
	}
	return tiles
}
