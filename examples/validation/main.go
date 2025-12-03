package main

import (
	"fmt"
	"log"

	"github.com/hafenkran/gocantile"
	"github.com/hafenkran/gocantile/validate"
)

func main() {
	// Load embedded TMS.
	tms, err := gocantile.LoadTileMatrixSet("WebMercatorQuad")
	if err != nil {
		log.Fatalf("load TMS: %v", err)
	}

	// Validate against schema.
	if err := validate.ValidateTileMatrixSet(tms.TileMatrixSet); err != nil {
		log.Fatalf("TMS schema validation failed: %v", err)
	}

	// Extract CRS.
	crs, err := gocantile.ProjectorFromTMS(tms)
	if err != nil {
		log.Fatalf("extract CRS: %v", err)
	}

	fmt.Println("TMS:", "WebMercatorQuad")
	fmt.Println("CRS:", crs.TargetCRS)
	fmt.Println("MinZoom:", tms.MinZoom(), "MaxZoom:", tms.MaxZoom())
	if bbox, err := tms.XYBBox(); err == nil {
		fmt.Printf("XY BBox: %+v\n", bbox)
	}
}
