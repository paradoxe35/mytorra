package database

import (
	"fmt"

	"github.com/restayway/gogis"
)

type Point = gogis.Point

type Polygon = gogis.Polygon

type LineString = gogis.LineString

type Geometry = gogis.Geometry

type GeometryCollection = gogis.GeometryCollection

func Distance(p1, p2 Point) float64 {
	const earthRadius = 6371000

	lat1Rad := p1.Lat * (3.14159265359 / 180)
	lat2Rad := p2.Lat * (3.14159265359 / 180)
	deltaLat := (p2.Lat - p1.Lat) * (3.14159265359 / 180)
	deltaLng := (p2.Lng - p1.Lng) * (3.14159265359 / 180)

	a := (deltaLat/2)*(deltaLat/2) +
		lat1Rad*lat2Rad*
			(deltaLng/2)*(deltaLng/2)
	c := 2 * a

	return earthRadius * c
}

func BoundingBox(points []Point) (min, max Point) {
	if len(points) == 0 {
		return
	}

	min = points[0]
	max = points[0]

	for _, p := range points[1:] {
		if p.Lat < min.Lat {
			min.Lat = p.Lat
		}
		if p.Lng < min.Lng {
			min.Lng = p.Lng
		}
		if p.Lat > max.Lat {
			max.Lat = p.Lat
		}
		if p.Lng > max.Lng {
			max.Lng = p.Lng
		}
	}

	return min, max
}

type PostGISQuery struct{}

func NewPostGISQuery() *PostGISQuery {
	return &PostGISQuery{}
}

func (q *PostGISQuery) PointsWithinRadius(center Point, radiusMeters float64) string {
	return fmt.Sprintf(
		"ST_DWithin(location::geography, ST_MakePoint(%f, %f)::geography, %f)",
		center.Lng, center.Lat, radiusMeters,
	)
}

func (q *PostGISQuery) PointsWithinPolygon(polygon Polygon) string {
	polygonWKT, _ := polygon.Value()
	return fmt.Sprintf(
		"ST_Within(location, ST_GeomFromText('%s', 4326))",
		polygonWKT,
	)
}

func (q *PostGISQuery) DistanceQuery(from Point) string {
	return fmt.Sprintf(
		"ST_Distance(location::geography, ST_MakePoint(%f, %f)::geography)",
		from.Lng, from.Lat,
	)
}

func (q *PostGISQuery) NearestNeighbors(from Point, limit int) string {
	return fmt.Sprintf(
		"location <-> ST_MakePoint(%f, %f)::geometry",
		from.Lng, from.Lat,
	)
}
