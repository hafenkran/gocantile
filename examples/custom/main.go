package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hafenkran/gocantile"
	"github.com/hafenkran/gocantile/tms"
)

func main() {
	// Custom single-level TMS with bottom-left origin and variable matrix widths.
	custom := tms.TileMatrixSet{
		Crs: "EPSG:3857",
		TileMatrices: []tms.TileMatrix{
			{
				Id:             "0",
				CellSize:       1000,
				TileWidth:      1,
				TileHeight:     1,
				MatrixWidth:    4,
				MatrixHeight:   4,
				PointOfOrigin:  []float64{0, 0}, // bottom-left
				CornerOfOrigin: tms.TileMatrixJsonCornerOfOriginBottomLeft,
				VariableMatrixWidths: []tms.VariableMatrixWidthJson{
					{Coalesce: 2, MinTileRow: 0, MaxTileRow: 1}, // top rows coalesce
				},
			},
		},
	}

	// Pretty-print the custom TMS for reference.
	raw, _ := json.MarshalIndent(custom, "", "  ")
	fmt.Println("Custom TMS:")
	fmt.Println(string(raw))

	adapter := gocantile.TileMatrix{TM: custom.TileMatrices[0]}

	// Tile at bottom-left origin: expect (0,0) at 500,500.
	tile, ok := adapter.TileForXY(500, 500)
	if !ok {
		log.Fatalf("expected tile")
	}
	fmt.Printf("Tile for XY (500,500): %+v\n", tile)

	// Bounds for a tile in coalesced row (row 0): width doubles.
	bounds, err := adapter.BoundsForTile(gocantile.TileIndex{Col: 1, Row: 0})
	if err != nil {
		log.Fatalf("bounds err: %v", err)
	}
	fmt.Printf("Bounds for tile (1,0): %+v\n", bounds)
}
