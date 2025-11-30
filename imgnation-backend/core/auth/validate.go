package auth

import (
	"regexp"
	"strings"
)

func ValidateUsername(username string) (usernameOut string, valid bool) {
	usernameOut = strings.ToLower(strings.TrimSpace(username))
	if len(usernameOut) > 64 {
		return "", false
	}
	match, _ := regexp.MatchString("^[a-z0-9][a-z0-9_]*$", usernameOut)
	return usernameOut, match
}
