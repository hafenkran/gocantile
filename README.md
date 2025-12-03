gocantile
=========

[![CI](https://github.com/hafenkran/gocantile/actions/workflows/ci.yml/badge.svg)](https://github.com/hafenkran/gocantile/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hafenkran/gocantile)](https://goreportcard.com/report/github.com/hafenkran/gocantile)
[![codecov](https://codecov.io/gh/hafenkran/gocantile/branch/main/graph/badge.svg)](https://codecov.io/gh/hafenkran/gocantile)
[![Go Reference](https://pkg.go.dev/badge/github.com/hafenkran/gocantile.svg)](https://pkg.go.dev/github.com/hafenkran/gocantile)

`gocantile` is a simple Go library for working with **OGC TileMatrixSet 2.0.0** ([spec](https://docs.ogc.org/is/17-083r4/17-083r4.html)), inspired by the excellent Python library [Morecantile](https://github.com/developmentseed/morecantile).

It provides a lightweight, type-safe interface for loading TMS definitions, converting Lon/Lat to tiles, computing bounds, and performing multi-zoom coverage analysis. Optional PROJ integration enables on-the-fly coordinate projection.

Features
--------

- OGC-compliant TileMatrixSet 2.0.0 models (schema-generated)
- Ready to use embedded schema registry (e.g. WebMercatorQuad, WorldCRS84Quad, WGS1984Quad)
- Tile calculations per zoom level; XY/LonLat → tile index, bounds, origin handling, variable matrix width
- Geometry coverage across zoom ranges with optional CRS reprojection via PROJ
- Validation utilities for TileMatrixSet / TileSet JSON schemas

Install
-------

Install like any standard Go module:

```sh
go get github.com/hafenkran/gocantile
```

`gocantile` requires **Go 1.23+** (see `go.mod`). PROJ is only needed if you plan to use reprojection features (e.g., `TilesForGeometryWithEPSG`, `Bounds(..., projector)`, etc.). Otherwise the library works without system dependencies.

Typical PROJ 8+ installs (Optional):

```sh
// macOS
brew install proj

// Debian/Ubuntu
apt-get install proj-bin

// Arch
pacman -S proj
```

Quick start
-----------

Tile lookup and bounds (WebMercatorQuad):

```go
// Load built-in TMS definition (WebMercatorQuad)
tms, err := gocantile.LoadTileMatrixSet("WebMercatorQuad")
if err != nil { panic(err) }

// Convert lon/lat → tile (z=10)
tile, ok, err := tms.TileForLonLat(13.4050, 52.5200, 10, nil)
if err != nil || !ok { panic("tile lookup failed") }

// Bounds in WebMercator meters
xyBounds, err := tms.XYBounds(tile)
if err != nil { panic(err) }

// Geographic bounds in lon/lat (auto-reprojected)
llBounds, err := tms.Bounds(tile, nil)
if err != nil { panic(err) }

fmt.Printf("Tile %d/%d/%d\n", tile.Zoom, tile.TileIndex.Col, tile.TileIndex.Row)
fmt.Printf("XY bounds: %+v\nLonLat bounds: %+v\n", xyBounds, llBounds)
```

Geometry coverage across zooms with auto-projection:

```go
// Load built-in TMS definition (WebMercatorQuad)
tms, err := gocantile.LoadTileMatrixSet("WebMercatorQuad")
if err != nil { panic(err) }

// Polygon in WGS84
poly := orb.Polygon{{
    {13.0, 52.4}, {13.8, 52.4},
    {13.8, 52.7}, {13.0, 52.7},
    {13.0, 52.4},
}}

// Compute all tiles covering the geometry from zoom 10 to 12
// Reprojection happens automatically based on supplied EPSG
tiles, err := tms.TilesForGeometryWithEPSG(poly, "EPSG:4326", 10, 12, 0)
if err != nil { panic(err) }

fmt.Println("Tile count (z10–12):", len(tiles))
```

List embedded TileMatrixSets:

```go
names := gocantile.AvailableTileMatrixSets()
for _, name := range names {
    tms, err := gocantile.LoadTileMatrixSet(name)
    if err != nil { panic(err) }
    fmt.Printf("%s (minzoom=%d, maxzoom=%d)\n", name, tms.MinZoom(), tms.MaxZoom())
}

// CanadianNAD83_LCC (minzoom=0, maxzoom=14)
// CDB1GlobalGrid (minzoom=0, maxzoom=22)
// EuropeanETRS89_LAEAQuad (minzoom=0, maxzoom=19)
// GNOSISGlobalGrid (minzoom=0, maxzoom=22)
// UPSAntarcticWGS84Quad (minzoom=0, maxzoom=20)
// UPSArcticWGS84Quad (minzoom=0, maxzoom=20)
// UTM31WGS84Quad (minzoom=0, maxzoom=20)
// WebMercatorQuad (minzoom=0, maxzoom=24)
// WGS1984Quad (minzoom=0, maxzoom=18)
// WorldCRS84Quad (minzoom=0, maxzoom=22)
// WorldMercatorWGS84Quad (minzoom=0, maxzoom=21)
```

More examples are under `examples/`.

Development
-----------

Go 1.23+ (see `go.mod`) is used for development. CI on pushes to `main` and all pull requests runs formatting, vetting, linting, and tests (see `.github/workflows/ci.yml`).

Run the [golangci-lint](https://golangci-lint.run/) locally before pushing:

```sh
golangci-lint run -verbose
```

To run the tests:

```sh
go test ./...
```

Licenses
--------

This repository is licensed under MIT ([license](LICENSE)).

It uses the following open source dependencies:

- go-proj v8 ([repo](https://github.com/everystreet/go-proj), [license](https://github.com/everystreet/go-proj/blob/v8.0.0/LICENSE)) – Apache-2.0
- jsonschema/v5 ([repo](https://github.com/santhosh-tekuri/jsonschema), [license](https://github.com/santhosh-tekuri/jsonschema/blob/v5.3.1/LICENSE)) – Apache-2.0
- orb ([repo](https://github.com/paulmach/orb), [license](https://github.com/paulmach/orb/blob/v0.12.0/LICENSE.md)) – MIT
- golang/geo ([repo](https://github.com/golang/geo), [license](https://github.com/golang/geo/blob/740aa86cb551/LICENSE)) – Apache-2.0

Embedded data:

- OGC TileMatrixSet 2.0.0 schemas and sample TileMatrixSet/TileSet JSON (`data/schema`, `data/tilematrixset`, `data/tileset`) sourced from [schemas.opengis.net](https://schemas.opengis.net/tms/2.0/json/) (OGC; see source for terms/attribution).
