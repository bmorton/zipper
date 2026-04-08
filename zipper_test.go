package zipper

import (
	"regexp"
	"testing"
)

func TestLookupKnown(t *testing.T) {
	tests := []struct {
		zip       string
		wantCity  string
		wantState string
	}{
		{"10001", "New York", "NY"},
		{"90210", "Beverly Hills", "CA"},
	}

	for _, tt := range tests {
		t.Run(tt.zip, func(t *testing.T) {
			entries := Lookup(tt.zip)
			if entries == nil {
				t.Fatalf("Lookup(%q) returned nil", tt.zip)
			}

			found := false
			for _, e := range entries {
				if e.City == tt.wantCity && e.StateCode == tt.wantState {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Lookup(%q) = %v, want city=%q state=%q", tt.zip, entries, tt.wantCity, tt.wantState)
			}
		})
	}
}

func TestLookupNotFound(t *testing.T) {
	entries := Lookup("00000")
	if entries != nil {
		t.Errorf("Lookup(00000) = %v, want nil", entries)
	}
}

func TestLookupMultiple(t *testing.T) {
	// Find a zip with multiple entries by scanning the data
	ensureLoaded()
	var multiZip string
	for zip, entries := range zipMap {
		if len(entries) > 1 {
			multiZip = zip
			break
		}
	}
	if multiZip == "" {
		t.Skip("No multi-entry zip found in dataset")
	}

	entries := Lookup(multiZip)
	if len(entries) <= 1 {
		t.Errorf("Lookup(%q) returned %d entries, want > 1", multiZip, len(entries))
	}
}

func TestLookupOne(t *testing.T) {
	e := LookupOne("10001")
	if e == nil {
		t.Fatal("LookupOne(10001) returned nil")
	}
	if e.ZipCode != "10001" {
		t.Errorf("LookupOne(10001).ZipCode = %q, want 10001", e.ZipCode)
	}

	e = LookupOne("00000")
	if e != nil {
		t.Errorf("LookupOne(00000) = %v, want nil", e)
	}
}

func TestNearestSanity(t *testing.T) {
	// Near Empire State Building
	e := Nearest(40.748, -73.985)
	if e == nil {
		t.Fatal("Nearest returned nil")
	}
	if e.StateCode != "NY" {
		t.Errorf("Nearest(40.748, -73.985).StateCode = %q, want NY", e.StateCode)
	}
}

func TestNearestN(t *testing.T) {
	results := NearestN(40.748, -73.985, 10)
	if len(results) != 10 {
		t.Fatalf("NearestN(10) returned %d results, want 10", len(results))
	}

	// Verify distance ordering
	for i := 1; i < len(results); i++ {
		d1 := haversine(40.748, -73.985, results[i-1].Lat, results[i-1].Lon)
		d2 := haversine(40.748, -73.985, results[i].Lat, results[i].Lon)
		if d1 > d2 {
			t.Errorf("NearestN results not distance-ordered: entry %d (%.4f km) > entry %d (%.4f km)", i-1, d1, i, d2)
		}
	}
}

func TestWithinRadius(t *testing.T) {
	results := WithinRadius(40.748, -73.985, 5.0)
	if len(results) == 0 {
		t.Fatal("WithinRadius(5km from NYC) returned no results")
	}

	for i, e := range results {
		d := haversine(40.748, -73.985, e.Lat, e.Lon)
		if d > 5.0 {
			t.Errorf("WithinRadius result %d at distance %.4f km exceeds radius 5.0 km", i, d)
		}
	}

	// Verify distance ordering
	for i := 1; i < len(results); i++ {
		d1 := haversine(40.748, -73.985, results[i-1].Lat, results[i-1].Lon)
		d2 := haversine(40.748, -73.985, results[i].Lat, results[i].Lon)
		if d1 > d2 {
			t.Errorf("WithinRadius results not distance-ordered at index %d", i)
		}
	}
}

func TestWithinRadiusZero(t *testing.T) {
	results := WithinRadius(40.748, -73.985, 0)
	// Should return empty or only an exact match
	for _, e := range results {
		d := haversine(40.748, -73.985, e.Lat, e.Lon)
		if d > 0 {
			t.Errorf("WithinRadius(0) returned entry at distance %.4f km", d)
		}
	}
}

func TestDataNotEmpty(t *testing.T) {
	n := Size()
	if n <= 30000 {
		t.Errorf("Size() = %d, want > 30000", n)
	}
}

func TestDataVersion(t *testing.T) {
	v := DataVersion()
	matched, err := regexp.MatchString(`^\d{4}\.\d{2}$`, v)
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Errorf("DataVersion() = %q, want YYYY.MM pattern", v)
	}
}

func TestDeterminism(t *testing.T) {
	a := All()
	b := All()
	if len(a) != len(b) {
		t.Fatalf("All() lengths differ: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Errorf("All()[%d] differs: %v vs %v", i, a[i], b[i])
			break
		}
	}
}

// Benchmarks

func BenchmarkLookup(b *testing.B) {
	ensureLoaded()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Lookup("10001")
	}
}

func BenchmarkNearest(b *testing.B) {
	ensureLoaded()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Nearest(40.748, -73.985)
	}
}

func BenchmarkNearestN_10(b *testing.B) {
	ensureLoaded()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NearestN(40.748, -73.985, 10)
	}
}

func BenchmarkWithinRadius_50(b *testing.B) {
	ensureLoaded()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WithinRadius(40.748, -73.985, 50)
	}
}

func BenchmarkInit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Can't truly re-init sync.Once, so this benchmarks the no-op path
		ensureLoaded()
	}
}
