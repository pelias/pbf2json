package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeCentroidWithEntranceNode(t *testing.T) {

	var latlons = []map[string]string{
		map[string]string{"lat": "1", "lon": "2", "entrance": "1"},
	}

	var centroid = computeCentroid(latlons)
	assert.Equal(t, "1", centroid["lat"])
	assert.Equal(t, "2", centroid["lon"])
}

func TestComputeCentroidWithMainEntranceNode(t *testing.T) {

	var latlons = []map[string]string{
		map[string]string{"lat": "0", "lon": "0", "entrance": "1"},
		map[string]string{"lat": "1", "lon": "2", "entrance": "2"},
		map[string]string{"lat": "-1", "lon": "-2", "entrance": "1", "wheelchair": "2"},
	}

	var centroid = computeCentroid(latlons)
	assert.Equal(t, "1", centroid["lat"])
	assert.Equal(t, "2", centroid["lon"])
}

func TestComputeCentroidWithAccessibleEntranceNode(t *testing.T) {

	var latlons = []map[string]string{
		map[string]string{"lat": "0", "lon": "0", "entrance": "1"},
		map[string]string{"lat": "-1", "lon": "-2", "entrance": "1", "wheelchair": "2"},
	}

	var centroid = computeCentroid(latlons)
	assert.Equal(t, "-1", centroid["lat"])
	assert.Equal(t, "-2", centroid["lon"])
}

func TestComputeCentroidWithRegularEntranceNode(t *testing.T) {

	var latlons = []map[string]string{
		map[string]string{"lat": "0", "lon": "0", "entrance": "1"},
	}

	var centroid = computeCentroid(latlons)
	assert.Equal(t, "0", centroid["lat"])
	assert.Equal(t, "0", centroid["lon"])
}

func TestComputeCentroidForClosedPolygon(t *testing.T) {

	var latlons = []map[string]string{
		map[string]string{"lat": "1", "lon": "1"},
		map[string]string{"lat": "-1", "lon": "1"},
		map[string]string{"lat": "-1", "lon": "-1"},
		map[string]string{"lat": "1", "lon": "-1"},
		map[string]string{"lat": "1", "lon": "1"},
	}

	var centroid = computeCentroid(latlons)
	assert.Equal(t, "0.0000000", centroid["lat"])
	assert.Equal(t, "0.0000000", centroid["lon"])
}

func TestComputeCentroidForHillboroPublicLibrary(t *testing.T) {

	var latlons = []map[string]string{
		map[string]string{"lat": "45.5424694", "lon": "-122.9356798"},
		map[string]string{"lat": "45.5424261", "lon": "-122.9361523"},
		map[string]string{"lat": "45.5432827", "lon": "-122.9363111"},
		map[string]string{"lat": "45.5433259", "lon": "-122.9358387"},
		map[string]string{"lat": "45.5430581", "lon": "-122.9357890"},
		map[string]string{"lat": "45.5429060", "lon": "-122.9357608"},
		map[string]string{"lat": "45.5424694", "lon": "-122.9356798"},
	}

	var centroid = computeCentroid(latlons)
	assert.Equal(t, "45.5428760", centroid["lat"])
	assert.Equal(t, "-122.9359955", centroid["lon"])
}

func TestComputeCentroidForOpenLineString(t *testing.T) {

	var latlons = []map[string]string{
		map[string]string{"lat": "1", "lon": "1"},
		map[string]string{"lat": "0", "lon": "0"},
		map[string]string{"lat": "-1", "lon": "-1"},
	}

	var centroid = computeCentroid(latlons)
	assert.Equal(t, "0.0000000", centroid["lat"])
	assert.Equal(t, "0.0000000", centroid["lon"])
}
