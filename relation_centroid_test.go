package main

import (
	geo "github.com/paulmach/go.geo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPolygonFromWays(t *testing.T) {
	var way1 = geo.NewPointSet()
	way1.Push(geo.NewPoint(9.869660139083862, 53.545229135979696))
	way1.Push(geo.NewPoint(9.869220256805418, 53.54410713060241))

	var way2 = geo.NewPointSet()
	way2.Push(geo.NewPoint(9.869220256805418, 53.54410713060241))
	way2.Push(geo.NewPoint(9.80645354837179, 53.55131298705578))
	way2.Push(geo.NewPoint(9.806404933333397, 53.55129247064265))

	var way3 = geo.NewPointSet()
	way3.Push(geo.NewPoint(9.806404933333397, 53.55129247064265))
	way3.Push(geo.NewPoint(9.80623260140419, 53.55131577569369))
	way3.Push(geo.NewPoint(9.806125648319721, 53.55134983689904))
	way3.Push(geo.NewPoint(9.806103184819221, 53.55139963393354))

	var way4 = geo.NewPointSet()
	way4.Push(geo.NewPoint(9.806103184819221, 53.55139963393354))
	way4.Push(geo.NewPoint(9.869660139083862, 53.545229135979696))

	var toBeIgnoredWay = geo.NewPointSet()
	toBeIgnoredWay.Push(geo.NewPoint(9.123456111111111, 53.891011111111111))
	toBeIgnoredWay.Push(geo.NewPoint(9.234567111111111, 53.910111111111111))
	toBeIgnoredWay.Push(geo.NewPoint(9.345678111111111, 53.101112111111111))

	factoredPolygons := getPolygonsFromWays([]*geo.PointSet{way1, toBeIgnoredWay, way2, way3, way4})

	expectedPolygon := geo.NewPointSet()
	expectedPolygon.Push(geo.NewPoint(9.869660139083862, 53.545229135979696))
	expectedPolygon.Push(geo.NewPoint(9.869220256805418, 53.54410713060241))
	expectedPolygon.Push(geo.NewPoint(9.869220256805418, 53.54410713060241))
	expectedPolygon.Push(geo.NewPoint(9.80645354837179, 53.55131298705578))
	expectedPolygon.Push(geo.NewPoint(9.806404933333397, 53.55129247064265))
	expectedPolygon.Push(geo.NewPoint(9.806404933333397, 53.55129247064265))
	expectedPolygon.Push(geo.NewPoint(9.80623260140419, 53.55131577569369))
	expectedPolygon.Push(geo.NewPoint(9.806125648319721, 53.55134983689904))
	expectedPolygon.Push(geo.NewPoint(9.806103184819221, 53.55139963393354))
	expectedPolygon.Push(geo.NewPoint(9.806103184819221, 53.55139963393354))
	expectedPolygon.Push(geo.NewPoint(9.869660139083862, 53.545229135979696))

	assert.Len(t, factoredPolygons, 1)
	assert.True(t, expectedPolygon.Equals(factoredPolygons[0]))
}

func TestGetPolygonFromClockwiseWays(t *testing.T) {
	var way1 = geo.NewPointSet()
	way1.Push(geo.NewPoint(9.869660139083862, 53.545229135979696))
	way1.Push(geo.NewPoint(9.869220256805418, 53.54410713060241))

	var way2 = geo.NewPointSet()
	way2.Push(geo.NewPoint(9.869220256805418, 53.54410713060241))
	way2.Push(geo.NewPoint(9.80645354837179, 53.55131298705578))
	way2.Push(geo.NewPoint(9.806404933333397, 53.55129247064265))

	var way3 = geo.NewPointSet()
	way3.Push(geo.NewPoint(9.806404933333397, 53.55129247064265))
	way3.Push(geo.NewPoint(9.80623260140419, 53.55131577569369))
	way3.Push(geo.NewPoint(9.806125648319721, 53.55134983689904))
	way3.Push(geo.NewPoint(9.806103184819221, 53.55139963393354))

	var clockwiseWay = geo.NewPointSet()
	clockwiseWay.Push(geo.NewPoint(9.869660139083862, 53.545229135979696))
	clockwiseWay.Push(geo.NewPoint(9.806103184819221, 53.55139963393354))

	var toBeIgnoredWay = geo.NewPointSet()
	toBeIgnoredWay.Push(geo.NewPoint(9.123456111111111, 53.891011111111111))
	toBeIgnoredWay.Push(geo.NewPoint(9.234567111111111, 53.910111111111111))
	toBeIgnoredWay.Push(geo.NewPoint(9.345678111111111, 53.101112111111111))

	factoredPolygons := getPolygonsFromWays([]*geo.PointSet{way1, toBeIgnoredWay, way2, way3, clockwiseWay})
	expectedPolygon := geo.NewPointSet()
	expectedPolygon.Push(geo.NewPoint(9.869660139083862, 53.545229135979696))
	expectedPolygon.Push(geo.NewPoint(9.806103184819221, 53.55139963393354))
	expectedPolygon.Push(geo.NewPoint(9.806103184819221, 53.55139963393354))
	expectedPolygon.Push(geo.NewPoint(9.806125648319721, 53.55134983689904))
	expectedPolygon.Push(geo.NewPoint(9.80623260140419, 53.55131577569369))
	expectedPolygon.Push(geo.NewPoint(9.806404933333397, 53.55129247064265))
	expectedPolygon.Push(geo.NewPoint(9.806404933333397, 53.55129247064265))
	expectedPolygon.Push(geo.NewPoint(9.80645354837179, 53.55131298705578))
	expectedPolygon.Push(geo.NewPoint(9.869220256805418, 53.54410713060241))
	expectedPolygon.Push(geo.NewPoint(9.869220256805418, 53.54410713060241))
	expectedPolygon.Push(geo.NewPoint(9.869660139083862, 53.545229135979696))

	assert.Len(t, factoredPolygons, 1)
	assert.True(t, expectedPolygon.Equals(factoredPolygons[0]))
}

func TestGetMultiplePolygonFromWays(t *testing.T) {
	var way1 = geo.NewPointSet()
	way1.Push(geo.NewPoint(9.869660139083862, 53.545229135979696))
	way1.Push(geo.NewPoint(9.869220256805418, 53.54410713060241))

	var way2 = geo.NewPointSet()
	way2.Push(geo.NewPoint(9.869220256805418, 53.54410713060241))
	way2.Push(geo.NewPoint(9.80645354837179, 53.55131298705578))
	way2.Push(geo.NewPoint(9.806404933333397, 53.55129247064265))

	var way3 = geo.NewPointSet()
	way3.Push(geo.NewPoint(9.806404933333397, 53.55129247064265))
	way3.Push(geo.NewPoint(9.80623260140419, 53.55131577569369))
	way3.Push(geo.NewPoint(9.806125648319721, 53.55134983689904))
	way3.Push(geo.NewPoint(9.806103184819221, 53.55139963393354))

	var way4 = geo.NewPointSet()
	way4.Push(geo.NewPoint(9.806103184819221, 53.55139963393354))
	way4.Push(geo.NewPoint(9.869660139083862, 53.545229135979696))

	var way5 = geo.NewPointSet()
	way5.Push(geo.NewPoint(9.806362353265285, 53.55137991431488))
	way5.Push(geo.NewPoint(9.80620376765728, 53.55143369507137))

	var way6 = geo.NewPointSet()
	way6.Push(geo.NewPoint(9.80620376765728, 53.55143369507137))
	way6.Push(geo.NewPoint(9.806171916425228, 53.55138529239359))

	var way7 = geo.NewPointSet()
	way7.Push(geo.NewPoint(9.806171916425228, 53.55138529239359))
	way7.Push(geo.NewPoint(9.806314073503017, 53.551339479108485))

	var way8 = geo.NewPointSet()
	way8.Push(geo.NewPoint(9.806314073503017, 53.551339479108485))
	way8.Push(geo.NewPoint(9.806362353265285, 53.55137991431488))

	var toBeIgnoredWay = geo.NewPointSet()
	toBeIgnoredWay.Push(geo.NewPoint(9.123456111111111, 53.891011111111111))
	toBeIgnoredWay.Push(geo.NewPoint(9.234567111111111, 53.910111111111111))
	toBeIgnoredWay.Push(geo.NewPoint(9.345678111111111, 53.101112111111111))

	factoredPolygons := getPolygonsFromWays([]*geo.PointSet{way1, toBeIgnoredWay, way2, way3, way4, way5, way6, way7, way8})

	expectedPolygon1 := geo.NewPointSet()
	expectedPolygon1.Push(geo.NewPoint(9.869660139083862, 53.545229135979696))
	expectedPolygon1.Push(geo.NewPoint(9.869220256805418, 53.54410713060241))
	expectedPolygon1.Push(geo.NewPoint(9.869220256805418, 53.54410713060241))
	expectedPolygon1.Push(geo.NewPoint(9.80645354837179, 53.55131298705578))
	expectedPolygon1.Push(geo.NewPoint(9.806404933333397, 53.55129247064265))
	expectedPolygon1.Push(geo.NewPoint(9.806404933333397, 53.55129247064265))
	expectedPolygon1.Push(geo.NewPoint(9.80623260140419, 53.55131577569369))
	expectedPolygon1.Push(geo.NewPoint(9.806125648319721, 53.55134983689904))
	expectedPolygon1.Push(geo.NewPoint(9.806103184819221, 53.55139963393354))
	expectedPolygon1.Push(geo.NewPoint(9.806103184819221, 53.55139963393354))
	expectedPolygon1.Push(geo.NewPoint(9.869660139083862, 53.545229135979696))

	expectedPolygon2 := geo.NewPointSet()
	expectedPolygon2.Push(geo.NewPoint(9.806362353265285, 53.55137991431488))
	expectedPolygon2.Push(geo.NewPoint(9.80620376765728, 53.55143369507137))
	expectedPolygon2.Push(geo.NewPoint(9.80620376765728, 53.55143369507137))
	expectedPolygon2.Push(geo.NewPoint(9.806171916425228, 53.55138529239359))
	expectedPolygon2.Push(geo.NewPoint(9.806171916425228, 53.55138529239359))
	expectedPolygon2.Push(geo.NewPoint(9.806314073503017, 53.551339479108485))
	expectedPolygon2.Push(geo.NewPoint(9.806314073503017, 53.551339479108485))
	expectedPolygon2.Push(geo.NewPoint(9.806362353265285, 53.55137991431488))

	assert.Len(t, factoredPolygons, 2)
	assert.True(t, expectedPolygon1.Equals(factoredPolygons[0]))
	assert.True(t, expectedPolygon2.Equals(factoredPolygons[1]))
}
