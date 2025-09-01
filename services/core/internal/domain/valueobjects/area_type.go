package valueobjects

type AreaType string

const (
	AreaTypeCountry  AreaType = "country"
	AreaTypeProvince AreaType = "province"
	AreaTypeTown     AreaType = "town"
	AreaTypeSubarea  AreaType = "subarea"
)

func (t AreaType) String() string {
	return string(t)
}

func (t AreaType) IsValid() bool {
	switch t {
	case AreaTypeCountry, AreaTypeProvince, AreaTypeTown, AreaTypeSubarea:
		return true
	default:
		return false
	}
}

func (t AreaType) Level() int {
	switch t {
	case AreaTypeCountry:
		return 1
	case AreaTypeProvince:
		return 2
	case AreaTypeTown:
		return 3
	case AreaTypeSubarea:
		return 4
	default:
		return 0
	}
}
