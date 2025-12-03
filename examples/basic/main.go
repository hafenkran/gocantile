package main

import (
	"fmt"
	"log"

	"github.com/hafenkran/gocantile"
)

func main() {
	tms, err := gocantile.LoadTileMatrixSet("WebMercatorQuad")
	if err != nil {
		log.Fatalf("load TMS: %v", err)
	}

	// Find tile for lon/lat at zoom 10.
	tile, ok, err := tms.TileForLonLat(13.4050, 52.5200, 10, nil)
	if err != nil || !ok {
		log.Fatalf("tile for lon/lat: %v ok=%v", err, ok)
	}

	xyBounds, err := tms.XYBounds(tile)
	if err != nil {
		log.Fatalf("xy bounds: %v", err)
	}
	llBounds, err := tms.Bounds(tile, nil)
	if err != nil {
		log.Fatalf("lon/lat bounds: %v", err)
	}

	fmt.Printf("Tile z=%d x=%d y=%d\n", tile.Zoom, tile.Col, tile.Row)
	fmt.Printf("XY bounds: %+v\n", xyBounds)
	fmt.Printf("Lon/Lat bounds: %+v\n", llBounds)
}
