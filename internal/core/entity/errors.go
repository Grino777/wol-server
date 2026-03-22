package entity

import "errors"

// ===== DB errors =====
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

// ===== API Errors =====
var (
	ErrInternalServerError = errors.New("internal error")
	ErrInvalidData         = errors.New("invalid data")
	ErrBadRequest          = errors.New("bad request")
)

// ===== Action Errors =====
var (
	ErrMagicPacketNotSent = errors.New("magic packet was not sent")
)
