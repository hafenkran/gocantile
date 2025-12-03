package gocantile

import "embed"

// SchemaFS embeds the OGC TileMatrixSet schema files used for validation.
//
//go:embed data/schema/*.json
var SchemaFS embed.FS
