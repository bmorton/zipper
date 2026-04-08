# zipper

Zero-dependency Go library for US zip code lookups with embedded data. No API keys, no network calls, no configuration.

## Install

```
go get github.com/bmorton/zipper
```

Import as:

```go
import "github.com/bmorton/zipper"
```

The module uses CalVer (`v0.YYYY.MMPATCH`). Upgrade to get fresher data; the API is stable.

## Data Model

Every function returns one or more `zipper.Entry` values:

```go
type Entry struct {
    CountryCode string  // "US"
    ZipCode     string  // e.g. "10001"
    City        string  // e.g. "New York"
    State       string  // full name, e.g. "New York"
    StateCode   string  // abbreviation, e.g. "NY"
    Lat         float64
    Lon         float64
}
```

A single zip code may map to multiple entries when it spans multiple cities.

## API Reference

### Lookup by zip code

```go
// Returns all entries for a zip (nil if not found).
// A zip may have multiple entries when it spans cities.
entries := zipper.Lookup("94608") // []Entry or nil

// Returns the first entry for a zip (nil if not found).
// Use when you only need one result.
entry := zipper.LookupOne("10001") // *Entry or nil
```

### Geo queries (coordinates are lat, lon as float64)

```go
// Closest zip code to a point (Haversine distance).
nearest := zipper.Nearest(40.748, -73.985) // *Entry

// N closest zip codes, ordered by distance ascending.
nearest10 := zipper.NearestN(40.748, -73.985, 10) // []Entry

// All zip codes within a radius (in kilometers), ordered by distance ascending.
nearby := zipper.WithinRadius(40.748, -73.985, 10.0) // []Entry
```

### Dataset metadata

```go
zipper.DataVersion() // string, e.g. "2025.04"
zipper.Size()        // int, total number of entries (~41k)
zipper.All()         // []Entry, entire dataset (do not modify)
```

## Behavior Notes

- All data is embedded in the binary. There are no external files, network calls, or runtime dependencies.
- Data loads lazily on first call and is cached for the process lifetime (thread-safe via `sync.Once`).
- `Lookup` and `LookupOne` return `nil` when a zip code is not found. Always nil-check before use.
- Geo queries use the Haversine formula and return results ordered by distance (ascending).
- The dataset contains ~41,000 US zip code entries sourced from GeoNames and the US Census Bureau.
- Zip codes that span multiple cities (e.g. 94608 → Emeryville + Oakland) return all associated entries.

## Common Patterns

### Get city and state from a zip code

```go
entry := zipper.LookupOne(zipCode)
if entry == nil {
    // handle invalid/unknown zip
}
city := entry.City
state := entry.StateCode
```

### Find nearby zip codes for a location-based feature

```go
results := zipper.WithinRadius(userLat, userLon, 25.0) // 25 km radius
for _, r := range results {
    fmt.Printf("%s %s, %s\n", r.ZipCode, r.City, r.StateCode)
}
```

### Validate a zip code exists

```go
if zipper.LookupOne(input) == nil {
    return errors.New("invalid US zip code")
}
```
