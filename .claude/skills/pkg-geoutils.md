---
name: Package - geoutils (Geographic Utilities)
description: This skill should be used when the user asks to "calculate polygon area", "add geographic utilities", "compute area in hectares", or mentions anything about the pkg/geoutils/ package.
version: 1.0.0
---

# Package: geoutils — Geographic Utilities

The `geoutils` package provides geographic computation utilities, primarily polygon area calculation using the WGS84 ellipsoid model.

**Location:** `pkg/geoutils/`

---

## Functions

### CalculateAreaInHectares

Calculates the area of a GeoJSON polygon in hectares using the `orb` library with WGS84 geodesic calculations:

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
package geoutils

import (
    "github.com/paulmach/orb"
    "github.com/paulmach/orb/geo"
)

func CalculateAreaInHectares(coordinates [][][]float64) float64 {
    if len(coordinates) == 0 || len(coordinates[0]) == 0 {
        return 0
    }

    ring := make(orb.Ring, len(coordinates[0]))
    for i, coord := range coordinates[0] {
        ring[i] = orb.Point{coord[0], coord[1]}  // [lng, lat]
    }

    polygon := orb.Polygon{ring}
    areaInSquareMeters := geo.Area(polygon)

    return areaInSquareMeters / 10000  // Convert m² to hectares
}
```

---

## Key Rules

- **No internal imports** — standalone package
- **GeoJSON coordinate format** — `[longitude, latitude]` (not lat/lng)
- **WGS84 ellipsoid** — accurate geodesic area calculation
- **Returns hectares** — converts from square meters
- **Copy-ready** — depends on `github.com/paulmach/orb`

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
