package main

import "github.com/paulmach/go.geo"

// GetLineCentroid - compute the centroid of a line string
func GetLineCentroid(ps *geo.PointSet) *geo.Point {

	path := geo.NewPath()
	path.PointSet = *ps

	halfDistance := path.Distance() / 2
	travelled := 0.0

	for i := 0; i < len(path.PointSet)-1; i++ {

		segment := geo.NewLine(&path.PointSet[i], &path.PointSet[i+1])
		distance := segment.Distance()

		// middle line segment
		if (travelled + distance) > halfDistance {
			var remainder = halfDistance - travelled
			return segment.Interpolate(remainder / distance)
		}

		travelled += distance
	}

	return ps.GeoCentroid()
}
