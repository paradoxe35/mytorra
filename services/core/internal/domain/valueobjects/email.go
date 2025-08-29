package valueobjects

import (
	"errors"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type Email struct {
	value string
}

func NewEmail(value string) (Email, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if !emailRegex.MatchString(value) {
		return Email{}, errors.New("invalid email format")
	}
	return Email{value: value}, nil
}

func (e Email) String() string {
	return e.value
}

func (e Email) Equal(other Email) bool {
	return e.value == other.value
}

func (e Email) IsEmpty() bool {
	return e.value == ""
}
