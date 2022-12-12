package main

import (
	"fmt"
	"net/http"
	"log"
	"github.com/checkking/oauth2_practice/third-service/handlers"
)

func main() {
	// We create a simple server using http.Server and run.
	server := &http.Server{
		Addr: fmt.Sprintf(":8000"),
		Handler: handlers.New(),
	}

	log.Printf("Starting HTTP Server. Listening at %q", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("%v", err)
	} else {
		log.Println("Server closed!")
	}
}
