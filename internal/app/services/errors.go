package services

import "errors"

var (
	ErrReadingBody      = errors.New("error reading body")
	ErrDidntGetURL      = errors.New("error getting url")
	ErrInvalidURL       = errors.New("invalid url to shorten")
	ErrNoUserID         = errors.New("no user id in request context")
	ErrConvertingUserID = errors.New("cannot convert userID to string")
)
