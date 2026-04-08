package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultGeoNamesSource = "https://download.geonames.org/export/zip/US.zip"
	defaultCensusSource   = "https://www2.census.gov/geo/docs/maps-data/data/rel2020/zcta520/tab20_zcta520_place20_natl.txt"
)

type record struct {
	countryCode string
	postalCode  string
	placeName   string
	stateName   string
	stateCode   string
	latitude    float64
	longitude   float64
}

// geoBase holds the base GeoNames data for a zip code (lat, lon, state info).
type geoBase struct {
	countryCode string
	stateName   string
	stateCode   string
	latitude    float64
	longitude   float64
}

func main() {
	output := flag.String("output", "data/zipcodes.csv.gz", "output path for the gzipped CSV")
	geoSource := flag.String("source", defaultGeoNamesSource, "URL of the GeoNames US zip archive")
	censusSource := flag.String("census-source", defaultCensusSource, "URL of the Census ZCTA-to-Place file")
	flag.Parse()

	// Step 1: Download and parse GeoNames
	fmt.Printf("Downloading GeoNames data from %s...\n", *geoSource)
	geoData, err := download(*geoSource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading GeoNames: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Parsing GeoNames records...")
	geoRecords, err := parseZipArchive(geoData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing GeoNames: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Parsed %d GeoNames records\n", len(geoRecords))

	// Validate coordinates
	var valid []record
	var skipped int
	for _, r := range geoRecords {
		if !validateCoords(r.latitude, r.longitude) {
			skipped++
			continue
		}
		valid = append(valid, r)
	}
	if skipped > 0 {
		fmt.Printf("Skipped %d records with out-of-bounds coordinates\n", skipped)
	}

	// Build base map: zip → geoBase (first entry wins for lat/lon/state)
	baseMap := make(map[string]geoBase, len(valid))
	geoPlaces := make(map[string][]string, len(valid)) // zip → GeoNames place names
	for _, r := range valid {
		if _, exists := baseMap[r.postalCode]; !exists {
			baseMap[r.postalCode] = geoBase{
				countryCode: r.countryCode,
				stateName:   r.stateName,
				stateCode:   r.stateCode,
				latitude:    r.latitude,
				longitude:   r.longitude,
			}
		}
		geoPlaces[r.postalCode] = append(geoPlaces[r.postalCode], r.placeName)
	}
	fmt.Printf("Unique zip codes from GeoNames: %d\n", len(baseMap))

	// Step 2: Download and parse Census ZCTA-to-Place
	fmt.Printf("Downloading Census ZCTA-Place data from %s...\n", *censusSource)
	censusData, err := download(*censusSource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error downloading Census data: %v\n", err)
		os.Exit(1)
	}

	censusPlaces, err := parseCensusPlaces(censusData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing Census data: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Census ZCTA-Place mappings: %d ZCTAs\n", len(censusPlaces))

	// Step 3: Merge — for each zip, use Census places if available, else GeoNames
	var records []record
	censusUsed := 0
	for zip, base := range baseMap {
		places, hasCensus := censusPlaces[zip]
		if hasCensus && len(places) > 0 {
			censusUsed++
			for _, place := range places {
				records = append(records, record{
					countryCode: base.countryCode,
					postalCode:  zip,
					placeName:   place,
					stateName:   base.stateName,
					stateCode:   base.stateCode,
					latitude:    base.latitude,
					longitude:   base.longitude,
				})
			}
		} else {
			// Fall back to GeoNames place names
			for _, place := range geoPlaces[zip] {
				records = append(records, record{
					countryCode: base.countryCode,
					postalCode:  zip,
					placeName:   place,
					stateName:   base.stateName,
					stateCode:   base.stateCode,
					latitude:    base.latitude,
					longitude:   base.longitude,
				})
			}
		}
	}
	fmt.Printf("Zips enriched with Census places: %d\n", censusUsed)

	// Deduplicate on (postal_code, place_name)
	deduped := deduplicate(records)
	fmt.Printf("After deduplication: %d records\n", len(deduped))

	// Sort by postal_code ASC, then place_name ASC
	sort.Slice(deduped, func(i, j int) bool {
		if deduped[i].postalCode != deduped[j].postalCode {
			return deduped[i].postalCode < deduped[j].postalCode
		}
		return deduped[i].placeName < deduped[j].placeName
	})

	// Check previous file size
	var prevSize int64
	if info, err := os.Stat(*output); err == nil {
		prevSize = info.Size()
	}

	// Write gzipped CSV
	if err := os.MkdirAll(filepath.Dir(*output), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	if err := writeGzippedCSV(*output, deduped); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(*output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error stating output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Records:   %d\n", len(deduped))
	fmt.Printf("  File size: %s\n", humanSize(info.Size()))
	if prevSize > 0 {
		delta := info.Size() - prevSize
		sign := "+"
		if delta < 0 {
			sign = ""
		}
		fmt.Printf("  Delta:     %s%s\n", sign, humanSize(delta))
	}
	fmt.Println("Done.")
}

func download(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func parseZipArchive(data []byte) ([]record, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}

	for _, f := range r.File {
		if f.Name == "US.txt" {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open US.txt: %w", err)
			}
			defer func() { _ = rc.Close() }()
			return parseTSV(rc)
		}
	}

	return nil, fmt.Errorf("US.txt not found in archive")
}

func parseTSV(r io.Reader) ([]record, error) {
	reader := csv.NewReader(r)
	reader.Comma = '\t'
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1 // variable fields

	var records []record
	for {
		fields, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read TSV: %w", err)
		}

		// GeoNames format:
		// 0: country_code, 1: postal_code, 2: place_name,
		// 3: admin_name1 (state), 4: admin_code1 (state abbr),
		// 5: admin_name2, 6: admin_code2, 7: admin_name3, 8: admin_code3,
		// 9: latitude, 10: longitude, 11: accuracy
		if len(fields) < 11 {
			continue
		}

		lat, err := strconv.ParseFloat(fields[9], 64)
		if err != nil {
			continue
		}
		lon, err := strconv.ParseFloat(fields[10], 64)
		if err != nil {
			continue
		}

		records = append(records, record{
			countryCode: fields[0],
			postalCode:  fields[1],
			placeName:   fields[2],
			stateName:   fields[3],
			stateCode:   fields[4],
			latitude:    lat,
			longitude:   lon,
		})
	}

	return records, nil
}

func validateCoords(lat, lon float64) bool {
	return lat >= 17.0 && lat <= 72.0 && lon >= -180.0 && lon <= -64.0
}

// placeSuffixRe strips Census place type suffixes like "city", "town", "CDP", etc.
var placeSuffixRe = regexp.MustCompile(`\s+(city|town|village|CDP|borough|municipality|comunidad|zona urbana|corporation|government|\(balance\)|County|county|City|township)$`)

// parseCensusPlaces parses the Census ZCTA-to-Place relationship file and returns
// a map of ZCTA (zip code) → list of place names.
func parseCensusPlaces(data []byte) (map[string][]string, error) {
	result := make(map[string][]string)
	scanner := bufio.NewScanner(bytes.NewReader(data))

	// Skip header
	_ = scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, "|")
		if len(fields) < 11 {
			continue
		}

		zcta := fields[1]   // GEOID_ZCTA5_20
		place := fields[10] // NAMELSAD_PLACE_20

		if zcta == "" || place == "" {
			continue
		}

		// Strip the place type suffix
		cleanPlace := placeSuffixRe.ReplaceAllString(place, "")

		result[zcta] = append(result[zcta], cleanPlace)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan Census data: %w", err)
	}

	// Deduplicate within each ZCTA
	for zcta, places := range result {
		seen := make(map[string]bool)
		var unique []string
		for _, p := range places {
			if !seen[p] {
				seen[p] = true
				unique = append(unique, p)
			}
		}
		result[zcta] = unique
	}

	return result, nil
}

func deduplicate(records []record) []record {
	type key struct {
		postalCode string
		placeName  string
	}
	seen := make(map[key]bool, len(records))
	var result []record
	for _, r := range records {
		k := key{r.postalCode, r.placeName}
		if !seen[k] {
			seen[k] = true
			result = append(result, r)
		}
	}
	return result
}

func writeGzippedCSV(path string, records []record) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	gw, err := gzip.NewWriterLevel(f, gzip.BestCompression)
	if err != nil {
		return err
	}

	// Store data version in gzip comment
	now := time.Now()
	gw.Comment = fmt.Sprintf("%d.%02d", now.Year(), now.Month())

	w := csv.NewWriter(gw)

	// Write header
	if err := w.Write([]string{
		"country_code", "postal_code", "place_name",
		"admin_name1", "admin_code1", "latitude", "longitude",
	}); err != nil {
		return err
	}

	for _, r := range records {
		if err := w.Write([]string{
			r.countryCode,
			r.postalCode,
			r.placeName,
			r.stateName,
			r.stateCode,
			formatFloat(r.latitude),
			formatFloat(r.longitude),
		}); err != nil {
			return err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}

	if err := gw.Close(); err != nil {
		return err
	}

	return f.Close()
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func humanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
