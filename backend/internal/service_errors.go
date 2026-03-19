package internal

import "fmt"

var ErrNotFound = fmt.Errorf("not found")
var ErrAlreadyExists = fmt.Errorf("already exists")
var ErrInvalidInput = fmt.Errorf("invalid input")
var ErrCheckOutConflict = fmt.Errorf("check out conflict")
var ErrClaimUnwonPtwGame = fmt.Errorf("no play-to-win raffle winner")

// wrapDatabaseError attempts to wrap database errors with more specific errors that can be handled by the API layer.
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

// wrapErrorOrReturn wraps the error and returns the default value if there is an error.
// This is intended to reduce boilerplate code in the Services.
func wrapErrorOrReturn[T any](t *T, tDefault T, err error) (T, error) {
	if err != nil || t == nil {
		return tDefault, wrapDatabaseError(err)
	}
	return *t, nil
}
