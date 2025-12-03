package tms

// Friendly aliases for generated types to provide stable names without the
// "Json" suffix in downstream code.
type (
	TileMatrixSet    = TileMatrixSetJson
	TileMatrix       = TileMatrixJson
	TileMatrixLimits = TileMatrixLimitsJson
	A2DPoint         = A2DPointJson
	A2DBoundingBox   = A2DBoundingBoxJson
)
