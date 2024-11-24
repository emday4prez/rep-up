package data

import "errors"

var (
	ErrRecordNotFound       = errors.New("data: record not found")
	ErrInvalidInput         = errors.New("data: invalid input")
	ErrReferentialIntegrity = errors.New("data: cannot delete record due to referential integrity constraint")
)
