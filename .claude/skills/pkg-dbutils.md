---
name: Package - dbutils (Database Utilities)
description: This skill should be used when the user asks to "create database utilities", "add LIKE query support", "handle database dialect differences", "add PostgreSQL/SQLite compatibility", "create geometry queries", "add PostGIS support", or mentions anything about the pkg/dbutils/ package.
version: 1.0.0
---

# Package: dbutils — Database Dialect Utilities

The `dbutils` package provides database dialect-specific query helpers that abstract differences between PostgreSQL and SQLite (used in BDD tests). This includes text queries (LIKE), geometry/PostGIS operations, and any other dialect-specific SQL construction. Everything lives in a single `pkg/dbutils/` folder.

**Location:** `pkg/dbutils/`

---

## Architecture

```
pkg/dbutils/
├── query.go              ← Query interface (LIKE operations)
├── query_factory.go      ← Factory that creates dialect-specific implementation
├── postgres.go           ← PostgreSQL implementation
├── sqlite.go             ← SQLite implementation
├── geometry.go           ← SELECT clause generator for geometry (ST_AsGeoJSON)
└── geometry_insert.go    ← INSERT/UPDATE clause generator for geometry (ST_GeomFromGeoJSON)
```

**Note:** All database query utilities live in this ONE package. Do not create separate packages for related concerns (e.g., don't create a separate `pkg/queryutils/` — that's duplication).

---

## Text Query Interface

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
type Query interface {
    Like(column, value string) string
    LikeCaseSensitive(column, value string) string
}
```

### Factory

```go
func NewQuery(dialectName string) Query {
    switch dialectName {
    case "postgres":
        return &postgresQuery{}
    case "sqlite":
        return &sqliteQuery{}
    default:
        return &postgresQuery{}
    }
}
```

### PostgreSQL Implementation
```go
type postgresQuery struct{}

func (p *postgresQuery) Like(column, value string) string {
    return fmt.Sprintf("%s ILIKE '%%%s%%'", column, value)  // Case-insensitive
}

func (p *postgresQuery) LikeCaseSensitive(column, value string) string {
    return fmt.Sprintf("%s LIKE '%%%s%%'", column, value)
}
```

### SQLite Implementation
```go
type sqliteQuery struct{}

func (s *sqliteQuery) Like(column, value string) string {
    return fmt.Sprintf("%s LIKE '%%%s%%'", column, value)  // Case-insensitive by default
}

func (s *sqliteQuery) LikeCaseSensitive(column, value string) string {
    return fmt.Sprintf("%s GLOB '*%s*'", column, value)  // GLOB is case-sensitive
}
```

---

## Geometry Query Helpers

### SELECT Clause (reading geometry as GeoJSON)

```go
func GeometrySelectClause(dialectName string, columns ...string) string {
    if dialectName == "sqlite" {
        return strings.Join(columns, ", ")
    }
    // PostgreSQL: ST_AsGeoJSON(column) as column
    clauses := make([]string, len(columns))
    for i, col := range columns {
        clauses[i] = fmt.Sprintf("ST_AsGeoJSON(%s) as %s", col, col)
    }
    return strings.Join(clauses, ", ")
}
```

### INSERT/UPDATE Expression (writing GeoJSON to geometry)

```go
func GeometryInsertExpression(dialectName, geoJson string) string {
    if dialectName == "sqlite" {
        return fmt.Sprintf("'%s'", geoJson)
    }
    return fmt.Sprintf("ST_GeomFromGeoJSON('%s')", geoJson)
}
```

---

## Key Rules

- **No internal imports** — standalone package
- **All dialect utilities in ONE package** — don't split into separate packages
- **Dialect-aware** — handles PostgreSQL (production) vs SQLite (BDD tests) differences
- **Factory pattern** — create the right implementation based on dialect name
- **Copy-ready** — no external dependencies beyond stdlib

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
