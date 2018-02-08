package main

import "github.com/paulmach/go.geo"

// GetPolygonCentroid - compute the centroid of a polygon set
// using a spherical co-ordinate system
func GetPolygonCentroid(ps *geo.PointSet) *geo.Point {
	// GeoCentroid function added in https://github.com/paulmach/go.geo/pull/24
	return ps.GeoCentroid()
}
