# zipper

A zero-dependency Go library for US zip code lookups, powered by embedded [GeoNames](https://www.geonames.org/) postal code data.

## Install

```
go get github.com/bmorton/zipper
```

The module version follows CalVer (`v0.YYYY.MMPATCH`) to track data vintage. Upgrade to get fresher data; the API is stable.

## Usage

```go
package main

import (
    "fmt"
    "github.com/bmorton/zipper"
)

func main() {
    // Look up a zip code
    e := zipper.LookupOne("10001")
    fmt.Printf("%s, %s %s\n", e.City, e.StateCode, e.ZipCode)
    // Output: New York, NY 10001

    // Find the nearest zip to coordinates
    nearest := zipper.Nearest(40.748, -73.985)
    fmt.Printf("Nearest: %s %s\n", nearest.City, nearest.ZipCode)

    // Find 5 nearest zip codes
    nearby := zipper.NearestN(40.748, -73.985, 5)
    for _, n := range nearby {
        fmt.Printf("  %s %s\n", n.City, n.ZipCode)
    }

    // Find all zips within 10km
    within := zipper.WithinRadius(40.748, -73.985, 10.0)
    fmt.Printf("Found %d zips within 10km\n", len(within))

    // Dataset info
    fmt.Printf("Data version: %s, entries: %d\n", zipper.DataVersion(), zipper.Size())
}
```

## API

| Function | Description |
|---|---|
| `Lookup(zip) []Entry` | All entries for a zip code (nil if not found) |
| `LookupOne(zip) *Entry` | First entry for a zip code (nil if not found) |
| `Nearest(lat, lon) *Entry` | Closest zip to coordinates (Haversine) |
| `NearestN(lat, lon, n) []Entry` | N closest zips, distance-ordered |
| `WithinRadius(lat, lon, km) []Entry` | All zips within radius, distance-ordered |
| `All() []Entry` | Every entry in the dataset (do not modify) |
| `DataVersion() string` | Data vintage, e.g. "2025.04" |
| `Size() int` | Total number of entries |

## CLI

A demo CLI is included for quick lookups:

```bash
# Build the CLI
make build

# Look up a zip code
./bin/zipper lookup 10001
# ZIP    CITY      STATE          LAT      LON
# 10001  New York  New York (NY)  40.7484  -73.9967

# Find the 3 nearest zips to coordinates
./bin/zipper nearest -n 3 40.748 -73.985
# ZIP    CITY      STATE          LAT      LON
# 10118  New York  New York (NY)  40.7490  -73.9865
# 10120  New York  New York (NY)  40.7506  -73.9894
# 10123  New York  New York (NY)  40.7515  -73.9905

# Find all zips within 5 km
./bin/zipper within 40.748 -73.985 5

# Show dataset info
./bin/zipper info
```

## Data

The embedded dataset is built from two sources:

- **[GeoNames US Postal Codes](https://download.geonames.org/export/zip/US.zip)** — coordinates, state, and base place names for ~41k zip codes. Licensed under [CC-BY 4.0](https://creativecommons.org/licenses/by/4.0/).
- **[US Census Bureau ZCTA-to-Place Relationship File](https://www.census.gov/geographies/reference-files/time-series/geo/relationship-files.2020.html)** — maps zip codes to all overlapping Census Places (cities), providing multi-city coverage. Public domain (US government work).

This means zip codes that span multiple cities (e.g. 94608 → Emeryville + Oakland) correctly return all associated place names.

Data is refreshed monthly via GitHub Actions.

## Regenerating Data

```
make generate
```

This downloads the latest GeoNames export and Census ZCTA-Place file, merges them, and writes `data/zipcodes.csv.gz`.
