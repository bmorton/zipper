package zipper

import (
	"compress/gzip"
	"embed"
	"encoding/csv"
	"io"
	"strconv"
	"sync"
)

//go:embed data/zipcodes.csv.gz
var dataFS embed.FS

var (
	once    sync.Once
	zipMap  map[string][]Entry
	allData []Entry
	version string
)

func ensureLoaded() {
	once.Do(func() {
		f, err := dataFS.Open("data/zipcodes.csv.gz")
		if err != nil {
			panic("zipper: failed to open embedded data: " + err.Error())
		}
		defer f.Close()

		gz, err := gzip.NewReader(f)
		if err != nil {
			panic("zipper: failed to create gzip reader: " + err.Error())
		}
		defer gz.Close()

		if gz.Comment != "" {
			version = gz.Comment
		}

		r := csv.NewReader(gz)

		// Read header
		header, err := r.Read()
		if err != nil {
			panic("zipper: failed to read CSV header: " + err.Error())
		}

		colIndex := make(map[string]int, len(header))
		for i, col := range header {
			colIndex[col] = i
		}

		entries := make([]Entry, 0, 42000) // pre-allocate for ~40k records
		m := make(map[string][]Entry, 42000)

		for {
			record, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				panic("zipper: failed to read CSV record: " + err.Error())
			}

			lat, _ := strconv.ParseFloat(record[colIndex["latitude"]], 64)
			lon, _ := strconv.ParseFloat(record[colIndex["longitude"]], 64)

			e := Entry{
				CountryCode: record[colIndex["country_code"]],
				ZipCode:     record[colIndex["postal_code"]],
				City:        record[colIndex["place_name"]],
				State:       record[colIndex["admin_name1"]],
				StateCode:   record[colIndex["admin_code1"]],
				Lat:         lat,
				Lon:         lon,
			}

			entries = append(entries, e)
		}

		allData = entries

		for i := range allData {
			zip := allData[i].ZipCode
			m[zip] = append(m[zip], allData[i])
		}
		zipMap = m
	})
}
