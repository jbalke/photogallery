package models

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	// ErrNotFound is returned when a resource can not be found in the DB.
	ErrNotFound modelError = "models: resource not found"

	// ErrIDInvalid is returned when an invalid ID is provided to a method like Delete.
	ErrIDInvalid modelError = "models: ID provided is invalid"

	// ErrPasswordIncorrect is returned when the credentials provided to Authenticate() are incorrect.
	ErrPasswordIncorrect modelError = "models: incorrect password"

	// ErrEmailRequired is returned when an email address is not provided for user creation\update.
	ErrEmailRequired modelError = "models: email address is required"

	// ErrEmailInvalid is returned when a provided email does not match our expected pattern.
	ErrEmailInvalid modelError = "models: email address is not valid"

	// ErrEmailTaken is returned when a user attempts to register an email address that is taken by another user.
	ErrEmailTaken modelError = "models: email address is already taken"

	// ErrPasswordRequired is returned if the user does not provide a password when signing up.
	ErrPasswordRequired modelError = "models: password is required"

	// ErrRememberTooShort is returned if a user's remember token is less than 32 bytes.
	ErrRememberTooShort modelError = "models: remember token must be at least 32 bytes"

	// ErrRememberHashRequired is returned if a remember hash is not present on user create and update.
	ErrRememberHashRequired modelError = "models: remember hash is required"

	// ErrNameRequired is returned if a user does not provide a name on user catete and update.
	ErrNameRequired modelError = "models: name is required"
)

// ErrPasswordNotComplex is returned if a provided password does not meet complexity requirements.
// Passwords must be between 6 and 13 characters long and include lowercase and uppercase characters, as well as a number and symbol.
var ErrPasswordNotComplex = modelError(fmt.Sprint("models: password must be between %d and %d characters long and include a lowercase and uppercase character, a number and a symbol", minPasswordLength, maxPasswordLength))

type modelError string

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	s := strings.Replace(string(e), "models: ", "", 1)
	// use a rune slice to manipulate individial runes/characters
	a := []rune(s)
	a[0] = unicode.ToUpper(a[0])
	return string(a)
}