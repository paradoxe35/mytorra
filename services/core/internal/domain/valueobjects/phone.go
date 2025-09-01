package valueobjects

import (
	"errors"
	"regexp"
	"strings"
)

var phoneRegex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

type Phone struct {
	value string
}

func NewPhone(value string) (Phone, error) {
	value = strings.TrimSpace(value)
	if !phoneRegex.MatchString(value) {
		return Phone{}, errors.New("invalid phone format (must be E.164 format)")
	}
	return Phone{value: value}, nil
}

func (p Phone) String() string {
	return p.value
}

func (p Phone) Equal(other Phone) bool {
	return p.value == other.value
}

func (p Phone) IsEmpty() bool {
	return p.value == ""
}
