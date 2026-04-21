# AGENTS.md - GeoNet Development Guide

This file is intended for agentic coding agents working in this repository.

## Project Overview

GeoNet is a Go library for generating a network from multiple GPS tracks. It processes GPX files,
merges overlapping tracks, and produces simplified network representations.

- **Go version**: 1.21
- **Module**: `mnezerka/geonet`
- **Main packages**: `s2store` (S2 spatial index), `store` (core), `mongostore` (MongoDB), `tracks`, `cmd`

## Build, Lint, and Test Commands

### Build
```bash
go build -v
```

### Run all tests
```bash
go test -v ./...
```

### Run tests for a specific package
```bash
go test -v ./s2store
```

### Run a single test function
```bash
go test -v ./s2store -run TestSimplifySingleTrack
```

### Run tests with verbose output and coverage
```bash
go test -v -cover ./...
```

### Build and run the CLI
```bash
go run main.go --help
```

## Code Style Guidelines

### General
- Uses **tabs** for indentation (enforced by gofmt)
- Max line length: not strictly enforced, but aim for readability
- No trailing whitespace
- File encoding: UTF-8

### Naming Conventions
- **Packages**: lowercase, single word when possible (e.g., `s2store`, not `s2_store`)
- **Types/Structs**: PascalCase (e.g., `S2Store`, `S2Edge`, `TrackMeta`)
- **Methods/Functions**: PascalCase (e.g., `NewS2Store`, `GetStat`)
- **Private members**: starts with lowercase (e.g., `lastPointId`, `edges`)
- **Constants**: PascalCase (e.g., `NIL_ID`)
- **JSON/BSON struct tags**: lowercase with underscores (`json:"field_name"`)

### Imports
Grouped by comment-separated sections (gofmt does this automatically):
1. Standard library
2. External packages (third-party)
3. Internal packages (local `mnezerka/geonet/*`)

Example:
```go
import (
    "mnezerka/geonet/config"
    "mnezerka/geonet/log"
    "mnezerka/geonet/store"
    "mnezerka/geonet/tracks"

    "github.com/stretchr/testify/assert"
    "github.com/tkrajina/gpxgo/gpx"
)
```

### Error Handling
- **Prefer returning errors**: For public API methods that callers may need to handle
- **Use panic for unrecoverable states**: Only in internal functions when programming errors occur
- **Use log.Exitf for fatal errors with logging**: When the program should terminate with a message

### Maps and Slices
- Always initialize maps before use with `make(map[Key]Value)`
- For slices, append is preferred

### For Loops
- Uses C-style loops: `for i := 0; i < len(x); i++`
- Range is used when index is not needed

### Logging
- Uses custom `mnezerka/geonet/log` package
- Levels: DEBUG, INFO, WARN, ERROR (configurable via CLI flags `-v`, `-q`, `--verbosity`)

### Test Files
- Named `*_test.go`
- Use `testing` package combined with `stretchr/testify/assert`:
```go
func TestSimplifySingleTrack(t *testing.T) {
    assert.Nil(t, s.AddGpx(track))
    assert.Len(t, locIds, 2)
}
```

## Directory Structure

```
geonet/
├── cmd/              CLI commands (Cobra framework)
├── config/           Configuration handling
├── log/              Logging utilities
├── mongostore/       MongoDB-backed store implementation
├── s2store/         S2 spatial index store (primary implementation)
├── store/           Core store definitions
├── svg/             SVG generation
├── tracks/          GPX track handling
├── utils/           Utility functions
├── main.go          Entry point
└── go.mod
```

## Key Patterns

### Store Pattern
```go
type S2Store struct {
    cfg         *config.Configuration
    index       *SpatialIndex
    lastPointId int64
    lastTrackId int64
    tracks      map[int64]*store.Track
    edges       map[S2EdgeKey]*S2Edge
    stat        store.Stat
}
```

### Edge ID Pattern
Edge IDs are keyed by point pairs (sorted to avoid duplicates):
```go
type S2EdgeKey struct {
    P1 int64 `json:"p1"`
    P2 int64 `json:"p2"`
}
```

## CLI Development

Uses Cobra framework. Commands are in `cmd/`:
- Root command in `cmd/root.go`
- Subcommands in separate files (e.g., `cmd_net.go`, `cmd_tracks.go`)

## Testing Guidelines

1. Test file names must end with `_test.go`
2. Test functions must start with `Test`: `func TestFunctionName(t *testing.T)`
3. Use descriptive test names: `TestSimplifySingleTrack`, not `Test1`
4. Use stretchr/testify assertions for readable failures

### Adding a new test
```go
package s2store

import (
    "mnezerka/geonet/config"
    "mnezerka/geonet/tracks"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestMyNewFeature(t *testing.T) {
    cfg := config.Cfg
    s := NewS2Store(&cfg)
    // ...
}
```

### Running a specific test
```bash
go test -v ./packagename -run TestFunctionName
```