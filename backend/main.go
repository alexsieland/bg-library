package main

import (
	"log"
	"net/http"

	"github.com/alexsieland/bg-library/api"
	"github.com/gin-gonic/gin"
)

func main() {
	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code

	server := api.NewServer()
	err := server.Database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer server.Database.Close()

	r := gin.Default()

	api.RegisterSwagger(r)
	api.RegisterHandlers(r, server)

	// And we serve HTTP until the world ends.

	s := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:8080",
	}

	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}
