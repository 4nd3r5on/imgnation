package common

import (
	"errors"
	"regexp"
)

func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func ValidatePassword(s string) error {
	if len(s) < 8 {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}
