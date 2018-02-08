package main

import (
	"testing"

	"github.com/paulmach/go.geo"
	"github.com/stretchr/testify/assert"
)

// http://www.openstreetmap.org/way/264768896
func TestGetPolygonCentroid(t *testing.T) {

	var poly = geo.NewPointSet()
	poly.Push(geo.NewPoint(-73.989605, 40.740760))
	poly.Push(geo.NewPoint(-73.989615, 40.740762))
	poly.Push(geo.NewPoint(-73.989619, 40.740763))
	poly.Push(geo.NewPoint(-73.989855, 40.740864))
	poly.Push(geo.NewPoint(-73.989859, 40.740867))
	poly.Push(geo.NewPoint(-73.989866, 40.740874))
	poly.Push(geo.NewPoint(-73.989870, 40.740882))
	poly.Push(geo.NewPoint(-73.989872, 40.740891))
	poly.Push(geo.NewPoint(-73.989870, 40.740899))
	poly.Push(geo.NewPoint(-73.989865, 40.740907))
	poly.Push(geo.NewPoint(-73.989584, 40.741288))
	poly.Push(geo.NewPoint(-73.989575, 40.741294))
	poly.Push(geo.NewPoint(-73.989564, 40.741298))
	poly.Push(geo.NewPoint(-73.989559, 40.741300))
	poly.Push(geo.NewPoint(-73.989547, 40.741300))
	poly.Push(geo.NewPoint(-73.989535, 40.741299))
	poly.Push(geo.NewPoint(-73.989529, 40.741297))
	poly.Push(geo.NewPoint(-73.989519, 40.741293))
	poly.Push(geo.NewPoint(-73.989514, 40.741290))
	poly.Push(geo.NewPoint(-73.989507, 40.741283))
	poly.Push(geo.NewPoint(-73.989501, 40.741265))
	poly.Push(geo.NewPoint(-73.989570, 40.740776))
	poly.Push(geo.NewPoint(-73.989575, 40.740770))
	poly.Push(geo.NewPoint(-73.989581, 40.740765))
	poly.Push(geo.NewPoint(-73.989590, 40.740761))
	poly.Push(geo.NewPoint(-73.989595, 40.740760))
	poly.Push(geo.NewPoint(-73.989605, 40.740760))

	var centroid = GetPolygonCentroid(poly)
	assert.Equal(t, 40.74100992600508, centroid.Lat())
	assert.Equal(t, -73.98964244467275, centroid.Lng())
}
