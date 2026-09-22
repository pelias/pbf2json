package main

import (
	"math"
	"testing"

	"github.com/qedus/osmpbf"
	"github.com/stretchr/testify/assert"
)

func TestEncodingSimple(t *testing.T) {

	var node = &osmpbf.Node{ID: 100, Lat: -50, Lon: 77}
	var expectedBytes = []byte{0xc0, 0x49, 0x0, 0x0, 0x0, 0x0, 0x40, 0x53, 0x40, 0x0, 0x0, 0x0}
	var expectedLatlon = map[string]string{"lon": "77.0000000", "lat": "-50.0000000"}

	// encode
	var stringid, byteval = nodeToBytes(node)
	assert.Equal(t, "100", stringid)
	assert.Equal(t, expectedBytes, byteval)

	// decode
	var latlon = bytesToLatLon(byteval)
	assert.Equal(t, expectedLatlon, latlon)
}

func TestEncodingFloatPrecision(t *testing.T) {

	var node = &osmpbf.Node{ID: 100, Lat: -50.555555555, Lon: 77.777777777}
	var expectedBytes = []byte{0xc0, 0x49, 0x47, 0x1c, 0x71, 0xc5, 0x40, 0x53, 0x71, 0xc7, 0x1c, 0x70}
	var expectedLatlon = map[string]string{"lon": "77.7777778", "lat": "-50.5555556"}

	// encode
	var stringid, byteval = nodeToBytes(node)
	assert.Equal(t, "100", stringid)
	assert.Equal(t, expectedBytes, byteval)

	// decode
	var latlon = bytesToLatLon(byteval)
	assert.Equal(t, expectedLatlon, latlon)
}

func TestEncodingBitmaskValues(t *testing.T) {

	var tags = map[string]string{"entrance": "main", "wheelchair": "yes"}
	var node = &osmpbf.Node{ID: 100, Lat: -50, Lon: 77, Tags: tags}
	var expectedBytes = []byte{0xc0, 0x49, 0x0, 0x0, 0x0, 0x0, 0x40, 0x53, 0x40, 0x0, 0x0, 0x0, 0xa0}
	var expectedLatlon = map[string]string{"lon": "77.0000000", "lat": "-50.0000000", "entrance": "2", "wheelchair": "2"}

	// encode
	var stringid, byteval = nodeToBytes(node)
	assert.Equal(t, "100", stringid)
	assert.Equal(t, expectedBytes, byteval)

	// decode
	var latlon = bytesToLatLon(byteval)
	assert.Equal(t, expectedLatlon, latlon)
}

func TestEncodingAndDecodingIdsToBytes(t *testing.T) {

	var ids = []int64{0, 100, 100000, 100000000, math.MaxInt64}

	// encode
	var encoded = idSliceToBytes(ids)
	// assert.Equal(t, "100", stringid)
	// assert.Equal(t, expectedBytes, byteval)

	// decode
	var decoded = bytesToIDSlice(encoded)
	assert.Equal(t, decoded, ids)
}

func BenchmarkBytesToLatLon(b *testing.B) {
	node := &osmpbf.Node{
		ID:  123,
		Lat: 12.1234,
		Lon: -122.1234,
		Tags: map[string]string{
			"entrance": "main",
		},
	}
	_, data := nodeToBytes(node)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		bytesToLatLon(data)
	}
}

func BenchmarkNodeToBytes(b *testing.B) {
	node := &osmpbf.Node{
		ID:  123,
		Lat: 12.1234,
		Lon: -122.1234,
		Tags: map[string]string{
			"entrance": "main",
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		nodeToBytes(node)
	}
}

func BenchmarkIdSliceToBytes(b *testing.B) {
	ids := make([]int64, 100)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		idSliceToBytes(ids)
	}
}

func BenchmarkBytesToIDSlice(b *testing.B) {
	ids := make([]int64, 100)
	data := idSliceToBytes(ids)

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		bytesToIDSlice(data)
	}
}

// peliasTagList is the pre-parsed form of the default Pelias OSM importer tag conditions
// from openstreetmap/config/features.js, matching what getSettings() produces.
var peliasTagList = []tagGroup{
	{{key: "addr:housenumber"}, {key: "addr:street"}},
	{{key: "addr:housenumber"}, {key: "addr:place"}},
	{{key: "amenity"}, {key: "name"}},
	{{key: "building"}, {key: "name"}},
	{{key: "shop"}, {key: "name"}},
	{{key: "office"}, {key: "name"}},
	{{key: "public_transport"}, {key: "name"}},
	{{key: "cuisine"}, {key: "name"}},
	{{key: "railway", value: "station"}, {key: "name"}},
	{{key: "railway", value: "tram_stop"}, {key: "name"}},
	{{key: "railway", value: "halt"}, {key: "name"}},
	{{key: "railway", value: "subway_entrance"}, {key: "name"}},
	{{key: "railway", value: "train_station_entrance"}, {key: "name"}},
	{{key: "highway", value: "pedestrian"}, {key: "area", value: "yes"}, {key: "name"}},
	{{key: "place", value: "square"}, {key: "name"}},
	{{key: "sport"}, {key: "name"}},
	{{key: "natural"}, {key: "name"}},
	{{key: "tourism"}, {key: "name"}},
	{{key: "leisure"}, {key: "name"}},
	{{key: "historic"}, {key: "name"}},
	{{key: "man_made"}, {key: "name"}},
	{{key: "landuse"}, {key: "name"}},
	{{key: "waterway"}, {key: "name"}},
	{{key: "aerialway"}, {key: "name"}},
	{{key: "craft"}, {key: "name"}},
	{{key: "military"}, {key: "name"}},
	{{key: "aeroway", value: "terminal"}, {key: "name"}},
	{{key: "aeroway", value: "aerodrome"}, {key: "name"}},
	{{key: "aeroway", value: "helipad"}, {key: "name"}},
	{{key: "aeroway", value: "airstrip"}, {key: "name"}},
	{{key: "aeroway", value: "heliport"}, {key: "name"}},
	{{key: "aeroway", value: "areodrome"}, {key: "name"}},
	{{key: "aeroway", value: "spaceport"}, {key: "name"}},
	{{key: "aeroway", value: "landing_strip"}, {key: "name"}},
	{{key: "aeroway", value: "airfield"}, {key: "name"}},
	{{key: "aeroway", value: "airport"}, {key: "name"}},
	{{key: "brand"}, {key: "name"}},
	{{key: "healthcare"}, {key: "name"}},
}

// BenchmarkContainsValidTags_NoMatch benchmarks the common case: an element with tags
// that don't match any of the 37 pelias conditions (the vast majority of OSM elements).
func BenchmarkContainsValidTags_NoMatch(b *testing.B) {
	tags := map[string]string{
		"highway": "residential",
		"surface": "asphalt",
		"oneway":  "yes",
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		containsValidTags(tags, peliasTagList)
	}
}

// BenchmarkContainsValidTags_Match benchmarks the case where tags do match
// (e.g., a named amenity — common for venue records).
func BenchmarkContainsValidTags_Match(b *testing.B) {
	tags := map[string]string{
		"amenity": "restaurant",
		"name":    "The Green Elephant",
		"cuisine": "vegetarian",
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		containsValidTags(tags, peliasTagList)
	}
}

// BenchmarkTrimTags_NoTrimNeeded benchmarks the common case where no trimming is required.
func BenchmarkTrimTags_NoTrimNeeded(b *testing.B) {
	tags := map[string]string{
		"amenity":          "restaurant",
		"name":             "The Green Elephant",
		"cuisine":          "vegetarian",
		"addr:housenumber": "42",
		"addr:street":      "Main St",
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		trimTags(tags)
	}
}

// BenchmarkBitmaskInsert benchmarks inserting IDs into the bitmask.
func BenchmarkBitmaskInsert(b *testing.B) {
	mask := NewBitMask()
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		mask.Insert(int64(n))
	}
}

// BenchmarkBitmaskHas benchmarks checking membership in the bitmask (hit case).
func BenchmarkBitmaskHas(b *testing.B) {
	mask := NewBitMask()
	for i := 0; i < 1000; i++ {
		mask.Insert(int64(i))
	}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		mask.Has(int64(n % 1000))
	}
}
