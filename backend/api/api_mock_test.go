package api

import (
	"github.com/alexsieland/bg-library/internal"
	"github.com/gin-gonic/gin"
)

func setupTestServer() (*Server, *internal.LibraryService, *PatronApi) {
	gin.SetMode(gin.TestMode)
	var mockLibService = new(internal.LibraryService)
	var mockPatronApi = new(PatronApi)
	var server = Server{
		LibService: mockLibService,
		PatronApi:  mockPatronApi,
	}
	return &server, mockLibService, mockPatronApi
}
