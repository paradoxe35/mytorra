package valueobjects

type AreaStatus string

const (
	AreaStatusActive    AreaStatus = "active"
	AreaStatusInactive  AreaStatus = "inactive"
	AreaStatusTesting   AreaStatus = "testing"
	AreaStatusSuspended AreaStatus = "suspended"
)

func (s AreaStatus) String() string {
	return string(s)
}

func (s AreaStatus) IsValid() bool {
	switch s {
	case AreaStatusActive, AreaStatusInactive, AreaStatusTesting, AreaStatusSuspended:
		return true
	default:
		return false
	}
}

func (s AreaStatus) IsOperational() bool {
	return s == AreaStatusActive || s == AreaStatusTesting
}
