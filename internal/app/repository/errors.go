package repository

import "errors"

// ErrURLNotFound is returned when a URL is not found in the repository.
var ErrURLNotFound = errors.New("url not found")

// ErrOpenDB is returned when there is an error opening the database.
var ErrOpenDB = errors.New("error while opening db")

// ErrWarmDB is returned when there is an error warming up the database.
var ErrWarmDB = errors.New("error while warming db up")

// ErrPingDB is returned when there is an error pinging the database.
var ErrPingDB = errors.New("error while ping to db")

// ErrSaveURL is returned when there is an error saving a URL.
var ErrSaveURL = errors.New("error while saving url")

// ErrGetURL is returned when there is an error getting a URL.
var ErrGetURL = errors.New("error while getting url")

// ErrGetURLList is returned when there is an error getting a URL list.
var ErrGetURLList = errors.New("error while getting url list")

// ErrBeginTransaction is returned when there is an error beginning a transaction.
var ErrBeginTransaction = errors.New("error while beginning transaction")
