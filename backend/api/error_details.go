package api

import "fmt"

type ErrorDetails struct {
	Details []ErrorDetail
}

func (e *ErrorDetails) Empty() bool {
	return len(e.Details) == 0
}

func (e *ErrorDetails) AddErrorDetail(field string, message string) {
	e.Details = append(e.Details, ErrorDetail{
		Field:   field,
		Message: message,
	})
}

func (e *ErrorDetails) ValidateStringLength(fieldName string, str string, minLength int, maxLength int) {
	if len(str) < minLength || len(str) > maxLength {
		e.AddErrorDetail(fieldName, fmt.Sprintf("Length must be between %d and %d", minLength, maxLength))
	}
}

func (e *ErrorDetails) ValidateIntMin(fieldName string, i int32, minVal int32) {
	if i < minVal {
		e.AddErrorDetail(fieldName, fmt.Sprintf("Must be greater than or equal to %d", minVal))
	}
}

func (e *ErrorDetails) ValidateIntMax(fieldName string, i int32, maxVal int32) {
	if i > maxVal {
		e.AddErrorDetail(fieldName, fmt.Sprintf("Must be less than or equal to %d", maxVal))
	}
}
