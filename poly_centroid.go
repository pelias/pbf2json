package main

import (
	"math"

	"github.com/paulmach/go.geo"
)

// GetPolygonCentroid - compute the centroid of a polygon set
// using a spherical co-ordinate system
func GetPolygonCentroid(ps *geo.PointSet) *geo.Point {

	// remove any vertices not requried for the centroid calculation
	var simplified = simplify(ps, 1e-2)

	// GeoCentroid function added in https://github.com/paulmach/go.geo/pull/24
	return simplified.GeoCentroid()
}

// remove any vertices which do not contribute to the 'shape' of the polygon
// ie. they lie on a bearing between their neighbours (within a threshhold).
func simplify(ps *geo.PointSet, threshhold float64) *geo.PointSet {
	var res = geo.NewPointSet()

	// iterate over all points in polygon
	// note: the start and end points may be duplicated
	for i := 0; i < ps.Length(); i++ {

		// find the point before
		var prevI = i - 1
		if prevI < 0 {
			prevI = ps.Length() + prevI
		}

		// find the point after
		var nextI = i + 1
		if nextI > ps.Length()-1 {
			nextI = nextI - ps.Length()
		}

		var previous = ps.GetAt(prevI)
		var this = ps.GetAt(i)
		var next = ps.GetAt(nextI)

		// ensure we are not dealing with nulls
		if nil == previous || nil == this || nil == next {
			res.Push(this)
			continue
		}

		// compute the bearings
		var B1 = previous.BearingTo(this)
		var B2 = previous.BearingTo(next)

		// keep any points with a difference in bearing greater than threshhold.
		if math.Abs(B1-B2) > threshhold {
			res.Push(this)
		}
	}

	return res
}
