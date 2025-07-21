package services

import "errors"

// ErrReadingBody is returned when there is an error reading the request body.
var ErrReadingBody = errors.New("error reading body")

// ErrDidntGetURL is returned when the URL is missing in the request.
var ErrDidntGetURL = errors.New("error getting url")

// ErrInvalidURL is returned when the URL to shorten is invalid.
var ErrInvalidURL = errors.New("invalid url to shorten")

// ErrNoUserID is returned when there is no user ID in the request context.
var ErrNoUserID = errors.New("no user id in request context")

// ErrConvertingUserID is returned when the user ID cannot be converted to string.
var ErrConvertingUserID = errors.New("cannot convert userID to string")
