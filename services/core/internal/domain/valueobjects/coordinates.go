package valueobjects

import (
	"errors"
	"math"
)

type Coordinates struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

func NewCoordinates(lat, lng float64) (Coordinates, error) {
	if lat < -90 || lat > 90 {
		return Coordinates{}, errors.New("latitude must be between -90 and 90")
	}
	if lng < -180 || lng > 180 {
		return Coordinates{}, errors.New("longitude must be between -180 and 180")
	}
	return Coordinates{
		Latitude:  lat,
		Longitude: lng,
	}, nil
}

// DistanceTo calculates the distance in kilometers between two coordinates using Haversine formula
func (c Coordinates) DistanceTo(other Coordinates) float64 {
	const earthRadiusKm = 6371.0

	latRad1 := c.Latitude * math.Pi / 180
	latRad2 := other.Latitude * math.Pi / 180
	deltaLat := (other.Latitude - c.Latitude) * math.Pi / 180
	deltaLng := (other.Longitude - c.Longitude) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(latRad1)*math.Cos(latRad2)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c2 := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c2
}

func (c Coordinates) IsZero() bool {
	return c.Latitude == 0 && c.Longitude == 0
}
