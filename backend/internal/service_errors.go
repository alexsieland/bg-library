package internal

import "fmt"

var ErrNotFound = fmt.Errorf("not found")
var ErrAlreadyExists = fmt.Errorf("already exists")
var ErrInvalidInput = fmt.Errorf("invalid input")
var ErrCheckOutConflict = fmt.Errorf("check out conflict")

// Attempt to wrap database errors with more specific errors that can be handled by the API layer.
// For example, if a foreign key constraint is violated, we can return an ErrInvalidInput instead of the raw database error.
func wrapDatabaseError(err error) error {
	if err == nil {
		return err
	}
	if isForeignKeyConstraintViolation(err) {
		return ErrInvalidInput
	}
	if isUniqueConstraintViolation(err) {
		return ErrAlreadyExists
	}
	if isNotFound(err) {
		return ErrNotFound
	}
	return err
}
