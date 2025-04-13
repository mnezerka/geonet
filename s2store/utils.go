package s2store

import (
	"math"
)

func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371e3 // Earth radius in meters
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c // distance in meters
}

// create edge with sorted point ids to avoid duplicates (reverse direction of track movement)
func edgeIdFromPointIds(from, to int64) S2EdgeKey {
	edgePoints := []int64{min(from, to), max(from, to)}
	return S2EdgeKey{edgePoints[0], edgePoints[1]}
}
