package customError

import "errors"

var (
	ExistingLoginError     = errors.New("user with this login is already exists")
	IncorrectPasswordError = errors.New("incorrect password")
	UnexistingLoginError   = errors.New("user with this login does not exist")
	InvalidTokenError      = errors.New("token is invalid")
	ExpiredTokenError      = errors.New("token is expired")
)
