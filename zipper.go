package zipper

import (
	"container/heap"
	"sort"
)

// Entry represents a single zip code record.
type Entry struct {
	CountryCode string // ISO 3166-1 alpha-2, e.g. "US"
	ZipCode     string
	City        string
	State       string // full name, e.g. "California"
	StateCode   string // abbreviation, e.g. "CA"
	Lat         float64
	Lon         float64
}

// Lookup returns all entries for a given zip code.
// A zip may map to multiple places (e.g. shared zips across cities).
// Returns nil if not found.
func Lookup(zip string) []Entry {
	ensureLoaded()
	entries, ok := zipMap[zip]
	if !ok {
		return nil
	}
	return entries
}

// LookupOne returns the first entry for a zip code, or nil if not found.
// Convenience wrapper for the common single-result case.
func LookupOne(zip string) *Entry {
	entries := Lookup(zip)
	if len(entries) == 0 {
		return nil
	}
	return &entries[0]
}

// Nearest returns the closest zip code entry to the given coordinates
// using the Haversine formula. Returns nil only if the dataset is empty.
func Nearest(lat, lon float64) *Entry {
	ensureLoaded()
	if len(allData) == 0 {
		return nil
	}

	bestIdx := 0
	bestDist := haversine(lat, lon, allData[0].Lat, allData[0].Lon)

	for i := 1; i < len(allData); i++ {
		d := haversine(lat, lon, allData[i].Lat, allData[i].Lon)
		if d < bestDist {
			bestDist = d
			bestIdx = i
		}
	}

	return &allData[bestIdx]
}

// distEntry pairs an entry with its distance for sorting/heap purposes.
type distEntry struct {
	entry Entry
	dist  float64
}

// maxHeap implements heap.Interface for a max-heap of distEntry.
type maxHeap []distEntry

func (h maxHeap) Len() int            { return len(h) }
func (h maxHeap) Less(i, j int) bool  { return h[i].dist > h[j].dist } // max-heap: largest first
func (h maxHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *maxHeap) Push(x interface{}) { *h = append(*h, x.(distEntry)) }
func (h *maxHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// NearestN returns the N closest entries to the given coordinates,
// ordered by distance ascending.
func NearestN(lat, lon float64, n int) []Entry {
	ensureLoaded()
	if n <= 0 {
		return nil
	}
	if n >= len(allData) {
		// Return all, sorted by distance
		result := make([]distEntry, len(allData))
		for i := range allData {
			result[i] = distEntry{entry: allData[i], dist: haversine(lat, lon, allData[i].Lat, allData[i].Lon)}
		}
		sort.Slice(result, func(i, j int) bool { return result[i].dist < result[j].dist })
		entries := make([]Entry, len(result))
		for i, de := range result {
			entries[i] = de.entry
		}
		return entries
	}

	h := &maxHeap{}
	heap.Init(h)

	for i := range allData {
		d := haversine(lat, lon, allData[i].Lat, allData[i].Lon)
		if h.Len() < n {
			heap.Push(h, distEntry{entry: allData[i], dist: d})
		} else if d < (*h)[0].dist {
			(*h)[0] = distEntry{entry: allData[i], dist: d}
			heap.Fix(h, 0)
		}
	}

	result := make([]distEntry, h.Len())
	for i := h.Len() - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).(distEntry)
	}

	entries := make([]Entry, len(result))
	for i, de := range result {
		entries[i] = de.entry
	}
	return entries
}

// WithinRadius returns all entries within radiusKM kilometers
// of the given coordinates, ordered by distance ascending.
func WithinRadius(lat, lon float64, radiusKM float64) []Entry {
	ensureLoaded()
	var matches []distEntry

	for i := range allData {
		d := haversine(lat, lon, allData[i].Lat, allData[i].Lon)
		if d <= radiusKM {
			matches = append(matches, distEntry{entry: allData[i], dist: d})
		}
	}

	sort.Slice(matches, func(i, j int) bool { return matches[i].dist < matches[j].dist })

	entries := make([]Entry, len(matches))
	for i, de := range matches {
		entries[i] = de.entry
	}
	return entries
}

// All returns every entry in the dataset. The caller must not modify the slice.
func All() []Entry {
	ensureLoaded()
	return allData
}

// DataVersion returns the data vintage, e.g. "2025.04".
func DataVersion() string {
	ensureLoaded()
	return version
}

// Size returns the total number of entries.
func Size() int {
	ensureLoaded()
	return len(allData)
}
