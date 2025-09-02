package database

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (p Point) Value() (driver.Value, error) {
	if p.Lat == 0 && p.Lng == 0 {
		return nil, nil
	}
	return fmt.Sprintf("POINT(%f %f)", p.Lng, p.Lat), nil
}

func (p *Point) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var str string
	switch v := value.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	default:
		return fmt.Errorf("cannot scan %T into Point", value)
	}

	str = strings.TrimPrefix(str, "POINT(")
	str = strings.TrimSuffix(str, ")")
	parts := strings.Split(str, " ")

	if len(parts) != 2 {
		return fmt.Errorf("invalid point format: %s", str)
	}

	lng, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return err
	}

	lat, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return err
	}

	p.Lat = lat
	p.Lng = lng
	return nil
}

func (p Point) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]float64{
		"lat": p.Lat,
		"lng": p.Lng,
	})
}

func (p *Point) UnmarshalJSON(data []byte) error {
	var m map[string]float64
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	p.Lat = m["lat"]
	p.Lng = m["lng"]
	return nil
}

func (p Point) String() string {
	return fmt.Sprintf("POINT(%f %f)", p.Lng, p.Lat)
}

func (p Point) GormDataType() string {
	return "geometry(Point, 4326)"
}

func (p Point) BunDataType() string {
	return "geometry(Point, 4326)"
}

type Polygon struct {
	Points []Point `json:"points"`
}

func (p Polygon) Value() (driver.Value, error) {
	if len(p.Points) == 0 {
		return nil, nil
	}

	var coordinates []string
	for _, point := range p.Points {
		coordinates = append(coordinates, fmt.Sprintf("%f %f", point.Lng, point.Lat))
	}

	if len(coordinates) > 0 && coordinates[0] != coordinates[len(coordinates)-1] {
		coordinates = append(coordinates, coordinates[0])
	}

	return fmt.Sprintf("POLYGON((%s))", strings.Join(coordinates, ", ")), nil
}

func (p *Polygon) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var str string
	switch v := value.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	default:
		return fmt.Errorf("cannot scan %T into Polygon", value)
	}

	str = strings.TrimPrefix(str, "POLYGON((")
	str = strings.TrimSuffix(str, "))")

	coordinates := strings.Split(str, ", ")
	p.Points = make([]Point, 0, len(coordinates))

	for _, coord := range coordinates {
		parts := strings.Split(coord, " ")
		if len(parts) != 2 {
			continue
		}

		lng, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return err
		}

		lat, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return err
		}

		p.Points = append(p.Points, Point{Lat: lat, Lng: lng})
	}

	if len(p.Points) > 0 && p.Points[0] == p.Points[len(p.Points)-1] {
		p.Points = p.Points[:len(p.Points)-1]
	}

	return nil
}

func (p Polygon) GormDataType() string {
	return "geometry(Polygon, 4326)"
}

func (p Polygon) BunDataType() string {
	return "geometry(Polygon, 4326)"
}

type LineString struct {
	Points []Point `json:"points"`
}

func (l LineString) Value() (driver.Value, error) {
	if len(l.Points) == 0 {
		return nil, nil
	}

	var coordinates []string
	for _, point := range l.Points {
		coordinates = append(coordinates, fmt.Sprintf("%f %f", point.Lng, point.Lat))
	}

	return fmt.Sprintf("LINESTRING(%s)", strings.Join(coordinates, ", ")), nil
}

func (l *LineString) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var str string
	switch v := value.(type) {
	case []byte:
		str = string(v)
	case string:
		str = v
	default:
		return fmt.Errorf("cannot scan %T into LineString", value)
	}

	str = strings.TrimPrefix(str, "LINESTRING(")
	str = strings.TrimSuffix(str, ")")

	coordinates := strings.Split(str, ", ")
	l.Points = make([]Point, 0, len(coordinates))

	for _, coord := range coordinates {
		parts := strings.Split(coord, " ")
		if len(parts) != 2 {
			continue
		}

		lng, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return err
		}

		lat, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return err
		}

		l.Points = append(l.Points, Point{Lat: lat, Lng: lng})
	}

	return nil
}

func (l LineString) GormDataType() string {
	return "geometry(LineString, 4326)"
}

func (l LineString) BunDataType() string {
	return "geometry(LineString, 4326)"
}

type Geography struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}

func (g Geography) Value() (driver.Value, error) {
	if g.Type == "" {
		return nil, nil
	}

	data, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}

	return fmt.Sprintf("ST_GeogFromGeoJSON('%s')", string(data)), nil
}

func (g *Geography) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into Geography", value)
	}

	return json.Unmarshal(data, g)
}

func (g Geography) GormDataType() string {
	return "geography"
}

func (g Geography) BunDataType() string {
	return "geography"
}

func PointFromLatLng(lat, lng float64) Point {
	return Point{Lat: lat, Lng: lng}
}

func PolygonFromPoints(points ...Point) Polygon {
	return Polygon{Points: points}
}

func LineStringFromPoints(points ...Point) LineString {
	return LineString{Points: points}
}

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

type PostGISQuery struct {
	db interface{}
}

func NewPostGISQuery(db interface{}) *PostGISQuery {
	return &PostGISQuery{db: db}
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
