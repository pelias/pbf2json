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
