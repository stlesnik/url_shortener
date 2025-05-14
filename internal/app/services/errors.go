package services

import "errors"

var (
	ErrSave        = errors.New("save error")
	ErrURLNotFound = errors.New("url not found")
)
