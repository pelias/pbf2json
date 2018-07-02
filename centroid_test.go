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

func TestComputeCentroidWithAccessibleEntranceNode(t *testing.T) {

	var latlons = []map[string]string{
		map[string]string{"lat": "1", "lon": "2", "entrance": "1"},
		map[string]string{"lat": "-1", "lon": "-2", "entrance": "1", "wheelchair": "1"},
	}

	var centroid = computeCentroid(latlons)
	assert.Equal(t, "-1", centroid["lat"])
	assert.Equal(t, "-2", centroid["lon"])
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
	assert.Equal(t, "0.000000", centroid["lat"])
	assert.Equal(t, "0.000000", centroid["lon"])
}

func TestComputeCentroidForOpenLineString(t *testing.T) {

	var latlons = []map[string]string{
		map[string]string{"lat": "1", "lon": "1"},
		map[string]string{"lat": "0", "lon": "0"},
		map[string]string{"lat": "-1", "lon": "-1"},
	}

	var centroid = computeCentroid(latlons)
	assert.Equal(t, "0.000000", centroid["lat"])
	assert.Equal(t, "0.000000", centroid["lon"])
}
