package services

import "errors"

var (
	ErrSave        = errors.New("save error")
	ErrUrlNotFound = errors.New("url not found")
)
