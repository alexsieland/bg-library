package api

import (
	"fmt"

	"github.com/alexsieland/bg-library/db"
)

type ErrorDetails struct {
	Details []ErrorDetail
}

func (e ErrorDetails) Error() string {
	return fmt.Sprintf("Validation Error Details: %v", e.Details)
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

func (e *ErrorDetails) ValidateEnum(fieldName string, value string, allowedValues []string) {
	for _, allowedValue := range allowedValues {
		if value == allowedValue {
			return
		}
	}
	e.AddErrorDetail(fieldName, fmt.Sprintf("Value must be one of: %v", allowedValues))
}

func (e *ErrorDetails) playToWinGameDeletionReason(deletionReason *string) db.NullPlayToWinGameDeletionType {
	nullableReason := db.NullPlayToWinGameDeletionType{Valid: false}
	if deletionReason != nil {
		reason := db.PlayToWinGameDeletionType(*deletionReason)
		switch reason {
		case db.PlayToWinGameDeletionTypeClaimed, db.PlayToWinGameDeletionTypeMistake, db.PlayToWinGameDeletionTypeOther:
			nullableReason = db.NullPlayToWinGameDeletionType{
				PlayToWinGameDeletionType: reason,
				Valid:                     true,
			}
		default:
			e.AddErrorDetail("deletionReason", "Invalid play to win game deletion reason")
			return db.NullPlayToWinGameDeletionType{}
		}
	}

	return nullableReason
}

func (e *ErrorDetails) playToWinEntryDeletionReason(deletionReason *string) db.NullPlayToWinEntryDeletionType {
	nullableReason := db.NullPlayToWinEntryDeletionType{Valid: false}
	if deletionReason != nil {
		reason := db.PlayToWinEntryDeletionType(*deletionReason)
		switch reason {
		case db.PlayToWinEntryDeletionTypeDuplicateEntrant,
			db.PlayToWinEntryDeletionTypeFoulPlay,
			db.PlayToWinEntryDeletionTypeWon,
			db.PlayToWinEntryDeletionTypeFailedToClaim,
			db.PlayToWinEntryDeletionTypeOther:
			nullableReason = db.NullPlayToWinEntryDeletionType{
				PlayToWinEntryDeletionType: reason,
				Valid:                      true,
			}
		default:
			e.AddErrorDetail("deletionReason", "Invalid play to win entry deletion reason")
			return db.NullPlayToWinEntryDeletionType{}
		}
	}

	return nullableReason
}
