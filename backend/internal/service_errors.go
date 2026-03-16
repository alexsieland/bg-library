package internal

import "fmt"

var ErrNotFound = fmt.Errorf("not found")
var ErrAlreadyExists = fmt.Errorf("already exists")
var ErrInvalidInput = fmt.Errorf("invalid input")

func wrapDatabaseError[T any](v T, err error) (T, error) {
	if err == nil {
		return v, err
	}
	if isForeignKeyConstraintViolation(err) {
		return v, ErrInvalidInput
	}
	if isUniqueConstraintViolation(err) {
		return v, ErrAlreadyExists
	}
	if isNotFound(err) {
		return v, ErrNotFound
	}
	return v, err
}

type ValidationError struct {
	ErrorDetails []ValidationErrorDetail
}

type ValidationErrorDetail struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %v", e.ErrorDetails)
}

func (e *ValidationError) Empty() bool {
	return len(e.ErrorDetails) == 0
}

func (e *ValidationError) AddErrorDetail(field string, message string) {
	e.ErrorDetails = append(e.ErrorDetails, ValidationErrorDetail{
		Field:   field,
		Message: message,
	})
}

func (e *ValidationError) ValidateStringLength(fieldName string, str string, minLength int, maxLength int) {
	if len(str) < minLength || len(str) > maxLength {
		e.AddErrorDetail(fieldName, fmt.Sprintf("Length must be between %d and %d", minLength, maxLength))
	}
}

func (e *ValidationError) ValidateIntMin(fieldName string, i int32, minVal int32) {
	if i < minVal {
		e.AddErrorDetail(fieldName, fmt.Sprintf("Must be greater than or equal to %d", minVal))
	}
}

func (e *ValidationError) ValidateIntMax(fieldName string, i int32, maxVal int32) {
	if i > maxVal {
		e.AddErrorDetail(fieldName, fmt.Sprintf("Must be less than or equal to %d", maxVal))
	}
}
