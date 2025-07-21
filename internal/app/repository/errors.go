package repository

import "errors"

// ErrURLNotFound means URL not found.
var ErrURLNotFound = errors.New("url not found")

// ErrOpenDB means error opening DB.
var ErrOpenDB = errors.New("error while opening db")

// ErrWarmDB means error warming DB.
var ErrWarmDB = errors.New("error while warming db up")

// ErrPingDB means error pinging DB.
var ErrPingDB = errors.New("error while ping to db")

// ErrSaveURL means error saving URL.
var ErrSaveURL = errors.New("error while saving url")

// ErrGetURL means error getting URL.
var ErrGetURL = errors.New("error while getting url")

// ErrGetURLList means error getting URL list.
var ErrGetURLList = errors.New("error while getting url list")

// ErrBeginTransaction means error beginning transaction.
var ErrBeginTransaction = errors.New("error while beginning transaction")
