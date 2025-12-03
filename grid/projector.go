package grid

import (
	"fmt"

	"github.com/everystreet/go-proj/v8/proj"
	"github.com/hafenkran/gocantile/tms"
	"github.com/paulmach/orb"
)

// Projector converts between lon/lat (degrees) and projected CRS coordinates.
type Projector interface {
	Forward(lonDeg, latDeg float64) (x, y float64, err error)
	Inverse(x, y float64) (lonDeg, latDeg float64, err error)
}

// ProjProjector uses PROJ to transform between a source CRS (lon/lat in
// degrees if source is EPSG:4326) and a target CRS (e.g., "EPSG:3857").
type ProjProjector struct {
	SourceCRS string
	TargetCRS string
}

// NewProjProjector creates a PROJ-backed projector from source CRS to target
// CRS.
func NewProjProjector(sourceCRS, targetCRS string) ProjProjector {
	return ProjProjector{
		SourceCRS: sourceCRS,
		TargetCRS: targetCRS,
	}
}

// NewWGS84Projector convenience for EPSG:4326 -> target CRS.
func NewWGS84Projector(targetCRS string) ProjProjector {
	return NewProjProjector("EPSG:4326", targetCRS)
}

// ProjectorFromTMS builds a projector using the TileMatrixSet CRS as target and
// EPSG:4326 as source.
func ProjectorFromTMS(set tms.TileMatrixSet) (ProjProjector, error) {
	crs, err := ExtractCRS(set)
	if err != nil {
		return ProjProjector{}, err
	}
	return NewWGS84Projector(crs), nil
}

// Forward projects lon/lat (degrees) to target CRS coordinates.
func (p ProjProjector) Forward(lonDeg, latDeg float64) (float64, float64, error) {
	coord := proj.XY{X: lonDeg, Y: latDeg}
	if err := proj.CRSToCRS(p.SourceCRS, p.TargetCRS, func(pr proj.Projection) {
		proj.TransformForward(pr, &coord)
	}); err != nil {
		return 0, 0, fmt.Errorf("proj forward transform: %w", err)
	}
	return coord.X, coord.Y, nil
}

// Inverse projects target CRS coordinates to lon/lat (degrees).
func (p ProjProjector) Inverse(x, y float64) (float64, float64, error) {
	coord := proj.XY{X: x, Y: y}
	if err := proj.CRSToCRS(p.TargetCRS, p.SourceCRS, func(pr proj.Projection) {
		proj.TransformForward(pr, &coord)
	}); err != nil {
		return 0, 0, fmt.Errorf("proj inverse transform: %w", err)
	}
	return coord.X, coord.Y, nil
}

// ProjectGeometry projects an orb.Geometry from the source CRS to the target CRS
// using PROJ. If source and target are equal, the geometry is returned as-is.
func ProjectGeometry(g orb.Geometry, sourceCRS, targetCRS string) (orb.Geometry, error) {
	if sourceCRS == targetCRS {
		return g, nil
	}
	p := NewProjProjector(sourceCRS, targetCRS)

	switch geom := g.(type) {
	case orb.Point:
		x, y, err := p.Forward(geom[0], geom[1])
		if err != nil {
			return nil, err
		}
		return orb.Point{x, y}, nil
	case orb.Ring:
		out := make(orb.Ring, 0, len(geom))
		for _, pt := range geom {
			x, y, err := p.Forward(pt[0], pt[1])
			if err != nil {
				return nil, err
			}
			out = append(out, orb.Point{x, y})
		}
		return out, nil
	case orb.LineString:
		out := make(orb.LineString, 0, len(geom))
		for _, pt := range geom {
			x, y, err := p.Forward(pt[0], pt[1])
			if err != nil {
				return nil, err
			}
			out = append(out, orb.Point{x, y})
		}
		return out, nil
	case orb.MultiLineString:
		out := make(orb.MultiLineString, 0, len(geom))
		for _, ls := range geom {
			pj, err := ProjectGeometry(ls, sourceCRS, targetCRS)
			if err != nil {
				return nil, err
			}
			out = append(out, pj.(orb.LineString))
		}
		return out, nil
	case orb.Polygon:
		out := make(orb.Polygon, 0, len(geom))
		for _, ring := range geom {
			pj, err := ProjectGeometry(ring, sourceCRS, targetCRS)
			if err != nil {
				return nil, err
			}
			out = append(out, pj.(orb.Ring))
		}
		return out, nil
	case orb.MultiPolygon:
		out := make(orb.MultiPolygon, 0, len(geom))
		for _, poly := range geom {
			pj, err := ProjectGeometry(poly, sourceCRS, targetCRS)
			if err != nil {
				return nil, err
			}
			out = append(out, pj.(orb.Polygon))
		}
		return out, nil
	case orb.Collection:
		out := make(orb.Collection, 0, len(geom))
		for _, sub := range geom {
			pj, err := ProjectGeometry(sub, sourceCRS, targetCRS)
			if err != nil {
				return nil, err
			}
			out = append(out, pj)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported geometry type %T", g)
	}
}
