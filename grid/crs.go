package grid

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hafenkran/gocantile/tms"
)

// ExtractCRS returns a PROJ-compatible CRS string from the TileMatrixSet crs
// field. It supports string EPSG codes and minimal object forms with "uri".
func ExtractCRS(set tms.TileMatrixSet) (string, error) {
	if set.Crs == nil {
		return "", fmt.Errorf("crs is nil")
	}
	switch v := set.Crs.(type) {
	case string:
		return normalizeCRSString(v), nil
	case map[string]interface{}:
		if uri, ok := v["uri"].(string); ok && uri != "" {
			return normalizeCRSString(uri), nil
		}
		if wktVal, ok := v["wkt"]; ok {
			// For WKT/ProjJSON payloads, pass through as string if possible.
			if s, ok := wktVal.(string); ok {
				return s, nil
			}
			if m, ok := wktVal.(map[string]interface{}); ok {
				raw, err := json.Marshal(m)
				if err != nil {
					return "", fmt.Errorf("marshal projjson: %w", err)
				}
				return string(raw), nil
			}
			return "", fmt.Errorf("unsupported wkt type %T", wktVal)
		}
		if _, ok := v["referenceSystem"]; ok {
			return "", fmt.Errorf("referenceSystem parsing not supported")
		}
	}
	return "", fmt.Errorf("unsupported crs type %T", set.Crs)
}

func normalizeCRSString(s string) string {
	upper := strings.ToUpper(s)
	// Accept EPSG:xxxx or URIs containing /EPSG/0/xxxx
	if strings.HasPrefix(upper, "EPSG:") {
		return upper
	}
	lower := strings.ToLower(s)
	if strings.Contains(lower, "/epsg/0/") {
		parts := strings.Split(s, "/")
		if len(parts) > 0 {
			code := parts[len(parts)-1]
			return "EPSG:" + code
		}
	}
	if strings.HasPrefix(lower, "urn:ogc:def:crs:epsg::") {
		code := strings.TrimPrefix(lower, "urn:ogc:def:crs:epsg::")
		return "EPSG:" + code
	}
	return s
}
