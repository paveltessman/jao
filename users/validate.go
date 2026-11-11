package users

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var ErrBadString = errors.New("bad string")

func validatePassword(passwd string) error {
	return validateLen(8, 32, passwd, "password")
}

func validateName(name string) error {

	if strings.ContainsFunc(name, unicode.IsSpace) {
		return fmt.Errorf("%w: whitespaces are not allowed in name", ErrBadString)
	}

	if !onlyASCII(name) {
		return fmt.Errorf("%w: non-ascii chars are not allowed in username", ErrBadString)
	}

	return validateLen(1, 32, name, "username")
}

func validateLen(min, max int, value, valueName string) error {

	if len(value) < min {
		return fmt.Errorf("%w: %s is too short, min %d chars required", ErrBadString, valueName, min)
	}

	if len(value) > max {
		return fmt.Errorf("%w: %s is too long, max %d chars allowed", ErrBadString, valueName, max)
	}

	return nil
}

func onlyASCII(s string) bool {
	firstNonASCII := strings.IndexFunc(s, func(r rune) bool {
		return r > unicode.MaxASCII
	})
	return firstNonASCII == -1
}
