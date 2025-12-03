// Package gocantile exposes lightweight helpers around OGC TileMatrixSet data
// and tile math utilities.
package gocantile

import (
	"github.com/hafenkran/gocantile/grid"
	"github.com/hafenkran/gocantile/tms"
)

// Re-export data model types (except TileMatrixSet, which is wrapped with helpers).
type (
	TileMatrixLimits = tms.TileMatrixLimits
	A2DPoint         = tms.A2DPoint
	A2DBoundingBox   = tms.A2DBoundingBox
)

// Re-export grid adapter and helpers.
type (
	TileMatrix    = grid.TileMatrix
	TileIndex     = grid.TileIndex
	Tile          = grid.Tile
	Bounds        = grid.Bounds
	TileRange     = grid.TileRange
	Projector     = grid.Projector
	ProjProjector = grid.ProjProjector
)

// NewProjProjector creates a PROJ-backed projector from source CRS to target CRS.
func NewProjProjector(sourceCRS, targetCRS string) grid.ProjProjector {
	return grid.NewProjProjector(sourceCRS, targetCRS)
}

// NewWGS84Projector convenience for EPSG:4326 -> target CRS.
func NewWGS84Projector(targetCRS string) grid.ProjProjector {
	return grid.NewWGS84Projector(targetCRS)
}

// ProjectorFromTMS builds a projector from a TileMatrixSet CRS (target) with EPSG:4326 as source.
func ProjectorFromTMS(set *TileMatrixSet) (grid.ProjProjector, error) {
	return grid.ProjectorFromTMS(set.TileMatrixSet)
}
