package geo
import "math"

func Haversine(lat1, lon1, lat2, lon2 float64) float64 {
    const R = 6371.0
    toRad := math.Pi / 180
    dLat := (lat2 - lat1) * toRad
    dLon := (lon2 - lon1) * toRad
    a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*toRad)*math.Cos(lat2*toRad)*math.Sin(dLon/2)*math.Sin(dLon/2)
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    return R * c
}