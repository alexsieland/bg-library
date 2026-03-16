package api

import (
	"errors"
	"net/http"

	"github.com/alexsieland/bg-library/internal"
	"github.com/gin-gonic/gin"
)

func extractRequestBody[T any](c *gin.Context, request T) {
	err := c.ShouldBindBodyWithJSON(&request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(MALFORMEDREQUEST, "JSON body is malformed"))
	}
}

func internalError(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, NewInternalError(err))
}

func notFound(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusNotFound, NewErrorResponse(NOTFOUND, "Resource not found"))
}

func badRequest(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(NOTFOUND, message))
}

func malformedJson(c *gin.Context) {
}

func validationError(c *gin.Context, errorDetails ErrorDetails) {
	c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponseWithDetails(VALIDATIONERROR, "Validation error", errorDetails.Details))
}

func conflict(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusConflict, NewErrorResponse(CONFLICT, message))
}

func handleError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if errors.Is(err, (internal.ErrNotFound)) {
		notFound(c)
		return
	}
	if errors.Is(err, (internal.ErrAlreadyExists)) {
		conflict(c, "Resource already exists")
		return
	}
	if errors.Is(err, (internal.ErrInvalidInput)) {
		badRequest(c, "Invalid input")
	}
	var errorDetails ErrorDetails
	if errors.As(err, &errorDetails) {
		validationError(c, errorDetails)
		return
	}
	internalError(c, err)
}
