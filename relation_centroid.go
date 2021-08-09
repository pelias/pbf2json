package main

import (
	geo "github.com/paulmach/go.geo"
	"github.com/qedus/osmpbf"
	"log"
)

// ComputeRelationCentroidAndBounds Favors lat, long of member with role admin_centre for centroid, computes outer area if they are non-closed ways
func ComputeRelationCentroidAndBounds(relMemberLatLons map[osmpbf.Member][]map[string]string) (map[string]string, *geo.Bound) {
	centroid, isCentroidAdminCentre, bounds, openOuterAreaMemberPoints := getCentroidAndBoundsAndOpenWays(relMemberLatLons)

	if len(openOuterAreaMemberPoints) > 0 {
		outerPolygons := getPolygonsFromWays(openOuterAreaMemberPoints)

		if len(outerPolygons) > 0 {
			largestArea := 0.0

			if bounds != nil {
				largestArea = GetAreaOfBounds(bounds)
			}

			largestOuterPolygon := getLargestAreaPolygon(outerPolygons)
			largestOuterPolygonBounds := largestOuterPolygon.Bound()

			if GetAreaOfBounds(largestOuterPolygonBounds) > largestArea {
				bounds = largestOuterPolygonBounds
				if !isCentroidAdminCentre {
					centroidPoint := GetPolygonCentroid(largestOuterPolygon)
					centroid = PointToLatLon(centroidPoint)
				}
			}
		}
	}

	return centroid, bounds
}

func getCentroidAndBoundsAndOpenWays(relMemberLatLons map[osmpbf.Member][]map[string]string) (map[string]string, bool, *geo.Bound, []*geo.PointSet) {
	openOuterAreaMemberPoints := make([]*geo.PointSet, 0)
	foundAdminCentre := false

	var largestArea = 0.0
	var centroid map[string]string
	var bounds *geo.Bound

	for member, memberWayLatLons := range relMemberLatLons {
		if len(memberWayLatLons) < 1 {
			continue
		}

		memberWayPointSet := LatLngMapToPointSet(memberWayLatLons)
		isClosedWay := IsPointSetClosed(memberWayPointSet)

		if member.Role == "outer" && !isClosedWay {
			openOuterAreaMemberPoints = append(openOuterAreaMemberPoints, memberWayPointSet)
		} else if member.Type == 0 && member.Role == "admin_centre" {
			// prefer admin_center for centroid over computing it
			var latlons = memberWayLatLons[0]

			latlons["type"] = "admin_centre"
			centroid = latlons

			foundAdminCentre = true
		} else {
			wayCentroid, wayBounds := ComputeCentroidAndBounds(memberWayLatLons)

			// if for any reason we failed to find a valid bounds
			if nil == wayBounds {
				log.Println("[warn] failed to calculate bounds for relation member way")
				continue
			}

			area := GetAreaOfBounds(wayBounds)

			// find the way with the largest area
			if area > largestArea {
				largestArea = area
				if !foundAdminCentre {
					centroid = wayCentroid
				}
				bounds = wayBounds
			}
		}
	}

	return centroid, foundAdminCentre, bounds, openOuterAreaMemberPoints
}

func getLargestAreaPolygon(polygons []*geo.PointSet) *geo.PointSet {
	largestOuterArea := 0.0
	var largestPolygon *geo.PointSet

	for _, polygon := range polygons {
		bounds := polygon.Bound()
		area := GetAreaOfBounds(bounds)

		if area > largestOuterArea {
			largestOuterArea = area
			largestPolygon = polygon
		}
	}

	return largestPolygon
}

func getClosedWays(ways []*geo.PointSet) []*geo.PointSet {
	closedWays := make([]*geo.PointSet, 0)

	for _, way := range ways {
		if way.First().Equals(way.Last()) {
			closedWays = append(closedWays, way)
		}
	}

	return closedWays
}

func getPolygonsFromWays(ways []*geo.PointSet) []*geo.PointSet {
	lastIterationOpenWayCount := len(ways)
	tryForClockwiseWays := false

	for len(ways) > 1 {
		ways = connectWays(ways, tryForClockwiseWays)

		if len(ways) == lastIterationOpenWayCount {
			if tryForClockwiseWays {
				// no clockwise ways, abort
				break
			} else {
				// no directly connecting ways left, check for clockwise connecting ways
				tryForClockwiseWays = true
				lastIterationOpenWayCount = 0
			}
		}

		lastIterationOpenWayCount = len(ways)
	}

	closedWays := getClosedWays(ways)

	return closedWays
}

func connectWays(ways []*geo.PointSet, tryClockwiseWays bool) []*geo.PointSet {
	for i, way := range ways {
		for j, wayToCheck := range ways {
			if i == j {
				continue
			}

			connectingWay := geo.NewPointSet()

			if way.Last().Equals(wayToCheck.First()) {
				connectingWay = mergeWaysIntoWay([]*geo.PointSet{way, wayToCheck})
			} else if way.First().Equals(wayToCheck.Last()) {
				connectingWay = mergeWaysIntoWay([]*geo.PointSet{wayToCheck, way})
			} else if tryClockwiseWays {
				if way.Last().Equals(wayToCheck.Last()) {
					reversedWay := reverseWay(wayToCheck)

					connectingWay = mergeWaysIntoWay([]*geo.PointSet{way, reversedWay})
				} else if way.First().Equals(wayToCheck.First()) {
					reversedWay := reverseWay(way)

					connectingWay = mergeWaysIntoWay([]*geo.PointSet{reversedWay, wayToCheck})
				}
			}

			if connectingWay.Length() > 0 {
				ways = removeWaysByIndices([]int{i, j}, ways)
				ways = append(ways, connectingWay)

				return ways
			}
		}
	}

	return ways
}

func mergeWaysIntoWay(ways []*geo.PointSet) *geo.PointSet {
	singleWay := geo.NewPointSet()

	for _, way := range ways {
		for i := 0; i < way.Length(); i++ {
			singleWay.Push(way.GetAt(i))
		}
	}

	return singleWay
}

func removeWaysByIndices(indices []int, ways []*geo.PointSet) []*geo.PointSet {
	waysWithoutIndices := make([]*geo.PointSet, 0)

	for wayIndex := 0; wayIndex < len(ways); wayIndex++ {
		isMatchingAnyIndex := false
		for _, index := range indices {
			if index == wayIndex {
				isMatchingAnyIndex = true
				break
			}
		}
		if !isMatchingAnyIndex {
			waysWithoutIndices = append(waysWithoutIndices, ways[wayIndex])
		}
	}

	return waysWithoutIndices
}

func reverseWay(points *geo.PointSet) *geo.PointSet {
	reversedWay := geo.NewPointSet()

	for i := points.Length() - 1; i >= 0; i-- {
		reversedWay.Push(points.GetAt(i))
	}

	return reversedWay
}
