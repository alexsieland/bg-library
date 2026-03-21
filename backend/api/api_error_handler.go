package api

import (
	"errors"
	"log"
	"net/http"

	"github.com/alexsieland/bg-library/internal"
	"github.com/gin-gonic/gin"
)

func extractRequestBody[T any](c *gin.Context, request *T) {
	// Pass the request pointer directly so the binder can populate the value.
	err := c.ShouldBindBodyWithJSON(request)
	if err != nil {
		malformedJson(c)
	}
}

func internalError(c *gin.Context, err error) {
	log.Printf("Internal error: %v", err)
	safeAbortWithStatusJSON(c, http.StatusInternalServerError, NewInternalError(err))
}

func notFound(c *gin.Context) {
	safeAbortWithStatusJSON(c, http.StatusNotFound, NewErrorResponse(NOTFOUND, "Resource not found"))
}

func badRequest(c *gin.Context, message string) {
	safeAbortWithStatusJSON(c, http.StatusBadRequest, NewErrorResponse(NOTFOUND, message))
}

func malformedJson(c *gin.Context) {
	safeAbortWithStatusJSON(c, http.StatusBadRequest, NewErrorResponse(MALFORMEDREQUEST, "JSON body is malformed"))
}

func validationError(c *gin.Context, errorDetails ErrorDetails) {
	safeAbortWithStatusJSON(c, http.StatusBadRequest, NewErrorResponseWithDetails(VALIDATIONERROR, "Validation error", errorDetails.Details))
}

func conflict(c *gin.Context, message string) {
	safeAbortWithStatusJSON(c, http.StatusConflict, NewErrorResponse(CONFLICT, message))
}

// safeAbortWithStatusJSON attempts to write a JSON error response but avoids
// writing if the response has already been committed. It also recovers from
// any panic during the write so that middleware/tests can log and observe the
// original error instead of a panic caused while encoding the error body.
func safeAbortWithStatusJSON(c *gin.Context, code int, body any) {
	// Log the intent so server logs reflect what we attempted to return.
	log.Printf("Writing error response code=%d body=%T written=%v", code, body, c.Writer.Written())
	if c.Writer.Written() {
		// If the response was already written, do not try to write another body.
		log.Printf("safeAbortWithStatusJSON: response already written, skipping error write")
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered while writing error response: %v", r)
		}
	}()
	c.AbortWithStatusJSON(code, body)
}

func handleError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if errors.Is(err, internal.ErrNotFound) {
		notFound(c)
		return
	}
	if errors.Is(err, internal.ErrAlreadyExists) {
		conflict(c, "Resource already exists")
		return
	}
	if errors.Is(err, internal.ErrCheckOutConflict) {
		conflict(c, "Game already checked out by another patron")
		return
	}
	if errors.Is(err, internal.ErrClaimUnwonPtwGame) {
		badRequest(c, "Cannot claim play-to-win game without a winner")
		return
	}
	if errors.Is(err, internal.ErrInvalidInput) {
		badRequest(c, "Invalid input")
		return
	}
	var errorDetails ErrorDetails
	if errors.As(err, &errorDetails) {
		validationError(c, errorDetails)
		return
	}
	internalError(c, err)
}
