package main

import (
	"fmt"
	"log"

	"github.com/hafenkran/gocantile"
)

func main() {
	names := gocantile.AvailableTileMatrixSets()
	if len(names) == 0 {
		log.Fatalf("no embedded TileMatrixSets found")
	}
	fmt.Println("Embedded TileMatrixSets:")
	for _, name := range names {
		tms, err := gocantile.LoadTileMatrixSet(name)
		if err != nil {
			log.Fatalf("load %s: %v", name, err)
		}
		fmt.Printf("- %s (minzoom=%d, maxzoom=%d)\n", name, tms.MinZoom(), tms.MaxZoom())
	}
}
