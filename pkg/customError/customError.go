package customError

import "errors"

var (
	ExistingLoginError     = errors.New("user with such login already exists")
	IncorrectPasswordError = errors.New("incorrect password")
	UnexistingLoginError   = errors.New("user with such login does not exist")
	TokenNotProvidedError  = errors.New("token was not provided")
	InvalidTokenError      = errors.New("token is invalid")
	ExpiredTokenError      = errors.New("token is expired")
	NoPermission           = errors.New("no permission for such request")
	TypeNotAllowed         = errors.New("such file type is not allowed")
	ExistingFileError      = errors.New("such file already exists")
	UnexistingFileError    = errors.New("such file does not exist")
)
