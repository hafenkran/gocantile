package main

import (
	"fmt"
	"log"

	"github.com/hafenkran/gocantile"
	"github.com/paulmach/orb"
)

func main() {
	tms, err := gocantile.LoadTileMatrixSet("WebMercatorQuad")
	if err != nil {
		log.Fatalf("load TMS: %v", err)
	}

	// Simple polygon in lon/lat.
	poly := orb.Polygon{{
		{13.088626854245092, 52.416237574678775},
		{13.765208845079911, 52.438683015289115},
		{13.620002424454327, 52.471381583541444},
		{13.476885923873368, 52.66708748539642},
		{13.088626854245092, 52.416237574678775},
	}}

	// Auto-project lon/lat into the TMS CRS (EPSG:3857).
	tiles, err := tms.TilesForGeometryWithEPSG(poly, "EPSG:4326", 10, 12, 0)
	if err != nil {
		log.Fatalf("tiles for geometry: %v", err)
	}
	fmt.Printf("Tiles covering polygon (zoom 10-12): %d\n", len(tiles))
	for i, ti := range tiles {
		if i >= 5 {
			fmt.Printf("... and %d more\n", len(tiles)-5)
			break
		}
		fmt.Printf("z=%d x=%d y=%d\n", ti.Zoom, ti.Col, ti.Row)
	}
}
