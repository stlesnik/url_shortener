package repository

import "errors"

var (
	ErrURLNotFound      = errors.New("url not found")
	ErrOpenDB           = errors.New("error while opening db")
	ErrWarmDB           = errors.New("error while warming db up")
	ErrPingDB           = errors.New("error while ping to db")
	ErrSaveURL          = errors.New("error while saving url")
	ErrGetURL           = errors.New("error while getting url")
	ErrGetURLList       = errors.New("error while getting url list")
	ErrBeginTransaction = errors.New("error while beginning transaction")
)
