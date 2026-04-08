package zipper

import "math"

const earthRadiusKM = 6371.0

// haversine returns the great-circle distance in kilometers between two points.
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := degToRad(lat2 - lat1)
	dLon := degToRad(lon2 - lon1)

	lat1R := degToRad(lat1)
	lat2R := degToRad(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1R)*math.Cos(lat2R)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKM * c
}

func degToRad(deg float64) float64 {
	return deg * math.Pi / 180
}
