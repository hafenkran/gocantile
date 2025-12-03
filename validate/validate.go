package validate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path"
	"sync"

	"github.com/hafenkran/gocantile"
	"github.com/hafenkran/gocantile/tms"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

var (
	compileOnce sync.Once
	compileErr  error
	schemaCache map[string]*jsonschema.Schema
)

func compileSchemas() error {
	compileOnce.Do(func() {
		compiler := jsonschema.NewCompiler()
		base := "embed://schema/"

		entries, err := gocantile.SchemaFS.ReadDir("data/schema")
		if err != nil {
			compileErr = fmt.Errorf("read embedded schema dir: %w", err)
			return
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			raw, err := gocantile.SchemaFS.ReadFile(path.Join("data/schema", name))
			if err != nil {
				compileErr = fmt.Errorf("read schema %s: %w", name, err)
				return
			}
			if err := compiler.AddResource(base+name, bytes.NewReader(raw)); err != nil {
				compileErr = fmt.Errorf("add resource %s: %w", name, err)
				return
			}
		}

		schemaCache = make(map[string]*jsonschema.Schema)
		for _, name := range []string{"tileMatrixSet.json", "tileSet.json"} {
			s, err := compiler.Compile(base + name)
			if err != nil {
				compileErr = fmt.Errorf("compile schema %s: %w", name, err)
				return
			}
			schemaCache[name] = s
		}
	})
	return compileErr
}

func validateWithSchema(schemaName string, data []byte) error {
	if err := compileSchemas(); err != nil {
		return err
	}
	schema, ok := schemaCache[schemaName]
	if !ok {
		return fmt.Errorf("schema %q not loaded", schemaName)
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if err := schema.Validate(v); err != nil {
		return err
	}
	return nil
}

// ValidateTileMatrixSetJSON validates raw TileMatrixSet JSON against the embedded schema.
func ValidateTileMatrixSetJSON(data []byte) error {
	return validateWithSchema("tileMatrixSet.json", data)
}

// ValidateTileMatrixSet marshals the struct and validates it against the schema.
func ValidateTileMatrixSet(set tms.TileMatrixSet) error {
	raw, err := json.Marshal(set)
	if err != nil {
		return fmt.Errorf("marshal TileMatrixSet: %w", err)
	}
	return ValidateTileMatrixSetJSON(raw)
}

// ValidateTileSetJSON validates raw TileSet JSON against the embedded schema.
func ValidateTileSetJSON(data []byte) error {
	return validateWithSchema("tileSet.json", data)
}

// ValidateTileSet marshals any struct/map and validates it against the TileSet schema.
func ValidateTileSet(v interface{}) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal TileSet: %w", err)
	}
	return ValidateTileSetJSON(raw)
}
