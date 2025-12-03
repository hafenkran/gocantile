package grid

import "github.com/paulmach/orb"

// TileIndex identifies a single tile.
type TileIndex struct {
	Col int
	Row int
}

type Bounds struct {
	MinX float64
	MinY float64
	MaxX float64
	MaxY float64
}

// TileRange represents an inclusive range of tiles.
type TileRange struct {
	MinCol int
	MaxCol int
	MinRow int
	MaxRow int
}

// Tile identifies a tile at a given zoom level.
type Tile struct {
	Zoom int
	TileIndex
}

// tileMatrixRowInfo describes per-row matrix width and coalesce factor for
// variableMatrixWidths support.
type tileMatrixRowInfo struct {
	width    int
	coalesce int
}

// TilesList is a helper to accumulate tiles across zoom levels.
type TilesList []Tile

// Geometry is the minimal interface needed for tiling.
type Geometry interface {
	Bound() orb.Bound
}
