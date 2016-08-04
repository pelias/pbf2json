package main

import (
	"testing"

	"github.com/paulmach/go.geo"
	"github.com/stretchr/testify/assert"
)

// http://www.openstreetmap.org/way/46340228
func TestGetLineCentroid(t *testing.T) {

	var poly = geo.NewPointSet()
	poly.Push(geo.NewPoint(-74.001559, 40.719743))
	poly.Push(geo.NewPoint(-73.999914, 40.721679))
	poly.Push(geo.NewPoint(-73.997783, 40.724195))
	poly.Push(geo.NewPoint(-73.997318, 40.724745))
	poly.Push(geo.NewPoint(-73.996797, 40.725375))
	poly.Push(geo.NewPoint(-73.995203, 40.727239))
	poly.Push(geo.NewPoint(-73.993927, 40.728737))
	poly.Push(geo.NewPoint(-73.992407, 40.730535))
	poly.Push(geo.NewPoint(-73.991545, 40.731566))
	poly.Push(geo.NewPoint(-73.991417, 40.731843))
	poly.Push(geo.NewPoint(-73.990745, 40.734738))
	poly.Push(geo.NewPoint(-73.990199, 40.737495))
	poly.Push(geo.NewPoint(-73.989630, 40.739735))
	poly.Push(geo.NewPoint(-73.989370, 40.741459))
	poly.Push(geo.NewPoint(-73.989219, 40.742233))
	poly.Push(geo.NewPoint(-73.989119, 40.743025))
	poly.Push(geo.NewPoint(-73.988699, 40.745262))
	poly.Push(geo.NewPoint(-73.987904, 40.749446))
	poly.Push(geo.NewPoint(-73.987417, 40.752149))
	poly.Push(geo.NewPoint(-73.986938, 40.754016))
	poly.Push(geo.NewPoint(-73.986833, 40.754345))
	poly.Push(geo.NewPoint(-73.986321, 40.755897))
	poly.Push(geo.NewPoint(-73.986117, 40.756513))
	poly.Push(geo.NewPoint(-73.985720, 40.757348))
	poly.Push(geo.NewPoint(-73.985433, 40.757980))
	poly.Push(geo.NewPoint(-73.983607, 40.760503))
	poly.Push(geo.NewPoint(-73.979957, 40.765504))
	poly.Push(geo.NewPoint(-73.979264, 40.766480))

	var centroid = GetLineCentroid(poly)
	assert.Equal(t, 40.74239780132512, centroid.Lat())
	assert.Equal(t, -73.98919819175188, centroid.Lng())
}
