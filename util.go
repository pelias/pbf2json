package main

import (
	geo "github.com/paulmach/go.geo"
	"math"
	"strconv"
)

func IsPointSetClosed(points *geo.PointSet) bool {
	if points.Length() > 2 {
		return points.First().Equals(points.Last())
	}
	return false
}

func PointToLatLon(point *geo.Point) map[string]string {
	var latLon = make(map[string]string)
	latLon["lat"] = strconv.FormatFloat(point.Lat(), 'f', 7, 64)
	latLon["lon"] = strconv.FormatFloat(point.Lng(), 'f', 7, 64)

	return latLon
}

func GetAreaOfBounds(bound *geo.Bound) float64 {
	return math.Max(bound.GeoWidth(), 0.000001) * math.Max(bound.GeoHeight(), 0.000001)
}
