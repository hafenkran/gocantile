package tms

// Code generation for OGC TileMatrixSet schemas using go-jsonschema.
// Run `go generate ./tms` to refresh the generated types.

//go:generate go run github.com/atombender/go-jsonschema@latest -p tms -o tilematrixset_gen.go ../data/schema/tileMatrixSet.json
//go:generate go run github.com/atombender/go-jsonschema@latest -p tms -o tilematrixlimits_gen.go ../data/schema/tileMatrixLimits.json
